// Code generated by goa v3.15.2, DO NOT EDIT.
//
// package HTTP server
//
// Command:
// $ goa gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package server

import (
	"context"
	"net/http"
	"os"

	package_ "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
	otelhttp "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
	"goa.design/plugins/v3/cors"
)

// Server lists the package service endpoint HTTP handlers.
type Server struct {
	Mounts              []*MountPoint
	MonitorRequest      http.Handler
	Monitor             http.Handler
	List                http.Handler
	Show                http.Handler
	PreservationActions http.Handler
	Confirm             http.Handler
	Reject              http.Handler
	Move                http.Handler
	MoveStatus          http.Handler
	Upload              http.Handler
	CORS                http.Handler
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

// New instantiates HTTP handlers for all the package service endpoints using
// the provided encoder and decoder. The handlers are mounted on the given mux
// using the HTTP verb and path defined in the design. errhandler is called
// whenever a response fails to be encoded. formatter is used to format errors
// returned by the service methods prior to encoding. Both errhandler and
// formatter are optional and can be nil.
func New(
	e *package_.Endpoints,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
	upgrader goahttp.Upgrader,
	configurer *ConnConfigurer,
) *Server {
	if configurer == nil {
		configurer = &ConnConfigurer{}
	}
	return &Server{
		Mounts: []*MountPoint{
			{"MonitorRequest", "POST", "/package/monitor"},
			{"Monitor", "GET", "/package/monitor"},
			{"List", "GET", "/package"},
			{"Show", "GET", "/package/{id}"},
			{"PreservationActions", "GET", "/package/{id}/preservation-actions"},
			{"Confirm", "POST", "/package/{id}/confirm"},
			{"Reject", "POST", "/package/{id}/reject"},
			{"Move", "POST", "/package/{id}/move"},
			{"MoveStatus", "GET", "/package/{id}/move"},
			{"Upload", "POST", "/package/upload"},
			{"CORS", "OPTIONS", "/package/monitor"},
			{"CORS", "OPTIONS", "/package"},
			{"CORS", "OPTIONS", "/package/{id}"},
			{"CORS", "OPTIONS", "/package/{id}/preservation-actions"},
			{"CORS", "OPTIONS", "/package/{id}/confirm"},
			{"CORS", "OPTIONS", "/package/{id}/reject"},
			{"CORS", "OPTIONS", "/package/{id}/move"},
			{"CORS", "OPTIONS", "/package/upload"},
		},
		MonitorRequest:      NewMonitorRequestHandler(e.MonitorRequest, mux, decoder, encoder, errhandler, formatter),
		Monitor:             NewMonitorHandler(e.Monitor, mux, decoder, encoder, errhandler, formatter, upgrader, configurer.MonitorFn),
		List:                NewListHandler(e.List, mux, decoder, encoder, errhandler, formatter),
		Show:                NewShowHandler(e.Show, mux, decoder, encoder, errhandler, formatter),
		PreservationActions: NewPreservationActionsHandler(e.PreservationActions, mux, decoder, encoder, errhandler, formatter),
		Confirm:             NewConfirmHandler(e.Confirm, mux, decoder, encoder, errhandler, formatter),
		Reject:              NewRejectHandler(e.Reject, mux, decoder, encoder, errhandler, formatter),
		Move:                NewMoveHandler(e.Move, mux, decoder, encoder, errhandler, formatter),
		MoveStatus:          NewMoveStatusHandler(e.MoveStatus, mux, decoder, encoder, errhandler, formatter),
		Upload:              NewUploadHandler(e.Upload, mux, decoder, encoder, errhandler, formatter),
		CORS:                NewCORSHandler(),
	}
}

// Service returns the name of the service served.
func (s *Server) Service() string { return "package" }

// Use wraps the server handlers with the given middleware.
func (s *Server) Use(m func(http.Handler) http.Handler) {
	s.MonitorRequest = m(s.MonitorRequest)
	s.Monitor = m(s.Monitor)
	s.List = m(s.List)
	s.Show = m(s.Show)
	s.PreservationActions = m(s.PreservationActions)
	s.Confirm = m(s.Confirm)
	s.Reject = m(s.Reject)
	s.Move = m(s.Move)
	s.MoveStatus = m(s.MoveStatus)
	s.Upload = m(s.Upload)
	s.CORS = m(s.CORS)
}

// MethodNames returns the methods served.
func (s *Server) MethodNames() []string { return package_.MethodNames[:] }

// Mount configures the mux to serve the package endpoints.
func Mount(mux goahttp.Muxer, h *Server) {
	MountMonitorRequestHandler(mux, h.MonitorRequest)
	MountMonitorHandler(mux, h.Monitor)
	MountListHandler(mux, h.List)
	MountShowHandler(mux, h.Show)
	MountPreservationActionsHandler(mux, h.PreservationActions)
	MountConfirmHandler(mux, h.Confirm)
	MountRejectHandler(mux, h.Reject)
	MountMoveHandler(mux, h.Move)
	MountMoveStatusHandler(mux, h.MoveStatus)
	MountUploadHandler(mux, h.Upload)
	MountCORSHandler(mux, h.CORS)
}

// Mount configures the mux to serve the package endpoints.
func (s *Server) Mount(mux goahttp.Muxer) {
	Mount(mux, s)
}

// MountMonitorRequestHandler configures the mux to serve the "package" service
// "monitor_request" endpoint.
func MountMonitorRequestHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandlePackageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/package/monitor", otelhttp.WithRouteTag("/package/monitor", f).ServeHTTP)
}

// NewMonitorRequestHandler creates a HTTP handler which loads the HTTP request
// and calls the "package" service "monitor_request" endpoint.
func NewMonitorRequestHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeMonitorRequestRequest(mux, decoder)
		encodeResponse = EncodeMonitorRequestResponse(encoder)
		encodeError    = EncodeMonitorRequestError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "monitor_request")
		ctx = context.WithValue(ctx, goa.ServiceKey, "package")
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

// MountMonitorHandler configures the mux to serve the "package" service
// "monitor" endpoint.
func MountMonitorHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandlePackageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/package/monitor", otelhttp.WithRouteTag("/package/monitor", f).ServeHTTP)
}

// NewMonitorHandler creates a HTTP handler which loads the HTTP request and
// calls the "package" service "monitor" endpoint.
func NewMonitorHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
	upgrader goahttp.Upgrader,
	configurer goahttp.ConnConfigureFunc,
) http.Handler {
	var (
		decodeRequest = DecodeMonitorRequest(mux, decoder)
		encodeError   = EncodeMonitorError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "monitor")
		ctx = context.WithValue(ctx, goa.ServiceKey, "package")
		payload, err := decodeRequest(r)
		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				errhandler(ctx, w, err)
			}
			return
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		v := &package_.MonitorEndpointInput{
			Stream: &MonitorServerStream{
				upgrader:   upgrader,
				configurer: configurer,
				cancel:     cancel,
				w:          w,
				r:          r,
			},
			Payload: payload.(*package_.MonitorPayload),
		}
		_, err = endpoint(ctx, v)
		if err != nil {
			if v.Stream.(*MonitorServerStream).conn != nil {
				// Response writer has been hijacked, do not encode the error
				errhandler(ctx, w, err)
				return
			}
			if err := encodeError(ctx, w, err); err != nil {
				errhandler(ctx, w, err)
			}
			return
		}
	})
}

// MountListHandler configures the mux to serve the "package" service "list"
// endpoint.
func MountListHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandlePackageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/package", otelhttp.WithRouteTag("/package", f).ServeHTTP)
}

// NewListHandler creates a HTTP handler which loads the HTTP request and calls
// the "package" service "list" endpoint.
func NewListHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeListRequest(mux, decoder)
		encodeResponse = EncodeListResponse(encoder)
		encodeError    = EncodeListError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "list")
		ctx = context.WithValue(ctx, goa.ServiceKey, "package")
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

// MountShowHandler configures the mux to serve the "package" service "show"
// endpoint.
func MountShowHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandlePackageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/package/{id}", otelhttp.WithRouteTag("/package/{id}", f).ServeHTTP)
}

// NewShowHandler creates a HTTP handler which loads the HTTP request and calls
// the "package" service "show" endpoint.
func NewShowHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeShowRequest(mux, decoder)
		encodeResponse = EncodeShowResponse(encoder)
		encodeError    = EncodeShowError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "show")
		ctx = context.WithValue(ctx, goa.ServiceKey, "package")
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

// MountPreservationActionsHandler configures the mux to serve the "package"
// service "preservation_actions" endpoint.
func MountPreservationActionsHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandlePackageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/package/{id}/preservation-actions", otelhttp.WithRouteTag("/package/{id}/preservation-actions", f).ServeHTTP)
}

// NewPreservationActionsHandler creates a HTTP handler which loads the HTTP
// request and calls the "package" service "preservation_actions" endpoint.
func NewPreservationActionsHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodePreservationActionsRequest(mux, decoder)
		encodeResponse = EncodePreservationActionsResponse(encoder)
		encodeError    = EncodePreservationActionsError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "preservation_actions")
		ctx = context.WithValue(ctx, goa.ServiceKey, "package")
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

// MountConfirmHandler configures the mux to serve the "package" service
// "confirm" endpoint.
func MountConfirmHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandlePackageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/package/{id}/confirm", otelhttp.WithRouteTag("/package/{id}/confirm", f).ServeHTTP)
}

// NewConfirmHandler creates a HTTP handler which loads the HTTP request and
// calls the "package" service "confirm" endpoint.
func NewConfirmHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeConfirmRequest(mux, decoder)
		encodeResponse = EncodeConfirmResponse(encoder)
		encodeError    = EncodeConfirmError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "confirm")
		ctx = context.WithValue(ctx, goa.ServiceKey, "package")
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

// MountRejectHandler configures the mux to serve the "package" service
// "reject" endpoint.
func MountRejectHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandlePackageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/package/{id}/reject", otelhttp.WithRouteTag("/package/{id}/reject", f).ServeHTTP)
}

// NewRejectHandler creates a HTTP handler which loads the HTTP request and
// calls the "package" service "reject" endpoint.
func NewRejectHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeRejectRequest(mux, decoder)
		encodeResponse = EncodeRejectResponse(encoder)
		encodeError    = EncodeRejectError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "reject")
		ctx = context.WithValue(ctx, goa.ServiceKey, "package")
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

// MountMoveHandler configures the mux to serve the "package" service "move"
// endpoint.
func MountMoveHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandlePackageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/package/{id}/move", otelhttp.WithRouteTag("/package/{id}/move", f).ServeHTTP)
}

// NewMoveHandler creates a HTTP handler which loads the HTTP request and calls
// the "package" service "move" endpoint.
func NewMoveHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeMoveRequest(mux, decoder)
		encodeResponse = EncodeMoveResponse(encoder)
		encodeError    = EncodeMoveError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "move")
		ctx = context.WithValue(ctx, goa.ServiceKey, "package")
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

// MountMoveStatusHandler configures the mux to serve the "package" service
// "move_status" endpoint.
func MountMoveStatusHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandlePackageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/package/{id}/move", otelhttp.WithRouteTag("/package/{id}/move", f).ServeHTTP)
}

// NewMoveStatusHandler creates a HTTP handler which loads the HTTP request and
// calls the "package" service "move_status" endpoint.
func NewMoveStatusHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeMoveStatusRequest(mux, decoder)
		encodeResponse = EncodeMoveStatusResponse(encoder)
		encodeError    = EncodeMoveStatusError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "move_status")
		ctx = context.WithValue(ctx, goa.ServiceKey, "package")
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

// MountUploadHandler configures the mux to serve the "package" service
// "upload" endpoint.
func MountUploadHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandlePackageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/package/upload", otelhttp.WithRouteTag("/package/upload", f).ServeHTTP)
}

// NewUploadHandler creates a HTTP handler which loads the HTTP request and
// calls the "package" service "upload" endpoint.
func NewUploadHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeUploadRequest(mux, decoder)
		encodeResponse = EncodeUploadResponse(encoder)
		encodeError    = EncodeUploadError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "upload")
		ctx = context.WithValue(ctx, goa.ServiceKey, "package")
		payload, err := decodeRequest(r)
		if err != nil {
			if err := encodeError(ctx, w, err); err != nil {
				errhandler(ctx, w, err)
			}
			return
		}
		data := &package_.UploadRequestData{Payload: payload.(*package_.UploadPayload), Body: r.Body}
		res, err := endpoint(ctx, data)
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
// service package.
func MountCORSHandler(mux goahttp.Muxer, h http.Handler) {
	h = HandlePackageOrigin(h)
	mux.Handle("OPTIONS", "/package/monitor", h.ServeHTTP)
	mux.Handle("OPTIONS", "/package", h.ServeHTTP)
	mux.Handle("OPTIONS", "/package/{id}", h.ServeHTTP)
	mux.Handle("OPTIONS", "/package/{id}/preservation-actions", h.ServeHTTP)
	mux.Handle("OPTIONS", "/package/{id}/confirm", h.ServeHTTP)
	mux.Handle("OPTIONS", "/package/{id}/reject", h.ServeHTTP)
	mux.Handle("OPTIONS", "/package/{id}/move", h.ServeHTTP)
	mux.Handle("OPTIONS", "/package/upload", h.ServeHTTP)
}

// NewCORSHandler creates a HTTP handler which returns a simple 204 response.
func NewCORSHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
}

// HandlePackageOrigin applies the CORS response headers corresponding to the
// origin for the service package.
func HandlePackageOrigin(h http.Handler) http.Handler {
	originStr0, present := os.LookupEnv("ENDURO_API_CORS_ORIGIN")
	if !present {
		panic("CORS origin environment variable \"ENDURO_API_CORS_ORIGIN\" not set!")
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			// Not a CORS request
			h.ServeHTTP(w, r)
			return
		}
		if cors.MatchOrigin(origin, originStr0) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			if acrm := r.Header.Get("Access-Control-Request-Method"); acrm != "" {
				// We are handling a preflight request
				w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.WriteHeader(204)
				return
			}
			h.ServeHTTP(w, r)
			return
		}
		h.ServeHTTP(w, r)
		return
	})
}
