// Code generated by goa v3.8.4, DO NOT EDIT.
//
// storage client HTTP transport
//
// Command:
// $ goa-v3.8.4 gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package client

import (
	"context"
	"net/http"

	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

// Client lists the storage service endpoint HTTP clients.
type Client struct {
	// Submit Doer is the HTTP client used to make requests to the submit endpoint.
	SubmitDoer goahttp.Doer

	// Update Doer is the HTTP client used to make requests to the update endpoint.
	UpdateDoer goahttp.Doer

	// Download Doer is the HTTP client used to make requests to the download
	// endpoint.
	DownloadDoer goahttp.Doer

	// Locations Doer is the HTTP client used to make requests to the locations
	// endpoint.
	LocationsDoer goahttp.Doer

	// AddLocation Doer is the HTTP client used to make requests to the
	// add_location endpoint.
	AddLocationDoer goahttp.Doer

	// Move Doer is the HTTP client used to make requests to the move endpoint.
	MoveDoer goahttp.Doer

	// MoveStatus Doer is the HTTP client used to make requests to the move_status
	// endpoint.
	MoveStatusDoer goahttp.Doer

	// Reject Doer is the HTTP client used to make requests to the reject endpoint.
	RejectDoer goahttp.Doer

	// Show Doer is the HTTP client used to make requests to the show endpoint.
	ShowDoer goahttp.Doer

	// ShowLocation Doer is the HTTP client used to make requests to the
	// show_location endpoint.
	ShowLocationDoer goahttp.Doer

	// LocationPackages Doer is the HTTP client used to make requests to the
	// location_packages endpoint.
	LocationPackagesDoer goahttp.Doer

	// CORS Doer is the HTTP client used to make requests to the  endpoint.
	CORSDoer goahttp.Doer

	// RestoreResponseBody controls whether the response bodies are reset after
	// decoding so they can be read again.
	RestoreResponseBody bool

	scheme  string
	host    string
	encoder func(*http.Request) goahttp.Encoder
	decoder func(*http.Response) goahttp.Decoder
}

// NewClient instantiates HTTP clients for all the storage service servers.
func NewClient(
	scheme string,
	host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restoreBody bool,
) *Client {
	return &Client{
		SubmitDoer:           doer,
		UpdateDoer:           doer,
		DownloadDoer:         doer,
		LocationsDoer:        doer,
		AddLocationDoer:      doer,
		MoveDoer:             doer,
		MoveStatusDoer:       doer,
		RejectDoer:           doer,
		ShowDoer:             doer,
		ShowLocationDoer:     doer,
		LocationPackagesDoer: doer,
		CORSDoer:             doer,
		RestoreResponseBody:  restoreBody,
		scheme:               scheme,
		host:                 host,
		decoder:              dec,
		encoder:              enc,
	}
}

// Submit returns an endpoint that makes HTTP requests to the storage service
// submit server.
func (c *Client) Submit() goa.Endpoint {
	var (
		encodeRequest  = EncodeSubmitRequest(c.encoder)
		decodeResponse = DecodeSubmitResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.BuildSubmitRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		err = encodeRequest(req, v)
		if err != nil {
			return nil, err
		}
		resp, err := c.SubmitDoer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("storage", "submit", err)
		}
		return decodeResponse(resp)
	}
}

// Update returns an endpoint that makes HTTP requests to the storage service
// update server.
func (c *Client) Update() goa.Endpoint {
	var (
		decodeResponse = DecodeUpdateResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.BuildUpdateRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		resp, err := c.UpdateDoer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("storage", "update", err)
		}
		return decodeResponse(resp)
	}
}

// Download returns an endpoint that makes HTTP requests to the storage service
// download server.
func (c *Client) Download() goa.Endpoint {
	var (
		decodeResponse = DecodeDownloadResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.BuildDownloadRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		resp, err := c.DownloadDoer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("storage", "download", err)
		}
		return decodeResponse(resp)
	}
}

// Locations returns an endpoint that makes HTTP requests to the storage
// service locations server.
func (c *Client) Locations() goa.Endpoint {
	var (
		decodeResponse = DecodeLocationsResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.BuildLocationsRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		resp, err := c.LocationsDoer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("storage", "locations", err)
		}
		return decodeResponse(resp)
	}
}

// AddLocation returns an endpoint that makes HTTP requests to the storage
// service add_location server.
func (c *Client) AddLocation() goa.Endpoint {
	var (
		encodeRequest  = EncodeAddLocationRequest(c.encoder)
		decodeResponse = DecodeAddLocationResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.BuildAddLocationRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		err = encodeRequest(req, v)
		if err != nil {
			return nil, err
		}
		resp, err := c.AddLocationDoer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("storage", "add_location", err)
		}
		return decodeResponse(resp)
	}
}

// Move returns an endpoint that makes HTTP requests to the storage service
// move server.
func (c *Client) Move() goa.Endpoint {
	var (
		encodeRequest  = EncodeMoveRequest(c.encoder)
		decodeResponse = DecodeMoveResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.BuildMoveRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		err = encodeRequest(req, v)
		if err != nil {
			return nil, err
		}
		resp, err := c.MoveDoer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("storage", "move", err)
		}
		return decodeResponse(resp)
	}
}

// MoveStatus returns an endpoint that makes HTTP requests to the storage
// service move_status server.
func (c *Client) MoveStatus() goa.Endpoint {
	var (
		decodeResponse = DecodeMoveStatusResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.BuildMoveStatusRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		resp, err := c.MoveStatusDoer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("storage", "move_status", err)
		}
		return decodeResponse(resp)
	}
}

// Reject returns an endpoint that makes HTTP requests to the storage service
// reject server.
func (c *Client) Reject() goa.Endpoint {
	var (
		decodeResponse = DecodeRejectResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.BuildRejectRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		resp, err := c.RejectDoer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("storage", "reject", err)
		}
		return decodeResponse(resp)
	}
}

// Show returns an endpoint that makes HTTP requests to the storage service
// show server.
func (c *Client) Show() goa.Endpoint {
	var (
		decodeResponse = DecodeShowResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.BuildShowRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		resp, err := c.ShowDoer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("storage", "show", err)
		}
		return decodeResponse(resp)
	}
}

// ShowLocation returns an endpoint that makes HTTP requests to the storage
// service show_location server.
func (c *Client) ShowLocation() goa.Endpoint {
	var (
		decodeResponse = DecodeShowLocationResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.BuildShowLocationRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		resp, err := c.ShowLocationDoer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("storage", "show_location", err)
		}
		return decodeResponse(resp)
	}
}

// LocationPackages returns an endpoint that makes HTTP requests to the storage
// service location_packages server.
func (c *Client) LocationPackages() goa.Endpoint {
	var (
		decodeResponse = DecodeLocationPackagesResponse(c.decoder, c.RestoreResponseBody)
	)
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req, err := c.BuildLocationPackagesRequest(ctx, v)
		if err != nil {
			return nil, err
		}
		resp, err := c.LocationPackagesDoer.Do(req)
		if err != nil {
			return nil, goahttp.ErrRequestError("storage", "location_packages", err)
		}
		return decodeResponse(resp)
	}
}
