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
	"net/url"
	"os"
	"time"
	"unicode/utf8"

	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	goahttp "goa.design/goa/v3/http"
	goahttpmwr "goa.design/goa/v3/http/middleware"
	"goa.design/goa/v3/middleware"

	"github.com/artefactual-labs/enduro/internal/api/gen/batch"
	batchsvr "github.com/artefactual-labs/enduro/internal/api/gen/http/batch/server"
	packagesvr "github.com/artefactual-labs/enduro/internal/api/gen/http/package_/server"
	storagesvr "github.com/artefactual-labs/enduro/internal/api/gen/http/storage/server"
	swaggersvr "github.com/artefactual-labs/enduro/internal/api/gen/http/swagger/server"
	"github.com/artefactual-labs/enduro/internal/api/gen/package_"
	"github.com/artefactual-labs/enduro/internal/api/gen/storage"
	intbatch "github.com/artefactual-labs/enduro/internal/batch"
	intpkg "github.com/artefactual-labs/enduro/internal/package_"
	intstorage "github.com/artefactual-labs/enduro/internal/storage"
)

func HTTPServer(
	logger logr.Logger, config *Config,
	batchsvc intbatch.Service,
	pkgsvc intpkg.Service,
	storagesvc intstorage.Service,
) *http.Server {
	dec := goahttp.RequestDecoder
	enc := goahttp.ResponseEncoder
	var mux goahttp.Muxer = goahttp.NewMuxer()

	websocketUpgrader := &websocket.Upgrader{
		HandshakeTimeout: time.Second,
		CheckOrigin:      sameOriginChecker(logger),
	}

	// Batch service.
	var batchEndpoints *batch.Endpoints = batch.NewEndpoints(batchsvc)
	batchErrorHandler := errorHandler(logger, "Batch error.")
	var batchServer *batchsvr.Server = batchsvr.New(batchEndpoints, mux, dec, enc, batchErrorHandler, nil)
	batchsvr.Mount(mux, batchServer)

	// Package service.
	var packageEndpoints *package_.Endpoints = package_.NewEndpoints(pkgsvc.Goa())
	packageErrorHandler := errorHandler(logger, "Package error.")
	var packageServer *packagesvr.Server = packagesvr.New(packageEndpoints, mux, dec, enc, packageErrorHandler, nil, websocketUpgrader, nil)
	packagesvr.Mount(mux, packageServer)

	// Storage service.
	var storageEndpoints *storage.Endpoints = storage.NewEndpoints(storagesvc)
	storageErrorHandler := errorHandler(logger, "Storage error.")
	var storageServer *storagesvr.Server = storagesvr.New(storageEndpoints, mux, dec, enc, storageErrorHandler, nil)
	storagesvr.Mount(mux, storageServer)

	// Swagger service.
	var swaggerService *swaggersvr.Server = swaggersvr.New(nil, nil, nil, nil, nil, nil, nil)
	swaggersvr.Mount(mux, swaggerService)

	// Global middlewares.
	var handler http.Handler = mux
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

func versionHeaderMiddleware(version string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Enduro-Version", version)
			h.ServeHTTP(w, r)
		})
	}
}

func sameOriginChecker(logger logr.Logger) func(r *http.Request) bool {
	return func(r *http.Request) bool {
		origin := r.Header["Origin"]
		if len(origin) == 0 {
			return true
		}
		u, err := url.Parse(origin[0])
		if err != nil {
			logger.V(1).Info("WebSocket client rejected (origin parse error)", "err", err)
			return false
		}
		eq := equalASCIIFold(u.Host, r.Host)
		if !eq {
			logger.V(1).Info("WebSocket client rejected (origin and host not equal)", "origin-host", u.Host, "request-host", r.Host)
		}
		return eq
	}
}

// equalASCIIFold returns true if s is equal to t with ASCII case folding as
// defined in RFC 4790.
func equalASCIIFold(s, t string) bool {
	for s != "" && t != "" {
		sr, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		tr, size := utf8.DecodeRuneInString(t)
		t = t[size:]
		if sr == tr {
			continue
		}
		if 'A' <= sr && sr <= 'Z' {
			sr = sr + 'a' - 'A'
		}
		if 'A' <= tr && tr <= 'Z' {
			tr = tr + 'a' - 'A'
		}
		if sr != tr {
			return false
		}
	}
	return s == t
}
