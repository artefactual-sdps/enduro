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
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"go.artefactual.dev/tools/log"
	temporal_tools "go.artefactual.dev/tools/temporal"
	"go.opentelemetry.io/otel/codes"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_client "go.temporal.io/sdk/client"
	temporalsdk_contrib_opentelemetry "go.temporal.io/sdk/contrib/opentelemetry"
	temporalsdk_interceptor "go.temporal.io/sdk/interceptor"
	temporalsdk_worker "go.temporal.io/sdk/worker"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-sdps/enduro/internal/about"
	"github.com/artefactual-sdps/enduro/internal/api"
	"github.com/artefactual-sdps/enduro/internal/api/auth"
	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/db"
	event "github.com/artefactual-sdps/enduro/internal/event3"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	entclient "github.com/artefactual-sdps/enduro/internal/persistence/ent/client"
	entdb "github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/storage"
	storage_activities "github.com/artefactual-sdps/enduro/internal/storage/activities"
	storage_persistence "github.com/artefactual-sdps/enduro/internal/storage/persistence"
	storage_entclient "github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/client"
	storage_entdb "github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
	storage_workflows "github.com/artefactual-sdps/enduro/internal/storage/workflows"
	"github.com/artefactual-sdps/enduro/internal/telemetry"
	"github.com/artefactual-sdps/enduro/internal/version"
	"github.com/artefactual-sdps/enduro/internal/watcher"
	"github.com/artefactual-sdps/enduro/internal/workflow"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

const appName = "enduro"

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
		l := logger.WithName("migrate")
		if err := db.MigrateEnduroDatabase(l, enduroDatabase); err != nil {
			l.Error(err, "Enduro database migration failed.")
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
		l := logger.WithName("storage-migrate")
		if err := db.MigrateEnduroStorageDatabase(l, storageDatabase); err != nil {
			l.Error(err, "Storage database migration failed.")
			os.Exit(1)
		}
	}

	// Set up the Temporal client.
	tracingInterceptor, err := temporalsdk_contrib_opentelemetry.NewTracingInterceptor(
		temporalsdk_contrib_opentelemetry.TracerOptions{
			Tracer: tp.Tracer("temporal-sdk-go"),
		},
	)
	if err != nil {
		logger.Error(err, "Unable to create OpenTelemetry interceptor.")
		os.Exit(1)
	}
	temporalClient, err := temporalsdk_client.Dial(temporalsdk_client.Options{
		Namespace:    cfg.Temporal.Namespace,
		HostPort:     cfg.Temporal.Address,
		Logger:       temporal_tools.Logger(logger.WithName("temporal-client")),
		Interceptors: []temporalsdk_interceptor.ClientInterceptor{tracingInterceptor},
	})
	if err != nil {
		logger.Error(err, "Error creating Temporal client.")
		os.Exit(1)
	}

	// Set up the event services.
	ingestEventSvc, err := event.NewIngestEventServiceRedis(logger.WithName("ingest-events"), tp, &cfg.Event)
	if err != nil {
		logger.Error(err, "Error creating Ingest Event service.")
		os.Exit(1)
	}

	storageEventSvc, err := event.NewStorageEventServiceRedis(logger.WithName("storage-events"), tp, &cfg.Event)
	if err != nil {
		logger.Error(err, "Error creating Storage Event service.")
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

	// Set up the WebSocket/downloads ticket provider.
	var ticketProvider auth.TicketProvider
	{
		var store auth.TicketStore
		if cfg.API.Auth.Enabled {
			if cfg.API.Auth.Ticket != nil && cfg.API.Auth.Ticket.Redis != nil {
				var err error
				store, err = auth.NewRedisStore(ctx, tp, cfg.API.Auth.Ticket.Redis)
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

	// Set up the persistence service.
	var perSvc persistence.Service
	{
		drv := sqlcomment.NewDriver(
			sql.OpenDB(cfg.Database.Driver, enduroDatabase),
			sqlcomment.WithDriverVerTag(),
			sqlcomment.WithTags(sqlcomment.Tags{
				sqlcomment.KeyApplication: appName,
			}),
		)
		client := entdb.NewClient(entdb.Driver(drv))
		perSvc = persistence.WithTelemetry(
			entclient.New(logger.WithName("persistence"), client),
			tp.Tracer("persistence"),
		)
	}

	// Set up internal storage bucket.
	internalStorage, err := cfg.InternalStorage.OpenBucket(ctx)
	if err != nil {
		logger.Error(err, "Error setting up internal storage.")
		os.Exit(1)
	}
	defer internalStorage.Close()

	// Set up the ingest service.
	var ingestsvc ingest.Service
	{
		ingestsvc = ingest.NewService(
			logger.WithName("ingest"),
			enduroDatabase,
			temporalClient,
			ingestEventSvc,
			perSvc,
			tokenVerifier,
			ticketProvider,
			cfg.Temporal.TaskQueue,
			internalStorage,
			cfg.Upload.MaxSize,
			rand.Reader,
		)
	}

	// Set up the storage persistence layer.
	var storagePersistence storage_persistence.Storage
	{
		drv := sqlcomment.NewDriver(
			sql.OpenDB(cfg.Database.Driver, storageDatabase),
			sqlcomment.WithDriverVerTag(),
			sqlcomment.WithTags(sqlcomment.Tags{
				sqlcomment.KeyApplication: appName,
			}),
		)
		client := storage_entdb.NewClient(storage_entdb.Driver(drv))
		storagePersistence = storage_persistence.WithTelemetry(
			storage_entclient.NewClient(client),
			tp.Tracer("storage/persistence"),
		)
	}

	// Set up the storage service.
	var storagesvc storage.Service
	{
		storagesvc, err = storage.NewService(
			logger.WithName("storage"),
			cfg.Storage,
			storagePersistence,
			temporalClient,
			storageEventSvc,
			tokenVerifier,
			ticketProvider,
			rand.Reader,
		)
		if err != nil {
			logger.Error(err, "Error setting up storage service.")
			os.Exit(1)
		}
	}

	aboutsvc := about.NewService(
		logger.WithName("about"),
		cfg.Preservation.TaskQueue,
		cfg.Preprocessing,
		cfg.Poststorage,
		cfg.Upload,
		tokenVerifier,
	)

	// Set up the watcher service.
	var wsvc watcher.Service
	{
		wsvc, err = watcher.New(ctx, tp, logger.WithName("watcher"), &cfg.Watcher)
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
				srv = api.HTTPServer(logger, tp, &cfg.API, ingestsvc, storagesvc, aboutsvc)
				return srv.ListenAndServe()
			},
			func(err error) {
				ctx, cancel := context.WithTimeout(ctx, time.Second*5)
				defer cancel()
				_ = srv.Shutdown(ctx)
			},
		)
	}

	// Internal API server.
	// Recreate ingest and storage services with different
	// logger names and using &auth.NoopTokenVerifier{}.
	{
		ips := ingest.NewService(
			logger.WithName("internal-ingest"),
			enduroDatabase,
			temporalClient,
			ingestEventSvc,
			perSvc,
			&auth.NoopTokenVerifier{},
			ticketProvider,
			cfg.Temporal.TaskQueue,
			internalStorage,
			cfg.Upload.MaxSize,
			rand.Reader,
		)

		iss, err := storage.NewService(
			logger.WithName("internal-storage"),
			cfg.Storage,
			storagePersistence,
			temporalClient,
			storageEventSvc,
			&auth.NoopTokenVerifier{},
			ticketProvider,
			rand.Reader,
		)
		if err != nil {
			logger.Error(err, "Error setting up internal storage service.")
			os.Exit(1)
		}

		ias := about.NewService(
			logger.WithName("internal-about"),
			cfg.Preservation.TaskQueue,
			cfg.Preprocessing,
			cfg.Poststorage,
			cfg.Upload,
			&auth.NoopTokenVerifier{},
		)

		var srv *http.Server

		g.Add(
			func() error {
				srv = api.HTTPServer(logger, tp, &cfg.InternalAPI, ips, iss, ias)
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
			g.Add(
				func() error {
					for {
						select {
						case <-done:
							return nil
						default:
							ctx, span := tp.Tracer("enduro").Start(ctx, "watcher.poll")
							event, clean, err := w.Watch(ctx)
							if err != nil {
								if !errors.Is(err, watcher.ErrWatchTimeout) {
									logger.Error(err, "Error monitoring watcher interface.", "watcher", w)
									span.RecordError(err)
									span.SetStatus(codes.Error, err.Error())
								}
								span.End()
								continue
							}
							logger.V(1).
								Info("Starting new workflow", "watcher", event.WatcherName, "bucket", event.Bucket, "key", event.Key, "dir", event.IsDir)
							go func() {
								defer span.End()
								req := ingest.ProcessingWorkflowRequest{
									WatcherName:     event.WatcherName,
									RetentionPeriod: event.RetentionPeriod,
									CompletedDir:    event.CompletedDir,
									Key:             event.Key,
									IsDir:           event.IsDir,
									Type:            event.WorkflowType,
									SIPUUID:         uuid.New(),
									SIPName:         event.Key,
								}
								if err := ingest.InitProcessingWorkflow(ctx, temporalClient, cfg.Temporal.TaskQueue, &req); err != nil {
									logger.Error(err, "Error initializing processing workflow.")
									span.RecordError(err)
									span.SetStatus(codes.Error, err.Error())
								} else {
									if err := clean(ctx); err != nil {
										span.RecordError(err)
										span.SetStatus(codes.Error, err.Error())
									}
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
		workerOpts := temporalsdk_worker.Options{
			Interceptors: []temporalsdk_interceptor.WorkerInterceptor{
				temporal_tools.NewLoggerInterceptor(logger.WithName("worker")),
			},
		}
		w := temporalsdk_worker.New(temporalClient, cfg.Temporal.TaskQueue, workerOpts)
		if err != nil {
			logger.Error(err, "Error creating Temporal worker.")
			os.Exit(1)
		}

		w.RegisterWorkflowWithOptions(
			workflow.NewProcessingWorkflow(cfg, rand.Reader, ingestsvc, wsvc).Execute,
			temporalsdk_workflow.RegisterOptions{Name: ingest.ProcessingWorkflowName},
		)
		w.RegisterActivityWithOptions(
			activities.NewDeleteOriginalActivity(wsvc).Execute,
			temporalsdk_activity.RegisterOptions{Name: activities.DeleteOriginalActivityName},
		)
		w.RegisterActivityWithOptions(
			activities.NewDisposeOriginalActivity(wsvc).Execute,
			temporalsdk_activity.RegisterOptions{Name: activities.DisposeOriginalActivityName},
		)

		w.RegisterWorkflowWithOptions(
			storage_workflows.NewStorageDeleteWorkflow(storagesvc).Execute,
			temporalsdk_workflow.RegisterOptions{Name: storage.StorageDeleteWorkflowName},
		)
		w.RegisterWorkflowWithOptions(
			storage_workflows.NewStorageUploadWorkflow().Execute,
			temporalsdk_workflow.RegisterOptions{Name: storage.StorageUploadWorkflowName},
		)
		w.RegisterWorkflowWithOptions(
			storage_workflows.NewStorageMoveWorkflow(storagesvc).Execute,
			temporalsdk_workflow.RegisterOptions{Name: storage.StorageMoveWorkflowName},
		)

		w.RegisterActivityWithOptions(
			storage_activities.NewCopyToPermanentLocationActivity(storagesvc).Execute,
			temporalsdk_activity.RegisterOptions{Name: storage.CopyToPermanentLocationActivityName},
		)
		w.RegisterActivityWithOptions(
			storage_activities.NewDeleteFromAMSSLocationActivity().Execute,
			temporalsdk_activity.RegisterOptions{Name: storage.DeleteFromAMSSLocationActivityName},
		)

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
