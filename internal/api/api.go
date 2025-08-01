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
	"net/http"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"go.artefactual.dev/tools/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
	goahttp "goa.design/goa/v3/http"
	goahttpmwr "goa.design/goa/v3/http/middleware"
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
	tp trace.TracerProvider,
	config *Config,
	ingestsvc intingest.Service,
	storagesvc intstorage.Service,
	aboutsvc *intabout.Service,
) *http.Server {
	dec := goahttp.RequestDecoder
	enc := goahttp.ResponseEncoder
	mux := goahttp.NewMuxer()

	websocketUpgrader := &websocket.Upgrader{
		HandshakeTimeout: time.Second,
		CheckOrigin:      sameOriginChecker(logger),
	}

	// Ingest service.
	ingestEndpoints := ingest.NewEndpoints(ingestsvc.Goa())
	ingestErrorHandler := errorHandler(logger, "Ingest error.")
	ingestServer := ingestsvr.New(ingestEndpoints, mux, dec, enc, ingestErrorHandler, nil, websocketUpgrader, nil)
	ingestServer.DownloadSip = middleware.WriteTimeout(0)(ingestServer.DownloadSip)
	ingestServer.UploadSip = middleware.WriteTimeout(0)(ingestServer.UploadSip)
	ingestServer.UploadSip = middleware.ReadTimeout(0)(ingestServer.UploadSip)
	ingestsvr.Mount(mux, ingestServer)

	// Storage service.
	storageEndpoints := storage.NewEndpoints(storagesvc)
	storageErrorHandler := errorHandler(logger, "Storage error.")
	storageServer := storagesvr.New(storageEndpoints, mux, dec, enc, storageErrorHandler, nil, websocketUpgrader, nil)
	storageServer.DownloadAip = middleware.WriteTimeout(0)(storageServer.DownloadAip)
	storagesvr.Mount(mux, storageServer)

	// About service.
	aboutEndpoints := about.NewEndpoints(aboutsvc)
	aboutErrorHandler := errorHandler(logger, "About error.")
	aboutServer := aaboutsvr.New(aboutEndpoints, mux, dec, enc, aboutErrorHandler, nil)
	aaboutsvr.Mount(mux, aboutServer)

	// Global middlewares.
	var handler http.Handler = mux
	handler = middleware.Recover(logger)(handler)
	handler = otelhttp.NewHandler(handler, "api", otelhttp.WithTracerProvider(tp))
	handler = middleware.VersionHeader("X-Enduro-Version", version.Short)(handler)
	if config.Debug {
		handler = goahttpmwr.Log(loggerAdapter(logger))(handler) //nolint SA1019: deprecated - use OpenTelemetry.
		handler = goahttpmwr.Debug(mux, os.Stdout)(handler)
	}

	return &http.Server{
		Addr:         config.Listen,
		Handler:      handler,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
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

		// Only write the error if the connection is not hijacked.
		var ws bool
		if _, err := w.Write(nil); err == http.ErrHijacked {
			ws = true
		} else {
			_ = json.NewEncoder(w).Encode(&errorMessage{RequestID: reqID})
		}

		logger.Error(err, "Service error.", "reqID", reqID, "ws", ws, "msg", msg)
	}
}
