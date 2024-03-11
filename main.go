package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ariga.io/sqlcomment"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"go.artefactual.dev/tools/log"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_client "go.temporal.io/sdk/client"
	temporalsdk_worker "go.temporal.io/sdk/worker"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"
	goahttp "goa.design/goa/v3/http"

	"github.com/artefactual-sdps/enduro/internal/api"
	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goahttpstorage "github.com/artefactual-sdps/enduro/internal/api/gen/http/storage/client"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/db"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/package_"
	"github.com/artefactual-sdps/enduro/internal/storage"
	storage_activities "github.com/artefactual-sdps/enduro/internal/storage/activities"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	storage_entclient "github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/client"
	storage_entdb "github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
	storage_workflows "github.com/artefactual-sdps/enduro/internal/storage/workflows"
	"github.com/artefactual-sdps/enduro/internal/telemetry"
	"github.com/artefactual-sdps/enduro/internal/temporal"
	"github.com/artefactual-sdps/enduro/internal/upload"
	"github.com/artefactual-sdps/enduro/internal/version"
	"github.com/artefactual-sdps/enduro/internal/watcher"
	"github.com/artefactual-sdps/enduro/internal/workflow"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

const (
	appName        = "enduro"
	autoApproveAIP = true
)

// This represents the first permanent location defined in the
// mysql-create-locations-job Kubernetes manifest.
var defaultPermanentLocationID = uuid.MustParse("f2cc963f-c14d-4eaa-b950-bd207189a1f1")

func main() {
	p := pflag.NewFlagSet(appName, pflag.ExitOnError)

	p.String("config", "", "Configuration file")
	p.Bool("version", false, "Show version information")
	_ = p.Parse(os.Args[1:])

	if v, _ := p.GetBool("version"); v {
		fmt.Println(version.Info(appName))
		os.Exit(0)
	}

	var cfg config.Configuration
	configFile, _ := p.GetString("config")
	configFileFound, configFileUsed, err := config.Read(&cfg, configFile)
	if err != nil {
		fmt.Printf("Failed to read configuration: %v\n", err)
		os.Exit(1)
	}

	logger := log.New(os.Stderr,
		log.WithName(appName),
		log.WithDebug(cfg.Debug),
		log.WithLevel(cfg.Verbosity),
	)
	defer log.Sync(logger)

	logger.Info("Starting...", "version", version.Long, "pid", os.Getpid())

	if configFileFound {
		logger.Info("Configuration file loaded.", "path", configFileUsed)
	} else {
		logger.Info("Configuration file not found.")
	}

	logger.V(1).Info("Preservation config", "TaskQueue", cfg.Preservation.TaskQueue)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up the tracer provider.
	tp, shutdown, err := telemetry.TracerProvider(ctx, logger, cfg.Telemetry, appName, version.Long)
	if err != nil {
		logger.Error(err, "Error creating tracer provider.")
		os.Exit(1)
	}
	defer func() { _ = shutdown(ctx) }()

	// Set up the Enduro database client handler.
	enduroDatabase, err := db.Connect(ctx, tp, cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		logger.Error(err, "Enduro database configuration failed.")
		os.Exit(1)
	}
	if cfg.Database.Migrate {
		if err := db.MigrateEnduroDatabase(enduroDatabase); err != nil {
			logger.Error(err, "Enduro database migration failed.")
			os.Exit(1)
		}
	}

	// Set up the Storage database client handler.
	storageDatabase, err := db.Connect(ctx, tp, cfg.Storage.Database.Driver, cfg.Storage.Database.DSN)
	if err != nil {
		logger.Error(err, "Storage database configuration failed.")
		os.Exit(1)
	}
	if cfg.Storage.Database.Migrate {
		if err := db.MigrateEnduroStorageDatabase(storageDatabase); err != nil {
			logger.Error(err, "Storage database migration failed.")
			os.Exit(1)
		}
	}

	temporalClient, err := temporalsdk_client.Dial(temporalsdk_client.Options{
		Namespace: cfg.Temporal.Namespace,
		HostPort:  cfg.Temporal.Address,
		Logger:    temporal.Logger(logger.WithName("temporal-client")),
	})
	if err != nil {
		logger.Error(err, "Error creating Temporal client.")
		os.Exit(1)
	}

	// Set up the event service.
	evsvc, err := event.NewEventServiceRedis(logger.WithName("events"), &cfg.Event)
	if err != nil {
		logger.Error(err, "Error creating Event service.")
		os.Exit(1)
	}

	// Set up the OIDC token verifier.
	var tokenVerifier auth.TokenVerifier
	{
		if cfg.API.Auth.Enabled {
			tokenVerifier, err = auth.NewOIDCTokenVerifier(ctx, cfg.API.Auth.OIDC)
			if err != nil {
				logger.Error(err, "Error connecting to OIDC provider.")
				os.Exit(1)
			}
		} else {
			tokenVerifier = &auth.NoopTokenVerifier{}
		}
	}

	// Set up the WebSocket ticket provider.
	var ticketProvider *auth.TicketProvider
	{
		var store auth.TicketStore
		if cfg.API.Auth.Enabled {
			if cfg.API.Auth.Ticket.Redis != nil {
				var err error
				store, err = auth.NewRedisStore(ctx, cfg.API.Auth.Ticket.Redis)
				if err != nil {
					logger.Error(err, "Error creating ticket provider redis store.")
					os.Exit(1)
				}
			} else {
				store = auth.NewInMemStore()
			}
		}
		ticketProvider = auth.NewTicketProvider(ctx, store, rand.Reader)
		defer ticketProvider.Close()
	}

	// Set up the package service.
	var pkgsvc package_.Service
	{
		pkgsvc = package_.NewService(logger.WithName("package"), enduroDatabase, temporalClient, evsvc, tokenVerifier, ticketProvider, cfg.Temporal.TaskQueue)
	}

	// Set up the ent db client.
	var storagePersistence persistence.Storage
	{
		drv := sqlcomment.NewDriver(
			sql.OpenDB(cfg.Database.Driver, storageDatabase),
			sqlcomment.WithDriverVerTag(),
			sqlcomment.WithTags(sqlcomment.Tags{
				sqlcomment.KeyApplication: appName,
			}),
		)
		client := storage_entdb.NewClient(storage_entdb.Driver(drv))
		storagePersistence = storage_entclient.NewClient(client)
	}

	// Set up the storage service.
	var storagesvc storage.Service
	{
		storagesvc, err = storage.NewService(logger.WithName("storage"), cfg.Storage, storagePersistence, temporalClient, tokenVerifier, rand.Reader)
		if err != nil {
			logger.Error(err, "Error setting up storage service.")
			os.Exit(1)
		}
	}

	// Set up the upload service.
	var uploadsvc upload.Service
	{
		uploadsvc, err = upload.NewService(logger.WithName("upload"), cfg.Upload, upload.UPLOAD_MAX_SIZE, tokenVerifier)
		if err != nil {
			logger.Error(err, "Error setting up upload service.")
			os.Exit(1)
		}
		defer uploadsvc.Close()
	}

	// Set up the watcher service.
	var wsvc watcher.Service
	{
		wsvc, err = watcher.New(ctx, logger.WithName("watcher"), &cfg.Watcher)
		if err != nil {
			logger.Error(err, "Error setting up watchers.")
			os.Exit(1)
		}
	}

	var g run.Group

	// API server.
	{
		var srv *http.Server

		g.Add(
			func() error {
				srv = api.HTTPServer(logger, tp, &cfg.API, pkgsvc, storagesvc, uploadsvc)
				return srv.ListenAndServe()
			},
			func(err error) {
				ctx, cancel := context.WithTimeout(ctx, time.Second*5)
				defer cancel()
				_ = srv.Shutdown(ctx)
			},
		)
	}

	// Watchers, where each watcher is a group actor.
	{
		for _, w := range wsvc.Watchers() {
			done := make(chan struct{})
			cur := w
			g.Add(
				func() error {
					for {
						select {
						case <-done:
							return nil
						default:
							event, clean, err := cur.Watch(ctx)
							if err != nil {
								if !errors.Is(err, watcher.ErrWatchTimeout) {
									logger.Error(err, "Error monitoring watcher interface.", "watcher", cur)
								}
								continue
							}
							logger.V(1).Info("Starting new workflow", "watcher", event.WatcherName, "bucket", event.Bucket, "key", event.Key, "dir", event.IsDir)
							go func() {
								req := package_.ProcessingWorkflowRequest{
									WatcherName:                event.WatcherName,
									RetentionPeriod:            event.RetentionPeriod,
									CompletedDir:               event.CompletedDir,
									StripTopLevelDir:           event.StripTopLevelDir,
									Key:                        event.Key,
									IsDir:                      event.IsDir,
									AutoApproveAIP:             autoApproveAIP,
									DefaultPermanentLocationID: &defaultPermanentLocationID,
									GlobalTaskQueue:            cfg.Temporal.TaskQueue,
									PreservationTaskQueue:      cfg.Preservation.TaskQueue,
									PollInterval:               cfg.AM.PollInterval,
									TransferDeadline:           cfg.AM.TransferDeadline,
								}
								if err := package_.InitProcessingWorkflow(ctx, temporalClient, &req); err != nil {
									logger.Error(err, "Error initializing processing workflow.")
								} else {
									_ = clean(ctx)
								}
							}()
						}
					}
				},
				func(err error) {
					close(done)
				},
			)
		}
	}

	// Workflow and activity worker.
	{
		done := make(chan struct{})
		workerOpts := temporalsdk_worker.Options{}
		w := temporalsdk_worker.New(temporalClient, cfg.Temporal.TaskQueue, workerOpts)
		if err != nil {
			logger.Error(err, "Error creating Temporal worker.")
			os.Exit(1)
		}

		w.RegisterWorkflowWithOptions(workflow.NewProcessingWorkflow(logger, cfg, pkgsvc, wsvc).Execute, temporalsdk_workflow.RegisterOptions{Name: package_.ProcessingWorkflowName})
		w.RegisterActivityWithOptions(activities.NewDeleteOriginalActivity(wsvc).Execute, temporalsdk_activity.RegisterOptions{Name: activities.DeleteOriginalActivityName})
		w.RegisterActivityWithOptions(activities.NewDisposeOriginalActivity(wsvc).Execute, temporalsdk_activity.RegisterOptions{Name: activities.DisposeOriginalActivityName})

		w.RegisterWorkflowWithOptions(storage_workflows.NewStorageUploadWorkflow().Execute, temporalsdk_workflow.RegisterOptions{Name: storage.StorageUploadWorkflowName})
		w.RegisterWorkflowWithOptions(storage_workflows.NewStorageMoveWorkflow(storagesvc).Execute, temporalsdk_workflow.RegisterOptions{Name: storage.StorageMoveWorkflowName})

		w.RegisterActivityWithOptions(storage_activities.NewCopyToPermanentLocationActivity(storagesvc).Execute, temporalsdk_activity.RegisterOptions{Name: storage.CopyToPermanentLocationActivityName})

		w.RegisterWorkflowWithOptions(workflow.NewMoveWorkflow(logger, pkgsvc).Execute, temporalsdk_workflow.RegisterOptions{Name: package_.MoveWorkflowName})

		httpClient := cleanhttp.DefaultPooledClient()
		storageHttpClient := goahttpstorage.NewClient("http", cfg.Storage.EnduroAddress, httpClient, goahttp.RequestEncoder, goahttp.ResponseDecoder, false)
		storageClient := goastorage.NewClient(
			storageHttpClient.Submit(),
			storageHttpClient.Update(),
			storageHttpClient.Download(),
			storageHttpClient.Locations(),
			storageHttpClient.AddLocation(),
			storageHttpClient.Move(),
			storageHttpClient.MoveStatus(),
			storageHttpClient.Reject(),
			storageHttpClient.Show(),
			storageHttpClient.ShowLocation(),
			storageHttpClient.LocationPackages(),
		)
		w.RegisterActivityWithOptions(activities.NewMoveToPermanentStorageActivity(storageClient).Execute, temporalsdk_activity.RegisterOptions{Name: activities.MoveToPermanentStorageActivityName})
		w.RegisterActivityWithOptions(activities.NewPollMoveToPermanentStorageActivity(storageClient).Execute, temporalsdk_activity.RegisterOptions{Name: activities.PollMoveToPermanentStorageActivityName})

		g.Add(
			func() error {
				if err := w.Start(); err != nil {
					return err
				}
				<-done
				return nil
			},
			func(err error) {
				w.Stop()
				close(done)
			},
		)
	}

	// Observability server.
	{
		srv := &http.Server{
			Addr:         cfg.DebugListen,
			ReadTimeout:  time.Second * 1,
			WriteTimeout: time.Second * 1,
			IdleTimeout:  time.Second * 30,
		}

		g.Add(func() error {
			mux := http.NewServeMux()

			// Health check.
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "OK")
			})

			// Prometheus metrics.
			mux.Handle("/metrics", promhttp.Handler())

			// Profiling data.
			mux.HandleFunc("/debug/pprof/", pprof.Index)
			mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
			mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
			mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
			mux.Handle("/debug/pprof/block", pprof.Handler("block"))
			mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
			mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
			mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))

			srv.Handler = mux

			return srv.ListenAndServe()
		}, func(error) {
			ctx, cancel := context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			_ = srv.Shutdown(ctx)
		})
	}

	// Signal handler.
	{
		var (
			cancelInterrupt = make(chan struct{})
			ch              = make(chan os.Signal, 2)
		)
		defer close(ch)

		g.Add(
			func() error {
				signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

				select {
				case <-ch:
				case <-cancelInterrupt:
				}

				return nil
			}, func(err error) {
				logger.Info("Quitting...")
				close(cancelInterrupt)
				cancel()
				signal.Stop(ch)
			},
		)
	}

	err = g.Run()
	if err != nil {
		logger.Error(err, "Application failure.")
		os.Exit(1)
	}
	logger.Info("Bye!")
}
