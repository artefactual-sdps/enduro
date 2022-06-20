// Code generated by goa v3.7.6, DO NOT EDIT.
//
// storage HTTP client encoders and decoders
//
// Command:
// $ goa-v3.7.6 gen github.com/artefactual-labs/enduro/internal/api/design -o
// internal/api

package client

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"

	storage "github.com/artefactual-labs/enduro/internal/api/gen/storage"
	goahttp "goa.design/goa/v3/http"
)

// BuildSubmitRequest instantiates a HTTP request object with method and path
// set to call the "storage" service "submit" endpoint
func (c *Client) BuildSubmitRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: SubmitStoragePath()}
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
//	- "not_available" (type *goa.ServiceError): http.StatusConflict
//	- "not_valid" (type *goa.ServiceError): http.StatusBadRequest
//	- error: internal error
func DecodeSubmitResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
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
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("storage", "submit", resp.StatusCode, string(body))
		}
	}
}

// BuildUpdateRequest instantiates a HTTP request object with method and path
// set to call the "storage" service "update" endpoint
func (c *Client) BuildUpdateRequest(ctx context.Context, v interface{}) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: UpdateStoragePath()}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("storage", "update", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// EncodeUpdateRequest returns an encoder for requests sent to the storage
// update server.
func EncodeUpdateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	return func(req *http.Request, v interface{}) error {
		p, ok := v.(*storage.UpdatePayload)
		if !ok {
			return goahttp.ErrInvalidType("storage", "update", "*storage.UpdatePayload", v)
		}
		body := NewUpdateRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("storage", "update", err)
		}
		return nil
	}
}

// DecodeUpdateResponse returns a decoder for responses returned by the storage
// update endpoint. restoreBody controls whether the response body should be
// restored after having been read.
// DecodeUpdateResponse may return the following errors:
//	- "not_available" (type *goa.ServiceError): http.StatusConflict
//	- "not_valid" (type *goa.ServiceError): http.StatusBadRequest
//	- error: internal error
func DecodeUpdateResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (interface{}, error) {
	return func(resp *http.Response) (interface{}, error) {
		if restoreBody {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusAccepted:
			var (
				body UpdateResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("storage", "update", err)
			}
			err = ValidateUpdateResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("storage", "update", err)
			}
			res := NewUpdateResultAccepted(&body)
			return res, nil
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
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("storage", "update", resp.StatusCode, string(body))
		}
	}
}
