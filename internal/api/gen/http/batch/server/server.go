// Code generated by goa v3.7.6, DO NOT EDIT.
//
// batch HTTP server
//
// Command:
// $ goa-v3.7.6 gen github.com/artefactual-labs/enduro/internal/api/design -o
// internal/api

package server

import (
	"context"
	"net/http"

	batch "github.com/artefactual-labs/enduro/internal/api/gen/batch"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
	"goa.design/plugins/v3/cors"
)

// Server lists the batch service endpoint HTTP handlers.
type Server struct {
	Mounts []*MountPoint
	Submit http.Handler
	Status http.Handler
	Hints  http.Handler
	CORS   http.Handler
}

// ErrorNamer is an interface implemented by generated error structs that
// exposes the name of the error as defined in the design.
type ErrorNamer interface {
	ErrorName() string
}

// MountPoint holds information about the mounted endpoints.
type MountPoint struct {
	// Method is the name of the service method served by the mounted HTTP handler.
	Method string
	// Verb is the HTTP method used to match requests to the mounted handler.
	Verb string
	// Pattern is the HTTP request path pattern used to match requests to the
	// mounted handler.
	Pattern string
}

// New instantiates HTTP handlers for all the batch service endpoints using the
// provided encoder and decoder. The handlers are mounted on the given mux
// using the HTTP verb and path defined in the design. errhandler is called
// whenever a response fails to be encoded. formatter is used to format errors
// returned by the service methods prior to encoding. Both errhandler and
// formatter are optional and can be nil.
func New(
	e *batch.Endpoints,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(err error) goahttp.Statuser,
) *Server {
	return &Server{
		Mounts: []*MountPoint{
			{"Submit", "POST", "/batch"},
			{"Status", "GET", "/batch"},
			{"Hints", "GET", "/batch/hints"},
			{"CORS", "OPTIONS", "/batch"},
			{"CORS", "OPTIONS", "/batch/hints"},
		},
		Submit: NewSubmitHandler(e.Submit, mux, decoder, encoder, errhandler, formatter),
		Status: NewStatusHandler(e.Status, mux, decoder, encoder, errhandler, formatter),
		Hints:  NewHintsHandler(e.Hints, mux, decoder, encoder, errhandler, formatter),
		CORS:   NewCORSHandler(),
	}
}

// Service returns the name of the service served.
func (s *Server) Service() string { return "batch" }

// Use wraps the server handlers with the given middleware.
func (s *Server) Use(m func(http.Handler) http.Handler) {
	s.Submit = m(s.Submit)
	s.Status = m(s.Status)
	s.Hints = m(s.Hints)
	s.CORS = m(s.CORS)
}

// Mount configures the mux to serve the batch endpoints.
func Mount(mux goahttp.Muxer, h *Server) {
	MountSubmitHandler(mux, h.Submit)
	MountStatusHandler(mux, h.Status)
	MountHintsHandler(mux, h.Hints)
	MountCORSHandler(mux, h.CORS)
}

// Mount configures the mux to serve the batch endpoints.
func (s *Server) Mount(mux goahttp.Muxer) {
	Mount(mux, s)
}

// MountSubmitHandler configures the mux to serve the "batch" service "submit"
// endpoint.
func MountSubmitHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandleBatchOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/batch", f)
}

// NewSubmitHandler creates a HTTP handler which loads the HTTP request and
// calls the "batch" service "submit" endpoint.
func NewSubmitHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeSubmitRequest(mux, decoder)
		encodeResponse = EncodeSubmitResponse(encoder)
		encodeError    = EncodeSubmitError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "submit")
		ctx = context.WithValue(ctx, goa.ServiceKey, "batch")
		payload, err := decodeRequest(r)
		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				errhandler(ctx, w, err)
			}
			return
		}
		res, err := endpoint(ctx, payload)
		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				errhandler(ctx, w, err)
			}
			return
		}
		if err := encodeResponse(ctx, w, res); err != nil {
			errhandler(ctx, w, err)
		}
	})
}

// MountStatusHandler configures the mux to serve the "batch" service "status"
// endpoint.
func MountStatusHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandleBatchOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/batch", f)
}

// NewStatusHandler creates a HTTP handler which loads the HTTP request and
// calls the "batch" service "status" endpoint.
func NewStatusHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(err error) goahttp.Statuser,
) http.Handler {
	var (
		encodeResponse = EncodeStatusResponse(encoder)
		encodeError    = goahttp.ErrorEncoder(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "status")
		ctx = context.WithValue(ctx, goa.ServiceKey, "batch")
		var err error
		res, err := endpoint(ctx, nil)
		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				errhandler(ctx, w, err)
			}
			return
		}
		if err := encodeResponse(ctx, w, res); err != nil {
			errhandler(ctx, w, err)
		}
	})
}

// MountHintsHandler configures the mux to serve the "batch" service "hints"
// endpoint.
func MountHintsHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandleBatchOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/batch/hints", f)
}

// NewHintsHandler creates a HTTP handler which loads the HTTP request and
// calls the "batch" service "hints" endpoint.
func NewHintsHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(err error) goahttp.Statuser,
) http.Handler {
	var (
		encodeResponse = EncodeHintsResponse(encoder)
		encodeError    = goahttp.ErrorEncoder(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "hints")
		ctx = context.WithValue(ctx, goa.ServiceKey, "batch")
		var err error
		res, err := endpoint(ctx, nil)
		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				errhandler(ctx, w, err)
			}
			return
		}
		if err := encodeResponse(ctx, w, res); err != nil {
			errhandler(ctx, w, err)
		}
	})
}

// MountCORSHandler configures the mux to serve the CORS endpoints for the
// service batch.
func MountCORSHandler(mux goahttp.Muxer, h http.Handler) {
	h = HandleBatchOrigin(h)
	mux.Handle("OPTIONS", "/batch", h.ServeHTTP)
	mux.Handle("OPTIONS", "/batch/hints", h.ServeHTTP)
}

// NewCORSHandler creates a HTTP handler which returns a simple 200 response.
func NewCORSHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
}

// HandleBatchOrigin applies the CORS response headers corresponding to the
// origin for the service batch.
func HandleBatchOrigin(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			// Not a CORS request
			h.ServeHTTP(w, r)
			return
		}
		if cors.MatchOrigin(origin, "*") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			if acrm := r.Header.Get("Access-Control-Request-Method"); acrm != "" {
				// We are handling a preflight request
				w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PUT, DELETE, OPTIONS")
			}
			h.ServeHTTP(w, r)
			return
		}
		h.ServeHTTP(w, r)
		return
	})
}
