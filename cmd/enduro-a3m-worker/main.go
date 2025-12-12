package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ariga.io/sqlcomment"
	"entgo.io/ent/dialect/sql"
	"github.com/artefactual-sdps/temporal-activities/archiveextract"
	"github.com/artefactual-sdps/temporal-activities/archivezip"
	"github.com/artefactual-sdps/temporal-activities/bagvalidate"
	"github.com/artefactual-sdps/temporal-activities/bucketcopy"
	"github.com/artefactual-sdps/temporal-activities/bucketdelete"
	"github.com/artefactual-sdps/temporal-activities/bucketdownload"
	"github.com/artefactual-sdps/temporal-activities/bucketupload"
	"github.com/artefactual-sdps/temporal-activities/removepaths"
	"github.com/artefactual-sdps/temporal-activities/xmlvalidate"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"go.artefactual.dev/tools/log"
	temporal_tools "go.artefactual.dev/tools/temporal"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_client "go.temporal.io/sdk/client"
	temporalsdk_contrib_opentelemetry "go.temporal.io/sdk/contrib/opentelemetry"
	temporalsdk_interceptor "go.temporal.io/sdk/interceptor"
	temporalsdk_worker "go.temporal.io/sdk/worker"
	goahttp "goa.design/goa/v3/http"

	"github.com/artefactual-sdps/enduro/internal/a3m"
	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goahttpstorage "github.com/artefactual-sdps/enduro/internal/api/gen/http/storage/client"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/db"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	entclient "github.com/artefactual-sdps/enduro/internal/persistence/ent/client"
	entdb "github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/sipsource"
	"github.com/artefactual-sdps/enduro/internal/telemetry"
	"github.com/artefactual-sdps/enduro/internal/temporal"
	"github.com/artefactual-sdps/enduro/internal/version"
	"github.com/artefactual-sdps/enduro/internal/watcher"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

const (
	appName = "enduro-a3m-worker"
)

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up the tracer provider.
	tp, shutdown, err := telemetry.TracerProvider(ctx, logger, cfg.Telemetry, appName, version.Long)
	if err != nil {
		logger.Error(err, "Error creating tracer provider.")
		os.Exit(1)
	}
	defer func() { _ = shutdown(ctx) }()

	enduroDatabase, err := db.Connect(ctx, tp, cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		logger.Error(err, "Enduro database configuration failed.")
		os.Exit(1)
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

	// Set up internal storage.
	internalStorage, err := cfg.InternalStorage.OpenBucket(ctx)
	if err != nil {
		logger.Error(err, "Error setting up internal bucket.")
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
			Logger:             logger.WithName("ingest"),
			DB:                 enduroDatabase,
			TemporalClient:     temporalClient,
			EventService:       ingestEventSvc,
			PersistenceService: perSvc,
			TokenVerifier:      &auth.NoopTokenVerifier{},
			TicketProvider:     nil,
			TaskQueue:          cfg.Temporal.TaskQueue,
			InternalStorage:    internalStorage,
			UploadMaxSize:      0,
			Rander:             rand.Reader,
			SIPSource:          sipSource,
		})
	}

	// Set up the watcher service.
	var wsvc watcher.Service
	{
		wsvc, err = watcher.New(ctx, tp, logger.WithName("watcher"), &cfg.Watcher)
		if err != nil {
			logger.Error(err, "Error setting up watchers.")
			os.Exit(1)
		}
	}

	// Set up the a3m gRPC client.
	var a3mClient *a3m.Client
	{
		a3mClient, err = a3m.NewClient(ctx, tp, cfg.A3m.Address)
		if err != nil {
			logger.Error(err, "Error setting up the a3m client.")
			os.Exit(1)
		}
	}

	var g run.Group

	// Activity worker.
	{
		done := make(chan struct{})
		workerOpts := temporalsdk_worker.Options{
			DisableWorkflowWorker: true,
			EnableSessionWorker:   true,
			// Allow only one concurrent session as a3m is not concurrency safe.
			MaxConcurrentSessionExecutionSize: 1,
			// Allow two concurrent activities, recreated sessions require a long
			// running activity slot in the same worker where the session is held.
			MaxConcurrentActivityExecutionSize: 2,
			Interceptors: []temporalsdk_interceptor.WorkerInterceptor{
				temporal_tools.NewLoggerInterceptor(logger),
			},
		}
		w := temporalsdk_worker.New(temporalClient, temporal.A3mWorkerTaskQueue, workerOpts)
		if err != nil {
			logger.Error(err, "Error creating Temporal worker.")
			os.Exit(1)
		}

		w.RegisterActivityWithOptions(
			activities.NewDownloadActivity(tp.Tracer(activities.DownloadActivityName), wsvc).Execute,
			temporalsdk_activity.RegisterOptions{Name: activities.DownloadActivityName},
		)
		// TODO: At some point there may be multiple SIP sources, this activity should
		// work similar to the watched bucket download activity and use the source ID
		// to determine which bucket to use. Alternatively, we could register multiple
		// copies of this activity, one per source.
		w.RegisterActivityWithOptions(
			bucketdownload.New(sipSource.Bucket).Execute,
			temporalsdk_activity.RegisterOptions{Name: activities.DownloadFromSIPSourceActivityName},
		)
		w.RegisterActivityWithOptions(
			bucketdownload.New(internalStorage).Execute,
			temporalsdk_activity.RegisterOptions{Name: activities.DownloadFromInternalBucketActivityName},
		)
		w.RegisterActivityWithOptions(
			activities.NewGetSIPExtensionActivity().Execute,
			temporalsdk_activity.RegisterOptions{Name: activities.GetSIPExtensionActivityName},
		)
		w.RegisterActivityWithOptions(
			archiveextract.New(cfg.ExtractActivity).Execute,
			temporalsdk_activity.RegisterOptions{Name: archiveextract.Name},
		)
		w.RegisterActivityWithOptions(
			xmlvalidate.New(xmlvalidate.NewXMLLintValidator()).Execute,
			temporalsdk_activity.RegisterOptions{Name: xmlvalidate.Name},
		)
		w.RegisterActivityWithOptions(
			activities.NewClassifySIPActivity().Execute,
			temporalsdk_activity.RegisterOptions{Name: activities.ClassifySIPActivityName},
		)
		w.RegisterActivityWithOptions(
			bagvalidate.New(nil).Execute,
			temporalsdk_activity.RegisterOptions{Name: bagvalidate.Name},
		)
		w.RegisterActivityWithOptions(
			activities.NewBundleActivity().Execute,
			temporalsdk_activity.RegisterOptions{Name: activities.BundleActivityName},
		)
		w.RegisterActivityWithOptions(
			a3m.NewCreateAIPActivity(
				tp.Tracer(a3m.CreateAIPActivityName),
				a3mClient.TransferClient,
				&cfg.A3m,
				ingestsvc,
			).Execute,
			temporalsdk_activity.RegisterOptions{Name: a3m.CreateAIPActivityName},
		)
		w.RegisterActivityWithOptions(
			removepaths.New().Execute,
			temporalsdk_activity.RegisterOptions{Name: removepaths.Name},
		)

		httpClient := cleanhttp.DefaultPooledClient()
		httpClient.Transport = otelhttp.NewTransport(
			httpClient.Transport,
			otelhttp.WithTracerProvider(tp),
			otelhttp.WithClientTrace(func(ctx context.Context) *httptrace.ClientTrace {
				return otelhttptrace.NewClientTrace(ctx)
			}),
		)
		storageHttpClient := goahttpstorage.NewClient(
			"http",
			cfg.Storage.EnduroAddress,
			httpClient,
			goahttp.RequestEncoder,
			goahttp.ResponseDecoder,
			false,
			nil,
			nil,
		)
		storageClient := goastorage.NewClient(
			storageHttpClient.MonitorRequest(),
			storageHttpClient.Monitor(),
			storageHttpClient.ListAips(),
			storageHttpClient.CreateAip(),
			storageHttpClient.SubmitAip(),
			storageHttpClient.SubmitAipComplete(),
			storageHttpClient.DownloadAipRequest(),
			storageHttpClient.DownloadAip(),
			storageHttpClient.MoveAip(),
			storageHttpClient.MoveAipStatus(),
			storageHttpClient.RejectAip(),
			storageHttpClient.ShowAip(),
			storageHttpClient.ListAipWorkflows(),
			storageHttpClient.RequestAipDeletion(),
			storageHttpClient.ReviewAipDeletion(),
			storageHttpClient.CancelAipDeletion(),
			storageHttpClient.AipDeletionReportRequest(),
			storageHttpClient.AipDeletionReport(),
			storageHttpClient.ListLocations(),
			storageHttpClient.CreateLocation(),
			storageHttpClient.ShowLocation(),
			storageHttpClient.ListLocationAips(),
		)
		w.RegisterActivityWithOptions(
			activities.NewUploadActivity(storageClient).Execute,
			temporalsdk_activity.RegisterOptions{Name: activities.UploadActivityName},
		)

		w.RegisterActivityWithOptions(
			activities.NewMoveToPermanentStorageActivity(storageClient).Execute,
			temporalsdk_activity.RegisterOptions{Name: activities.MoveToPermanentStorageActivityName},
		)
		w.RegisterActivityWithOptions(
			activities.NewPollMoveToPermanentStorageActivity(storageClient).Execute,
			temporalsdk_activity.RegisterOptions{Name: activities.PollMoveToPermanentStorageActivityName},
		)
		w.RegisterActivityWithOptions(
			activities.NewRejectSIPActivity(storageClient).Execute,
			temporalsdk_activity.RegisterOptions{Name: activities.RejectSIPActivityName},
		)
		w.RegisterActivityWithOptions(
			archivezip.New().Execute,
			temporalsdk_activity.RegisterOptions{Name: archivezip.Name},
		)
		w.RegisterActivityWithOptions(
			bucketcopy.New(internalStorage).Execute,
			temporalsdk_activity.RegisterOptions{Name: bucketcopy.Name},
		)
		w.RegisterActivityWithOptions(
			bucketdelete.New(internalStorage).Execute,
			temporalsdk_activity.RegisterOptions{Name: bucketdelete.Name},
		)
		w.RegisterActivityWithOptions(
			bucketupload.New(internalStorage).Execute,
			temporalsdk_activity.RegisterOptions{Name: bucketupload.Name},
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
