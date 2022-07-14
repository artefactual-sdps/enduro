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
	goahttp "goa.design/goa/v3/http"
	goahttpmwr "goa.design/goa/v3/http/middleware"
	"goa.design/goa/v3/middleware"

	"github.com/artefactual-sdps/enduro/internal/api/gen/batch"
	batchsvr "github.com/artefactual-sdps/enduro/internal/api/gen/http/batch/server"
	packagesvr "github.com/artefactual-sdps/enduro/internal/api/gen/http/package_/server"
	storagesvr "github.com/artefactual-sdps/enduro/internal/api/gen/http/storage/server"
	swaggersvr "github.com/artefactual-sdps/enduro/internal/api/gen/http/swagger/server"
	"github.com/artefactual-sdps/enduro/internal/api/gen/package_"
	"github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	intbatch "github.com/artefactual-sdps/enduro/internal/batch"
	intpkg "github.com/artefactual-sdps/enduro/internal/package_"
	intstorage "github.com/artefactual-sdps/enduro/internal/storage"
)

func HTTPServer(
	logger logr.Logger, config *Config,
	batchsvc intbatch.Service,
	pkgsvc intpkg.Service,
	storagesvc intstorage.Service,
) *http.Server {
	dec := goahttp.RequestDecoder
	enc := goahttp.ResponseEncoder
	mux := goahttp.NewMuxer()

	websocketUpgrader := &websocket.Upgrader{
		HandshakeTimeout: time.Second,
		CheckOrigin:      sameOriginChecker(logger),
	}

	// Batch service.
	batchEndpoints := batch.NewEndpoints(batchsvc)
	batchErrorHandler := errorHandler(logger, "Batch error.")
	batchServer := batchsvr.New(batchEndpoints, mux, dec, enc, batchErrorHandler, nil)
	batchsvr.Mount(mux, batchServer)

	// Package service.
	packageEndpoints := package_.NewEndpoints(pkgsvc.Goa())
	packageErrorHandler := errorHandler(logger, "Package error.")
	packageServer := packagesvr.New(packageEndpoints, mux, dec, enc, packageErrorHandler, nil, websocketUpgrader, nil)
	packagesvr.Mount(mux, packageServer)

	// Storage service.
	storageEndpoints := storage.NewEndpoints(storagesvc)
	storageErrorHandler := errorHandler(logger, "Storage error.")
	storageServer := storagesvr.New(storageEndpoints, mux, dec, enc, storageErrorHandler, nil)
	storageServer.Download = storagesvc.HTTPDownload(mux, dec)
	storagesvr.Mount(mux, storageServer)

	// Swagger service.
	swaggerService := swaggersvr.New(nil, nil, nil, nil, nil, nil, nil)
	swaggersvr.Mount(mux, swaggerService)

	// Global middlewares.
	var handler http.Handler = mux
	handler = recoverMiddleware(logger)(handler)
	handler = goahttpmwr.RequestID()(handler)
	handler = versionHeaderMiddleware(config.AppVersion)(handler)
	if config.Debug {
		handler = goahttpmwr.Log(loggerAdapter(logger))(handler)
		handler = goahttpmwr.Debug(mux, os.Stdout)(handler)
	}

	return &http.Server{
		Addr:        config.Listen,
		Handler:     handler,
		ReadTimeout: time.Second * 5,
		// WriteTimeout is set to 0 because we have streaming endpoints.
		// https://github.com/golang/go/issues/16100#issuecomment-285573480
		WriteTimeout: 0,
		IdleTimeout:  time.Second * 120,
	}
}

type errorMessage struct {
	RequestID string
	Error     error
}

// errorHandler returns a function that writes and logs the given error.
// The function also writes and logs the error unique ID so that it's possible
// to correlate.
func errorHandler(logger logr.Logger, msg string) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		reqID, ok := ctx.Value(middleware.RequestIDKey).(string)
		if !ok {
			reqID = "unknown"
		}
		_ = json.NewEncoder(w).Encode(&errorMessage{RequestID: reqID})
		logger.Error(err, "Package service error.", "reqID", reqID)
	}
}
