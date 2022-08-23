// Code generated by goa v3.8.3, DO NOT EDIT.
//
// storage HTTP client encoders and decoders
//
// Command:
// $ goa-v3.8.3 gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	storage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	storageviews "github.com/artefactual-sdps/enduro/internal/api/gen/storage/views"
	"github.com/google/uuid"
	goahttp "goa.design/goa/v3/http"
)

// BuildSubmitRequest instantiates a HTTP request object with method and path
// set to call the "storage" service "submit" endpoint
func (c *Client) BuildSubmitRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	var (
		aipID string
	)
	{
		p, ok := v.(*storage.SubmitPayload)
		if !ok {
			return nil, goahttp.ErrInvalidType("storage", "submit", "*storage.SubmitPayload", v)
		}
		aipID = p.AipID
	}
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: SubmitStoragePath(aipID)}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("storage", "submit", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// EncodeSubmitRequest returns an encoder for requests sent to the storage
// submit server.
func EncodeSubmitRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*storage.SubmitPayload)
		if !ok {
			return goahttp.ErrInvalidType("storage", "submit", "*storage.SubmitPayload", v)
		}
		body := NewSubmitRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("storage", "submit", err)
		}
		return nil
	}
}

// DecodeSubmitResponse returns a decoder for responses returned by the storage
// submit endpoint. restoreBody controls whether the response body should be
// restored after having been read.
// DecodeSubmitResponse may return the following errors:
//   - "not_available" (type *goa.ServiceError): http.StatusConflict
//   - "not_valid" (type *goa.ServiceError): http.StatusBadRequest
//   - error: internal error
func DecodeSubmitResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = io.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusAccepted:
			var (
				body SubmitResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "submit", err)
			}
			err = ValidateSubmitResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "submit", err)
			}
			res := NewSubmitResultAccepted(&body)
			return res, nil
		case http.StatusConflict:
			var (
				body SubmitNotAvailableResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "submit", err)
			}
			err = ValidateSubmitNotAvailableResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "submit", err)
			}
			return nil, NewSubmitNotAvailable(&body)
		case http.StatusBadRequest:
			var (
				body SubmitNotValidResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "submit", err)
			}
			err = ValidateSubmitNotValidResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "submit", err)
			}
			return nil, NewSubmitNotValid(&body)
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("storage", "submit", resp.StatusCode, string(body))
		}
	}
}

// BuildUpdateRequest instantiates a HTTP request object with method and path
// set to call the "storage" service "update" endpoint
func (c *Client) BuildUpdateRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	var (
		aipID string
	)
	{
		p, ok := v.(*storage.UpdatePayload)
		if !ok {
			return nil, goahttp.ErrInvalidType("storage", "update", "*storage.UpdatePayload", v)
		}
		aipID = p.AipID
	}
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: UpdateStoragePath(aipID)}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("storage", "update", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// DecodeUpdateResponse returns a decoder for responses returned by the storage
// update endpoint. restoreBody controls whether the response body should be
// restored after having been read.
// DecodeUpdateResponse may return the following errors:
//   - "not_available" (type *goa.ServiceError): http.StatusConflict
//   - "not_valid" (type *goa.ServiceError): http.StatusBadRequest
//   - error: internal error
func DecodeUpdateResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = io.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusAccepted:
			return nil, nil
		case http.StatusConflict:
			var (
				body UpdateNotAvailableResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "update", err)
			}
			err = ValidateUpdateNotAvailableResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "update", err)
			}
			return nil, NewUpdateNotAvailable(&body)
		case http.StatusBadRequest:
			var (
				body UpdateNotValidResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "update", err)
			}
			err = ValidateUpdateNotValidResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "update", err)
			}
			return nil, NewUpdateNotValid(&body)
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("storage", "update", resp.StatusCode, string(body))
		}
	}
}

// BuildDownloadRequest instantiates a HTTP request object with method and path
// set to call the "storage" service "download" endpoint
func (c *Client) BuildDownloadRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	var (
		aipID string
	)
	{
		p, ok := v.(*storage.DownloadPayload)
		if !ok {
			return nil, goahttp.ErrInvalidType("storage", "download", "*storage.DownloadPayload", v)
		}
		aipID = p.AipID
	}
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: DownloadStoragePath(aipID)}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("storage", "download", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// DecodeDownloadResponse returns a decoder for responses returned by the
// storage download endpoint. restoreBody controls whether the response body
// should be restored after having been read.
// DecodeDownloadResponse may return the following errors:
//   - "not_found" (type *storage.StoragePackageNotfound): http.StatusNotFound
//   - error: internal error
func DecodeDownloadResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = io.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusOK:
			var (
				body []byte
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "download", err)
			}
			return body, nil
		case http.StatusNotFound:
			var (
				body DownloadNotFoundResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "download", err)
			}
			err = ValidateDownloadNotFoundResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "download", err)
			}
			return nil, NewDownloadNotFound(&body)
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("storage", "download", resp.StatusCode, string(body))
		}
	}
}

// BuildLocationsRequest instantiates a HTTP request object with method and
// path set to call the "storage" service "locations" endpoint
func (c *Client) BuildLocationsRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: LocationsStoragePath()}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("storage", "locations", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// DecodeLocationsResponse returns a decoder for responses returned by the
// storage locations endpoint. restoreBody controls whether the response body
// should be restored after having been read.
func DecodeLocationsResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = io.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusOK:
			var (
				body LocationsResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "locations", err)
			}
			p := NewLocationsStoredLocationCollectionOK(body)
			view := "default"
			vres := storageviews.StoredLocationCollection{Projected: p, View: view}
			if err = storageviews.ValidateStoredLocationCollection(vres); err != nil {
				return nil, goahttp.ErrValidationError("storage", "locations", err)
			}
			res := storage.NewStoredLocationCollection(vres)
			return res, nil
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("storage", "locations", resp.StatusCode, string(body))
		}
	}
}

// BuildAddLocationRequest instantiates a HTTP request object with method and
// path set to call the "storage" service "add_location" endpoint
func (c *Client) BuildAddLocationRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: AddLocationStoragePath()}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("storage", "add_location", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// EncodeAddLocationRequest returns an encoder for requests sent to the storage
// add_location server.
func EncodeAddLocationRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*storage.AddLocationPayload)
		if !ok {
			return goahttp.ErrInvalidType("storage", "add_location", "*storage.AddLocationPayload", v)
		}
		body := NewAddLocationRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("storage", "add_location", err)
		}
		return nil
	}
}

// DecodeAddLocationResponse returns a decoder for responses returned by the
// storage add_location endpoint. restoreBody controls whether the response
// body should be restored after having been read.
// DecodeAddLocationResponse may return the following errors:
//   - "not_valid" (type *goa.ServiceError): http.StatusBadRequest
//   - error: internal error
func DecodeAddLocationResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = io.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusCreated:
			var (
				body AddLocationResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "add_location", err)
			}
			err = ValidateAddLocationResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "add_location", err)
			}
			res := NewAddLocationResultCreated(&body)
			return res, nil
		case http.StatusBadRequest:
			var (
				body AddLocationNotValidResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "add_location", err)
			}
			err = ValidateAddLocationNotValidResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "add_location", err)
			}
			return nil, NewAddLocationNotValid(&body)
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("storage", "add_location", resp.StatusCode, string(body))
		}
	}
}

// BuildMoveRequest instantiates a HTTP request object with method and path set
// to call the "storage" service "move" endpoint
func (c *Client) BuildMoveRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	var (
		aipID string
	)
	{
		p, ok := v.(*storage.MovePayload)
		if !ok {
			return nil, goahttp.ErrInvalidType("storage", "move", "*storage.MovePayload", v)
		}
		aipID = p.AipID
	}
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: MoveStoragePath(aipID)}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("storage", "move", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// EncodeMoveRequest returns an encoder for requests sent to the storage move
// server.
func EncodeMoveRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*storage.MovePayload)
		if !ok {
			return goahttp.ErrInvalidType("storage", "move", "*storage.MovePayload", v)
		}
		body := NewMoveRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("storage", "move", err)
		}
		return nil
	}
}

// DecodeMoveResponse returns a decoder for responses returned by the storage
// move endpoint. restoreBody controls whether the response body should be
// restored after having been read.
// DecodeMoveResponse may return the following errors:
//   - "not_available" (type *goa.ServiceError): http.StatusConflict
//   - "not_valid" (type *goa.ServiceError): http.StatusBadRequest
//   - "not_found" (type *storage.StoragePackageNotfound): http.StatusNotFound
//   - error: internal error
func DecodeMoveResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = io.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusAccepted:
			return nil, nil
		case http.StatusConflict:
			var (
				body MoveNotAvailableResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "move", err)
			}
			err = ValidateMoveNotAvailableResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "move", err)
			}
			return nil, NewMoveNotAvailable(&body)
		case http.StatusBadRequest:
			var (
				body MoveNotValidResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "move", err)
			}
			err = ValidateMoveNotValidResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "move", err)
			}
			return nil, NewMoveNotValid(&body)
		case http.StatusNotFound:
			var (
				body MoveNotFoundResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "move", err)
			}
			err = ValidateMoveNotFoundResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "move", err)
			}
			return nil, NewMoveNotFound(&body)
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("storage", "move", resp.StatusCode, string(body))
		}
	}
}

// BuildMoveStatusRequest instantiates a HTTP request object with method and
// path set to call the "storage" service "move_status" endpoint
func (c *Client) BuildMoveStatusRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	var (
		aipID string
	)
	{
		p, ok := v.(*storage.MoveStatusPayload)
		if !ok {
			return nil, goahttp.ErrInvalidType("storage", "move_status", "*storage.MoveStatusPayload", v)
		}
		aipID = p.AipID
	}
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: MoveStatusStoragePath(aipID)}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("storage", "move_status", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// DecodeMoveStatusResponse returns a decoder for responses returned by the
// storage move_status endpoint. restoreBody controls whether the response body
// should be restored after having been read.
// DecodeMoveStatusResponse may return the following errors:
//   - "failed_dependency" (type *goa.ServiceError): http.StatusFailedDependency
//   - "not_found" (type *storage.StoragePackageNotfound): http.StatusNotFound
//   - error: internal error
func DecodeMoveStatusResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = io.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusOK:
			var (
				body MoveStatusResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "move_status", err)
			}
			err = ValidateMoveStatusResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "move_status", err)
			}
			res := NewMoveStatusResultOK(&body)
			return res, nil
		case http.StatusFailedDependency:
			var (
				body MoveStatusFailedDependencyResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "move_status", err)
			}
			err = ValidateMoveStatusFailedDependencyResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "move_status", err)
			}
			return nil, NewMoveStatusFailedDependency(&body)
		case http.StatusNotFound:
			var (
				body MoveStatusNotFoundResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "move_status", err)
			}
			err = ValidateMoveStatusNotFoundResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "move_status", err)
			}
			return nil, NewMoveStatusNotFound(&body)
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("storage", "move_status", resp.StatusCode, string(body))
		}
	}
}

// BuildRejectRequest instantiates a HTTP request object with method and path
// set to call the "storage" service "reject" endpoint
func (c *Client) BuildRejectRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	var (
		aipID string
	)
	{
		p, ok := v.(*storage.RejectPayload)
		if !ok {
			return nil, goahttp.ErrInvalidType("storage", "reject", "*storage.RejectPayload", v)
		}
		aipID = p.AipID
	}
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: RejectStoragePath(aipID)}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("storage", "reject", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// DecodeRejectResponse returns a decoder for responses returned by the storage
// reject endpoint. restoreBody controls whether the response body should be
// restored after having been read.
// DecodeRejectResponse may return the following errors:
//   - "not_available" (type *goa.ServiceError): http.StatusConflict
//   - "not_valid" (type *goa.ServiceError): http.StatusBadRequest
//   - "not_found" (type *storage.StoragePackageNotfound): http.StatusNotFound
//   - error: internal error
func DecodeRejectResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = io.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusAccepted:
			return nil, nil
		case http.StatusConflict:
			var (
				body RejectNotAvailableResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "reject", err)
			}
			err = ValidateRejectNotAvailableResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "reject", err)
			}
			return nil, NewRejectNotAvailable(&body)
		case http.StatusBadRequest:
			var (
				body RejectNotValidResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "reject", err)
			}
			err = ValidateRejectNotValidResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "reject", err)
			}
			return nil, NewRejectNotValid(&body)
		case http.StatusNotFound:
			var (
				body RejectNotFoundResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "reject", err)
			}
			err = ValidateRejectNotFoundResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "reject", err)
			}
			return nil, NewRejectNotFound(&body)
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("storage", "reject", resp.StatusCode, string(body))
		}
	}
}

// BuildShowRequest instantiates a HTTP request object with method and path set
// to call the "storage" service "show" endpoint
func (c *Client) BuildShowRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	var (
		aipID string
	)
	{
		p, ok := v.(*storage.ShowPayload)
		if !ok {
			return nil, goahttp.ErrInvalidType("storage", "show", "*storage.ShowPayload", v)
		}
		aipID = p.AipID
	}
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: ShowStoragePath(aipID)}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("storage", "show", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// DecodeShowResponse returns a decoder for responses returned by the storage
// show endpoint. restoreBody controls whether the response body should be
// restored after having been read.
// DecodeShowResponse may return the following errors:
//   - "not_found" (type *storage.StoragePackageNotfound): http.StatusNotFound
//   - error: internal error
func DecodeShowResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = io.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusOK:
			var (
				body ShowResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "show", err)
			}
			p := NewShowStoredStoragePackageOK(&body)
			view := "default"
			vres := &storageviews.StoredStoragePackage{Projected: p, View: view}
			if err = storageviews.ValidateStoredStoragePackage(vres); err != nil {
				return nil, goahttp.ErrValidationError("storage", "show", err)
			}
			res := storage.NewStoredStoragePackage(vres)
			return res, nil
		case http.StatusNotFound:
			var (
				body ShowNotFoundResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "show", err)
			}
			err = ValidateShowNotFoundResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "show", err)
			}
			return nil, NewShowNotFound(&body)
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("storage", "show", resp.StatusCode, string(body))
		}
	}
}

// BuildShowLocationRequest instantiates a HTTP request object with method and
// path set to call the "storage" service "show-location" endpoint
func (c *Client) BuildShowLocationRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	var (
		uuid uuid.UUID
	)
	{
		p, ok := v.(*storage.ShowLocationPayload)
		if !ok {
			return nil, goahttp.ErrInvalidType("storage", "show-location", "*storage.ShowLocationPayload", v)
		}
		uuid = p.UUID
	}
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: ShowLocationStoragePath(uuid)}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("storage", "show-location", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// DecodeShowLocationResponse returns a decoder for responses returned by the
// storage show-location endpoint. restoreBody controls whether the response
// body should be restored after having been read.
// DecodeShowLocationResponse may return the following errors:
//   - "not_found" (type *storage.StorageLocationNotfound): http.StatusNotFound
//   - error: internal error
func DecodeShowLocationResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = io.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusOK:
			var (
				body ShowLocationResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "show-location", err)
			}
			p := NewShowLocationStoredLocationOK(&body)
			view := "default"
			vres := &storageviews.StoredLocation{Projected: p, View: view}
			if err = storageviews.ValidateStoredLocation(vres); err != nil {
				return nil, goahttp.ErrValidationError("storage", "show-location", err)
			}
			res := storage.NewStoredLocation(vres)
			return res, nil
		case http.StatusNotFound:
			var (
				body ShowLocationNotFoundResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "show-location", err)
			}
			err = ValidateShowLocationNotFoundResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "show-location", err)
			}
			return nil, NewShowLocationNotFound(&body)
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("storage", "show-location", resp.StatusCode, string(body))
		}
	}
}

// unmarshalStoredLocationResponseToStorageviewsStoredLocationView builds a
// value of type *storageviews.StoredLocationView from a value of type
// *StoredLocationResponse.
func unmarshalStoredLocationResponseToStorageviewsStoredLocationView(v *StoredLocationResponse) *storageviews.StoredLocationView {
	res := &storageviews.StoredLocationView{
		ID:          v.ID,
		Name:        v.Name,
		Description: v.Description,
		Source:      v.Source,
		Purpose:     v.Purpose,
		UUID:        v.UUID,
	}
	if v.Config != nil {
		switch *v.Config.Type {
		case "s3":
			var val *storageviews.S3ConfigView
			json.Unmarshal([]byte(*v.Config.Value), &val)
			res.Config = val
		}
	}

	return res
}
