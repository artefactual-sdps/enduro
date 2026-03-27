// Package ssblob provides a blob implementation for the Archivematica Storage
// Service. Use OpenBucket to construct a *blob.Bucket.
package ssblob

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"go.artefactual.dev/ssclient"
	"gocloud.dev/blob"
	"gocloud.dev/blob/driver"
	"gocloud.dev/gcerrors"
)

var errNotImplemented = errors.New("not implemented")

type APIError struct {
	Status string
	Code   int
	Cause  error
}

func (err *APIError) Error() string {
	if err == nil {
		return "<nil>"
	}
	if err.Status == "" && err.Cause != nil {
		return err.Cause.Error()
	}
	return err.Status
}

func (err *APIError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Cause
}

type bucket struct {
	// client is the Storage Service API client used to perform requests.
	client *ssclient.Client
}

type Options struct {
	URL      string
	Key      string
	Username string
	// HTTPClient optionally provides the outbound HTTP client used to create
	// the underlying Storage Service API client. If nil, http.DefaultClient is
	// used.
	HTTPClient *http.Client
}

func openBucket(opts *Options) (driver.Bucket, error) {
	if _, err := url.Parse(opts.URL); err != nil {
		return nil, err
	}

	ssclient, err := ssclient.New(ssclient.Config{
		BaseURL:    opts.URL,
		Username:   opts.Username,
		Key:        opts.Key,
		HTTPClient: opts.HTTPClient,
	})
	if err != nil {
		return nil, err
	}

	return &bucket{
		client: ssclient,
	}, nil
}

func OpenBucket(opts *Options) (*blob.Bucket, error) {
	drv, err := openBucket(opts)
	if err != nil {
		return nil, err
	}
	return blob.NewBucket(drv), nil
}

func (b *bucket) ErrorCode(err error) gcerrors.ErrorCode {
	if apiErr, ok := errors.AsType[*APIError](err); ok {
		switch {
		case apiErr.Code == http.StatusNotFound:
			return gcerrors.NotFound
		case apiErr.Code == http.StatusUnauthorized:
			return gcerrors.PermissionDenied
		case apiErr.Code >= 500:
			return gcerrors.Internal
		case apiErr.Code >= 400:
			return gcerrors.Unknown
		}
	}
	if errors.Is(err, errNotImplemented) {
		return gcerrors.Unimplemented
	}

	return gcerrors.Unknown
}

func (b *bucket) As(i any) bool {
	switch p := i.(type) {
	case **ssclient.Client:
		*p = b.client
		return true
	default:
		return false
	}
}

func (b *bucket) ErrorAs(err error, i any) bool {
	apiErr, ok := errors.AsType[*APIError](err)
	if !ok {
		return false
	}
	p, ok := i.(**APIError)
	if !ok {
		return false
	}

	*p = apiErr
	return true
}

func (b *bucket) Attributes(ctx context.Context, key string) (*driver.Attributes, error) {
	return nil, errNotImplemented
}

func (b *bucket) ListPaged(ctx context.Context, opts *driver.ListOptions) (*driver.ListPage, error) {
	return nil, errNotImplemented
}

func (b *bucket) NewRangeReader(
	ctx context.Context,
	key string,
	offset, length int64,
	opts *driver.ReaderOptions,
) (driver.Reader, error) {
	// Storage Service only supports full-object downloads.
	if offset != 0 || length >= 0 {
		return nil, errNotImplemented
	}

	id, err := uuid.Parse(key)
	if err != nil {
		return nil, err
	}

	stream, err := b.client.Packages().DownloadPackage(ctx, id)
	if err != nil {
		if apiErr := apiError(err); apiErr != nil {
			return nil, apiErr
		}
		return nil, err
	}

	return &reader{
		r: stream.Body,
		attrs: driver.ReaderAttributes{
			ContentType: stream.ContentType,
			Size:        stream.ContentLength,
		},
	}, nil
}

func (b *bucket) NewTypedWriter(
	ctx context.Context,
	key, contentType string,
	opts *driver.WriterOptions,
) (driver.Writer, error) {
	return nil, errNotImplemented
}

func (b *bucket) Copy(ctx context.Context, dstKey, srcKey string, opts *driver.CopyOptions) error {
	return errNotImplemented
}

func (b *bucket) Delete(ctx context.Context, key string) error {
	return errNotImplemented
}

func (b *bucket) SignedURL(ctx context.Context, key string, opts *driver.SignedURLOptions) (string, error) {
	return "", errNotImplemented
}

func (b *bucket) Close() error {
	return nil
}

// reader should be able to read an AIP object from Storage Service using *http.Client.
type reader struct {
	r     io.ReadCloser
	attrs driver.ReaderAttributes
}

func (r *reader) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

func (r *reader) Close() error {
	return r.r.Close()
}

func (r *reader) Attributes() *driver.ReaderAttributes {
	return &r.attrs
}

func (r *reader) As(i any) bool {
	return false
}

func apiError(err error) *APIError {
	if err == nil {
		return nil
	}

	if apiErr, ok := errors.AsType[*APIError](err); ok {
		return apiErr
	}

	if responseErr, ok := errors.AsType[*ssclient.ResponseError](err); ok {
		return &APIError{
			Status: statusText(responseErr.StatusCode, err),
			Code:   responseErr.StatusCode,
			Cause:  err,
		}
	}

	if unavailableErr, ok := errors.AsType[*ssclient.NotAvailableError](err); ok {
		return &APIError{
			Status: statusText(unavailableErr.StatusCode, err),
			Code:   unavailableErr.StatusCode,
			Cause:  err,
		}
	}

	return nil
}

func statusText(code int, fallback error) string {
	if status := http.StatusText(code); status != "" {
		return status
	}
	if fallback != nil {
		return fallback.Error()
	}
	return ""
}
