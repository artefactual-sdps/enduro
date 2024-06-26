// Code generated by goa v3.15.2, DO NOT EDIT.
//
// storage HTTP server
//
// Command:
// $ goa gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package server

import (
	"context"
	"net/http"
	"os"

	storage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	otelhttp "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
	"goa.design/plugins/v3/cors"
)

// Server lists the storage service endpoint HTTP handlers.
type Server struct {
	Mounts           []*MountPoint
	Create           http.Handler
	Submit           http.Handler
	Update           http.Handler
	Download         http.Handler
	Move             http.Handler
	MoveStatus       http.Handler
	Reject           http.Handler
	Show             http.Handler
	Locations        http.Handler
	AddLocation      http.Handler
	ShowLocation     http.Handler
	LocationPackages http.Handler
	CORS             http.Handler
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

// New instantiates HTTP handlers for all the storage service endpoints using
// the provided encoder and decoder. The handlers are mounted on the given mux
// using the HTTP verb and path defined in the design. errhandler is called
// whenever a response fails to be encoded. formatter is used to format errors
// returned by the service methods prior to encoding. Both errhandler and
// formatter are optional and can be nil.
func New(
	e *storage.Endpoints,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) *Server {
	return &Server{
		Mounts: []*MountPoint{
			{"Create", "POST", "/storage/package"},
			{"Submit", "POST", "/storage/package/{aip_id}/submit"},
			{"Update", "POST", "/storage/package/{aip_id}/update"},
			{"Download", "GET", "/storage/package/{aip_id}/download"},
			{"Move", "POST", "/storage/package/{aip_id}/store"},
			{"MoveStatus", "GET", "/storage/package/{aip_id}/store"},
			{"Reject", "POST", "/storage/package/{aip_id}/reject"},
			{"Show", "GET", "/storage/package/{aip_id}"},
			{"Locations", "GET", "/storage/location"},
			{"AddLocation", "POST", "/storage/location"},
			{"ShowLocation", "GET", "/storage/location/{uuid}"},
			{"LocationPackages", "GET", "/storage/location/{uuid}/packages"},
			{"CORS", "OPTIONS", "/storage/package"},
			{"CORS", "OPTIONS", "/storage/package/{aip_id}/submit"},
			{"CORS", "OPTIONS", "/storage/package/{aip_id}/update"},
			{"CORS", "OPTIONS", "/storage/package/{aip_id}/download"},
			{"CORS", "OPTIONS", "/storage/package/{aip_id}/store"},
			{"CORS", "OPTIONS", "/storage/package/{aip_id}/reject"},
			{"CORS", "OPTIONS", "/storage/package/{aip_id}"},
			{"CORS", "OPTIONS", "/storage/location"},
			{"CORS", "OPTIONS", "/storage/location/{uuid}"},
			{"CORS", "OPTIONS", "/storage/location/{uuid}/packages"},
		},
		Create:           NewCreateHandler(e.Create, mux, decoder, encoder, errhandler, formatter),
		Submit:           NewSubmitHandler(e.Submit, mux, decoder, encoder, errhandler, formatter),
		Update:           NewUpdateHandler(e.Update, mux, decoder, encoder, errhandler, formatter),
		Download:         NewDownloadHandler(e.Download, mux, decoder, encoder, errhandler, formatter),
		Move:             NewMoveHandler(e.Move, mux, decoder, encoder, errhandler, formatter),
		MoveStatus:       NewMoveStatusHandler(e.MoveStatus, mux, decoder, encoder, errhandler, formatter),
		Reject:           NewRejectHandler(e.Reject, mux, decoder, encoder, errhandler, formatter),
		Show:             NewShowHandler(e.Show, mux, decoder, encoder, errhandler, formatter),
		Locations:        NewLocationsHandler(e.Locations, mux, decoder, encoder, errhandler, formatter),
		AddLocation:      NewAddLocationHandler(e.AddLocation, mux, decoder, encoder, errhandler, formatter),
		ShowLocation:     NewShowLocationHandler(e.ShowLocation, mux, decoder, encoder, errhandler, formatter),
		LocationPackages: NewLocationPackagesHandler(e.LocationPackages, mux, decoder, encoder, errhandler, formatter),
		CORS:             NewCORSHandler(),
	}
}

// Service returns the name of the service served.
func (s *Server) Service() string { return "storage" }

// Use wraps the server handlers with the given middleware.
func (s *Server) Use(m func(http.Handler) http.Handler) {
	s.Create = m(s.Create)
	s.Submit = m(s.Submit)
	s.Update = m(s.Update)
	s.Download = m(s.Download)
	s.Move = m(s.Move)
	s.MoveStatus = m(s.MoveStatus)
	s.Reject = m(s.Reject)
	s.Show = m(s.Show)
	s.Locations = m(s.Locations)
	s.AddLocation = m(s.AddLocation)
	s.ShowLocation = m(s.ShowLocation)
	s.LocationPackages = m(s.LocationPackages)
	s.CORS = m(s.CORS)
}

// MethodNames returns the methods served.
func (s *Server) MethodNames() []string { return storage.MethodNames[:] }

// Mount configures the mux to serve the storage endpoints.
func Mount(mux goahttp.Muxer, h *Server) {
	MountCreateHandler(mux, h.Create)
	MountSubmitHandler(mux, h.Submit)
	MountUpdateHandler(mux, h.Update)
	MountDownloadHandler(mux, h.Download)
	MountMoveHandler(mux, h.Move)
	MountMoveStatusHandler(mux, h.MoveStatus)
	MountRejectHandler(mux, h.Reject)
	MountShowHandler(mux, h.Show)
	MountLocationsHandler(mux, h.Locations)
	MountAddLocationHandler(mux, h.AddLocation)
	MountShowLocationHandler(mux, h.ShowLocation)
	MountLocationPackagesHandler(mux, h.LocationPackages)
	MountCORSHandler(mux, h.CORS)
}

// Mount configures the mux to serve the storage endpoints.
func (s *Server) Mount(mux goahttp.Muxer) {
	Mount(mux, s)
}

// MountCreateHandler configures the mux to serve the "storage" service
// "create" endpoint.
func MountCreateHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandleStorageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/storage/package", otelhttp.WithRouteTag("/storage/package", f).ServeHTTP)
}

// NewCreateHandler creates a HTTP handler which loads the HTTP request and
// calls the "storage" service "create" endpoint.
func NewCreateHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeCreateRequest(mux, decoder)
		encodeResponse = EncodeCreateResponse(encoder)
		encodeError    = EncodeCreateError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "create")
		ctx = context.WithValue(ctx, goa.ServiceKey, "storage")
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

// MountSubmitHandler configures the mux to serve the "storage" service
// "submit" endpoint.
func MountSubmitHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandleStorageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/storage/package/{aip_id}/submit", otelhttp.WithRouteTag("/storage/package/{aip_id}/submit", f).ServeHTTP)
}

// NewSubmitHandler creates a HTTP handler which loads the HTTP request and
// calls the "storage" service "submit" endpoint.
func NewSubmitHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeSubmitRequest(mux, decoder)
		encodeResponse = EncodeSubmitResponse(encoder)
		encodeError    = EncodeSubmitError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "submit")
		ctx = context.WithValue(ctx, goa.ServiceKey, "storage")
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

// MountUpdateHandler configures the mux to serve the "storage" service
// "update" endpoint.
func MountUpdateHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandleStorageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/storage/package/{aip_id}/update", otelhttp.WithRouteTag("/storage/package/{aip_id}/update", f).ServeHTTP)
}

// NewUpdateHandler creates a HTTP handler which loads the HTTP request and
// calls the "storage" service "update" endpoint.
func NewUpdateHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeUpdateRequest(mux, decoder)
		encodeResponse = EncodeUpdateResponse(encoder)
		encodeError    = EncodeUpdateError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "update")
		ctx = context.WithValue(ctx, goa.ServiceKey, "storage")
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

// MountDownloadHandler configures the mux to serve the "storage" service
// "download" endpoint.
func MountDownloadHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandleStorageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/storage/package/{aip_id}/download", otelhttp.WithRouteTag("/storage/package/{aip_id}/download", f).ServeHTTP)
}

// NewDownloadHandler creates a HTTP handler which loads the HTTP request and
// calls the "storage" service "download" endpoint.
func NewDownloadHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeDownloadRequest(mux, decoder)
		encodeResponse = EncodeDownloadResponse(encoder)
		encodeError    = EncodeDownloadError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "download")
		ctx = context.WithValue(ctx, goa.ServiceKey, "storage")
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

// MountMoveHandler configures the mux to serve the "storage" service "move"
// endpoint.
func MountMoveHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandleStorageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/storage/package/{aip_id}/store", otelhttp.WithRouteTag("/storage/package/{aip_id}/store", f).ServeHTTP)
}

// NewMoveHandler creates a HTTP handler which loads the HTTP request and calls
// the "storage" service "move" endpoint.
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
		ctx = context.WithValue(ctx, goa.ServiceKey, "storage")
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

// MountMoveStatusHandler configures the mux to serve the "storage" service
// "move_status" endpoint.
func MountMoveStatusHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandleStorageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/storage/package/{aip_id}/store", otelhttp.WithRouteTag("/storage/package/{aip_id}/store", f).ServeHTTP)
}

// NewMoveStatusHandler creates a HTTP handler which loads the HTTP request and
// calls the "storage" service "move_status" endpoint.
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
		ctx = context.WithValue(ctx, goa.ServiceKey, "storage")
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

// MountRejectHandler configures the mux to serve the "storage" service
// "reject" endpoint.
func MountRejectHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandleStorageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/storage/package/{aip_id}/reject", otelhttp.WithRouteTag("/storage/package/{aip_id}/reject", f).ServeHTTP)
}

// NewRejectHandler creates a HTTP handler which loads the HTTP request and
// calls the "storage" service "reject" endpoint.
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
		ctx = context.WithValue(ctx, goa.ServiceKey, "storage")
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

// MountShowHandler configures the mux to serve the "storage" service "show"
// endpoint.
func MountShowHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandleStorageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/storage/package/{aip_id}", otelhttp.WithRouteTag("/storage/package/{aip_id}", f).ServeHTTP)
}

// NewShowHandler creates a HTTP handler which loads the HTTP request and calls
// the "storage" service "show" endpoint.
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
		ctx = context.WithValue(ctx, goa.ServiceKey, "storage")
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

// MountLocationsHandler configures the mux to serve the "storage" service
// "locations" endpoint.
func MountLocationsHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandleStorageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/storage/location", otelhttp.WithRouteTag("/storage/location", f).ServeHTTP)
}

// NewLocationsHandler creates a HTTP handler which loads the HTTP request and
// calls the "storage" service "locations" endpoint.
func NewLocationsHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeLocationsRequest(mux, decoder)
		encodeResponse = EncodeLocationsResponse(encoder)
		encodeError    = EncodeLocationsError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "locations")
		ctx = context.WithValue(ctx, goa.ServiceKey, "storage")
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

// MountAddLocationHandler configures the mux to serve the "storage" service
// "add_location" endpoint.
func MountAddLocationHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandleStorageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("POST", "/storage/location", otelhttp.WithRouteTag("/storage/location", f).ServeHTTP)
}

// NewAddLocationHandler creates a HTTP handler which loads the HTTP request
// and calls the "storage" service "add_location" endpoint.
func NewAddLocationHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeAddLocationRequest(mux, decoder)
		encodeResponse = EncodeAddLocationResponse(encoder)
		encodeError    = EncodeAddLocationError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "add_location")
		ctx = context.WithValue(ctx, goa.ServiceKey, "storage")
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

// MountShowLocationHandler configures the mux to serve the "storage" service
// "show_location" endpoint.
func MountShowLocationHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandleStorageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/storage/location/{uuid}", otelhttp.WithRouteTag("/storage/location/{uuid}", f).ServeHTTP)
}

// NewShowLocationHandler creates a HTTP handler which loads the HTTP request
// and calls the "storage" service "show_location" endpoint.
func NewShowLocationHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeShowLocationRequest(mux, decoder)
		encodeResponse = EncodeShowLocationResponse(encoder)
		encodeError    = EncodeShowLocationError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "show_location")
		ctx = context.WithValue(ctx, goa.ServiceKey, "storage")
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

// MountLocationPackagesHandler configures the mux to serve the "storage"
// service "location_packages" endpoint.
func MountLocationPackagesHandler(mux goahttp.Muxer, h http.Handler) {
	f, ok := HandleStorageOrigin(h).(http.HandlerFunc)
	if !ok {
		f = func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle("GET", "/storage/location/{uuid}/packages", otelhttp.WithRouteTag("/storage/location/{uuid}/packages", f).ServeHTTP)
}

// NewLocationPackagesHandler creates a HTTP handler which loads the HTTP
// request and calls the "storage" service "location_packages" endpoint.
func NewLocationPackagesHandler(
	endpoint goa.Endpoint,
	mux goahttp.Muxer,
	decoder func(*http.Request) goahttp.Decoder,
	encoder func(context.Context, http.ResponseWriter) goahttp.Encoder,
	errhandler func(context.Context, http.ResponseWriter, error),
	formatter func(ctx context.Context, err error) goahttp.Statuser,
) http.Handler {
	var (
		decodeRequest  = DecodeLocationPackagesRequest(mux, decoder)
		encodeResponse = EncodeLocationPackagesResponse(encoder)
		encodeError    = EncodeLocationPackagesError(encoder, formatter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), goahttp.AcceptTypeKey, r.Header.Get("Accept"))
		ctx = context.WithValue(ctx, goa.MethodKey, "location_packages")
		ctx = context.WithValue(ctx, goa.ServiceKey, "storage")
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

// MountCORSHandler configures the mux to serve the CORS endpoints for the
// service storage.
func MountCORSHandler(mux goahttp.Muxer, h http.Handler) {
	h = HandleStorageOrigin(h)
	mux.Handle("OPTIONS", "/storage/package", h.ServeHTTP)
	mux.Handle("OPTIONS", "/storage/package/{aip_id}/submit", h.ServeHTTP)
	mux.Handle("OPTIONS", "/storage/package/{aip_id}/update", h.ServeHTTP)
	mux.Handle("OPTIONS", "/storage/package/{aip_id}/download", h.ServeHTTP)
	mux.Handle("OPTIONS", "/storage/package/{aip_id}/store", h.ServeHTTP)
	mux.Handle("OPTIONS", "/storage/package/{aip_id}/reject", h.ServeHTTP)
	mux.Handle("OPTIONS", "/storage/package/{aip_id}", h.ServeHTTP)
	mux.Handle("OPTIONS", "/storage/location", h.ServeHTTP)
	mux.Handle("OPTIONS", "/storage/location/{uuid}", h.ServeHTTP)
	mux.Handle("OPTIONS", "/storage/location/{uuid}/packages", h.ServeHTTP)
}

// NewCORSHandler creates a HTTP handler which returns a simple 204 response.
func NewCORSHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
}

// HandleStorageOrigin applies the CORS response headers corresponding to the
// origin for the service storage.
func HandleStorageOrigin(h http.Handler) http.Handler {
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
