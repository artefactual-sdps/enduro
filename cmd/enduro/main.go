package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ariga.io/sqlcomment"
	"entgo.io/ent/dialect/sql"
	"github.com/artefactual-sdps/temporal-activities/bucketdelete"
	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
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
	"github.com/artefactual-sdps/enduro/internal/auditlog"
	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/db"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	entclient "github.com/artefactual-sdps/enduro/internal/persistence/ent/client"
	entdb "github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/sipsource"
	"github.com/artefactual-sdps/enduro/internal/storage"
	storage_activities "github.com/artefactual-sdps/enduro/internal/storage/activities"
	"github.com/artefactual-sdps/enduro/internal/storage/pdfs"
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

	// Set up the audit logger early on so is closed after most other services
	// close. Deferred functions run in "first in, last out" order.
	auditLogger := auditlog.NewFromConfig(cfg.Auditlog)
	defer auditLogger.Close()

	if cfg.Auditlog.Filepath != "" {
		logger.V(1).Info("Audit logging enabled.", "path", cfg.Auditlog.Filepath)
	} else {
		logger.V(1).Info("Audit logging disabled.")
	}

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

	// Set up the ingest event service.
	ingestEventSvc, err := event.NewServiceRedis(
		logger.WithName("ingest-events"),
		tp,
		cfg.Event.RedisAddress,
		cfg.Event.RedisChannel,
		&ingest.EventSerializer{},
	)
	if err != nil {
		logger.Error(err, "Error creating Ingest Event service.")
		os.Exit(1)
	}

	// Set up the storage event service.
	storageEventSvc, err := event.NewServiceRedis(
		logger.WithName("storage-events"),
		tp,
		cfg.Storage.Event.RedisAddress,
		cfg.Storage.Event.RedisChannel,
		&storage.EventSerializer{},
	)
	if err != nil {
		logger.Error(err, "Error creating Storage Event service.")
		os.Exit(1)
	}

	// Set up the OIDC token verifier.
	var tokenVerifier auth.TokenVerifier
	{
		if cfg.API.Auth.Enabled {
			tokenVerifier, err = auth.NewOIDCTokenVerifiers(ctx, cfg.API.Auth.OIDC)
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

	// Set up a SIP source, if one is configured.
	sipSource, err := sipsource.NewBucketSource(ctx, &cfg.SIPSource)
	if err != nil {
		logger.Error(err, "Error setting up SIP source.")
		os.Exit(1)
	}
	defer sipSource.Close()

	// Set up the ingest service.
	var ingestsvc ingest.Service
	{
		ingestsvc = ingest.NewService(ingest.ServiceParams{
			Logger:                logger.WithName("ingest"),
			DB:                    enduroDatabase,
			TemporalClient:        temporalClient,
			EventService:          ingestEventSvc,
			PersistenceService:    perSvc,
			TokenVerifier:         tokenVerifier,
			TicketProvider:        ticketProvider,
			TaskQueue:             cfg.Temporal.TaskQueue,
			InternalStorage:       internalStorage,
			UploadMaxSize:         cfg.Upload.MaxSize,
			UploadRetentionPeriod: cfg.Upload.RetentionPeriod,
			Rander:                rand.Reader,
			SIPSource:             sipSource,
			AuditLogger:           auditLogger,
		})
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
			auditLogger,
		)
		if err != nil {
			logger.Error(err, "Error setting up storage service.")
			os.Exit(1)
		}
	}

	aboutsvc := about.NewService(
		logger.WithName("about"),
		cfg.Preservation.TaskQueue,
		cfg.ChildWorkflows,
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
		ips := ingest.NewService(ingest.ServiceParams{
			Logger:                logger.WithName("internal-ingest"),
			DB:                    enduroDatabase,
			TemporalClient:        temporalClient,
			EventService:          ingestEventSvc,
			PersistenceService:    perSvc,
			TokenVerifier:         &auth.NoopTokenVerifier{},
			TicketProvider:        ticketProvider,
			TaskQueue:             cfg.Temporal.TaskQueue,
			InternalStorage:       internalStorage,
			UploadMaxSize:         cfg.Upload.MaxSize,
			UploadRetentionPeriod: cfg.Upload.RetentionPeriod,
			Rander:                rand.Reader,
			SIPSource:             sipSource,
			AuditLogger:           auditLogger,
		})

		iss, err := storage.NewService(
			logger.WithName("internal-storage"),
			cfg.Storage,
			storagePersistence,
			temporalClient,
			storageEventSvc,
			&auth.NoopTokenVerifier{},
			ticketProvider,
			rand.Reader,
			auditLogger,
		)
		if err != nil {
			logger.Error(err, "Error setting up internal storage service.")
			os.Exit(1)
		}

		ias := about.NewService(
			logger.WithName("internal-about"),
			cfg.Preservation.TaskQueue,
			cfg.ChildWorkflows,
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

		// Ingest processing workflow and activities.
		w.RegisterWorkflowWithOptions(
			workflow.NewProcessingWorkflow(cfg, rand.Reader, ingestsvc, wsvc).Execute,
			temporalsdk_workflow.RegisterOptions{Name: ingest.ProcessingWorkflowName},
		)
		w.RegisterActivityWithOptions(
			activities.NewDeleteOriginalActivity(wsvc).Execute,
			temporalsdk_activity.RegisterOptions{Name: activities.DeleteOriginalActivityName},
		)
		// TODO: At some point there may be multiple SIP sources, this activity should
		// work similar to the watched bucket delete original activity and use the
		// source ID to determine which bucket to use. Alternatively, we could register
		// multiple copies of this activity, one per source.
		w.RegisterActivityWithOptions(
			bucketdelete.New(sipSource.Bucket).Execute,
			temporalsdk_activity.RegisterOptions{Name: activities.DeleteOriginalFromSIPSourceActivityName},
		)
		w.RegisterActivityWithOptions(
			bucketdelete.New(internalStorage).Execute,
			temporalsdk_activity.RegisterOptions{Name: activities.DeleteOriginalFromInternalBucketActivityName},
		)
		w.RegisterActivityWithOptions(
			activities.NewDisposeOriginalActivity(wsvc).Execute,
			temporalsdk_activity.RegisterOptions{Name: activities.DisposeOriginalActivityName},
		)

		// Ingest batch workflow and activities.
		w.RegisterWorkflowWithOptions(
			workflow.NewBatchWorkflow(cfg, rand.Reader, ingestsvc, temporalClient).Execute,
			temporalsdk_workflow.RegisterOptions{Name: ingest.BatchWorkflowName},
		)
		w.RegisterActivityWithOptions(
			activities.NewPollSIPStatusesActivity(ingestsvc, time.Second*60).Execute,
			temporalsdk_activity.RegisterOptions{Name: activities.PollSIPStatusesActivityName},
		)

		// Storage workflows and activities.
		w.RegisterWorkflowWithOptions(
			storage_workflows.NewStorageDeleteWorkflow(
				cfg.Storage.AIPDeletion,
				storagesvc,
			).Execute,
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
			storage_activities.NewDeleteFromAMSSLocationActivity(
				cfg.Storage.AIPDeletion.ApproveAMSS,
				time.Second*60,
			).Execute,
			temporalsdk_activity.RegisterOptions{Name: storage.DeleteFromAMSSLocationActivityName},
		)
		w.RegisterActivityWithOptions(
			storage_activities.NewAIPDeletionReportActivity(
				clockwork.NewRealClock(),
				cfg.Storage.AIPDeletion,
				storagesvc,
				pdfs.NewPDFCPU(),
			).Execute,
			temporalsdk_activity.RegisterOptions{
				Name: storage_activities.AIPDeletionReportActivityName,
			},
		)

		g.Add(
			func() error {
				auditLogger.Log(ctx, &auditlog.Event{
					Level: slog.LevelInfo,
					Msg:   "Enduro starting",
					Type:  "system",
				})

				if err := w.Start(); err != nil {
					return err
				}
				<-done
				return nil
			},
			func(err error) {
				auditLogger.Log(ctx, &auditlog.Event{
					Level: slog.LevelInfo,
					Msg:   "Enduro stopping",
					Type:  "system",
				})

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
		log.Sync(logger)
		os.Exit(1)
	}
	logger.Info("Bye!")
}
