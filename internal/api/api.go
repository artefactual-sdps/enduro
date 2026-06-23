/*
Package api contains the API server.

HTTP is the only transport supported at the moment.

The design package is the Goa design package while the gen package contains all
the generated code produced with goa gen.
*/
package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"go.artefactual.dev/tools/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
	goahttp "goa.design/goa/v3/http"
	goamiddleware "goa.design/goa/v3/middleware"

	intabout "github.com/artefactual-sdps/enduro/internal/about"
	"github.com/artefactual-sdps/enduro/internal/api/gen/about"
	aaboutsvr "github.com/artefactual-sdps/enduro/internal/api/gen/http/about/server"
	ingestsvr "github.com/artefactual-sdps/enduro/internal/api/gen/http/ingest/server"
	storagesvr "github.com/artefactual-sdps/enduro/internal/api/gen/http/storage/server"
	"github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	intingest "github.com/artefactual-sdps/enduro/internal/ingest"
	intstorage "github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/version"
)

func HTTPServer(
	logger logr.Logger,
	apiLogger *slog.Logger,
	tp trace.TracerProvider,
	config *Config,
	ingestsvc intingest.Service,
	storagesvc intstorage.Service,
	aboutsvc *intabout.Service,
) *http.Server {
	dec := goahttp.RequestDecoder
	enc := goahttp.ResponseEncoder
	mux := goahttp.NewMuxer()
	mux.Use(otelhttp.NewMiddleware("api", otelhttp.WithTracerProvider(tp)))
	mux.Use(middleware.Recover(logger))

	operationInterceptors := newOperationInterceptors(logger)

	// Ingest service.
	ingestEndpoints := ingest.NewEndpoints(
		ingestsvc,
		&ingestServerInterceptors{operationInterceptors: operationInterceptors},
	)
	ingestErrorHandler := errorHandler(logger, "Ingest error.")
	ingestServer := ingestsvr.New(ingestEndpoints, mux, dec, enc, ingestErrorHandler, nil)
	ingestServer.Monitor = middleware.WriteTimeout(0)(ingestServer.Monitor)
	ingestServer.DownloadSip = middleware.WriteTimeout(0)(ingestServer.DownloadSip)
	ingestServer.UploadSip = middleware.WriteTimeout(0)(ingestServer.UploadSip)
	ingestServer.UploadSip = middleware.ReadTimeout(0)(ingestServer.UploadSip)
	ingestsvr.Mount(mux, ingestServer)

	// Storage service.
	storageEndpoints := storage.NewEndpoints(
		storagesvc,
		&storageServerInterceptors{operationInterceptors: operationInterceptors},
	)
	storageErrorHandler := errorHandler(logger, "Storage error.")
	storageServer := storagesvr.New(storageEndpoints, mux, dec, enc, storageErrorHandler, nil)
	storageServer.Monitor = middleware.WriteTimeout(0)(storageServer.Monitor)
	// Streaming downloads can legitimately take longer than the API write
	// timeout while bytes are sent to slow clients or proxies.
	storageServer.DownloadAip = middleware.WriteTimeout(0)(storageServer.DownloadAip)
	storageServer.AipDeletionReport = middleware.WriteTimeout(0)(storageServer.AipDeletionReport)
	storagesvr.Mount(mux, storageServer)

	// About service.
	aboutEndpoints := about.NewEndpoints(
		aboutsvc,
		&aboutServerInterceptors{operationInterceptors: operationInterceptors},
	)
	aboutErrorHandler := errorHandler(logger, "About error.")
	aboutServer := aaboutsvr.New(aboutEndpoints, mux, dec, enc, aboutErrorHandler, nil)
	aaboutsvr.Mount(mux, aboutServer)

	// Global middlewares.
	var handler http.Handler = mux
	handler = middleware.VersionHeader("X-Enduro-Version", version.Short)(handler)

	// Add logging middleware if an API logger is configured. The log level is
	// set to the configured log level.
	if apiLogger != nil {
		handler = requestLogger(apiLogger, config.Log.Level)(handler)
	}

	return &http.Server{
		Addr:        config.Listen,
		Handler:     handler,
		ReadTimeout: time.Second * 5,
		// Keep this above defaultAPIOperationTimeout so normal handlers can
		// return a timeout response. Streaming handlers opt out above.
		WriteTimeout: time.Second * 7,
		IdleTimeout:  time.Second * 120,
	}
}

type errorMessage struct {
	RequestID string
	Error     error
}

// errorHandler returns a function that writes and logs the given error
// including the request ID.
func errorHandler(logger logr.Logger, msg string) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		reqID, ok := ctx.Value(goamiddleware.RequestIDKey).(string)
		if !ok {
			reqID = "unknown"
		}

		_ = json.NewEncoder(w).Encode(&errorMessage{RequestID: reqID})

		logger.Error(err, "Service error.", "reqID", reqID, "msg", msg)
	}
}
