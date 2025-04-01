// Package ssblob provides a blob implementation for the Archivematica Storage
// Service. Use OpenBucket to construct a *blob.Bucket.
package ssblob

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"

	"gocloud.dev/blob"
	"gocloud.dev/blob/driver"
	"gocloud.dev/gcerrors"
)

var errNotImplemented = errors.New("not implemented")

type APIError struct {
	Status string
	Code   int
}

func (err *APIError) Error() string {
	return err.Status
}

type bucket struct {
	baseURL *url.URL
	client  *http.Client
}

type Options struct {
	URL      string
	Key      string
	Username string
}

func openBucket(opts *Options) (driver.Bucket, error) {
	u, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}
	return &bucket{
		baseURL: u,
		client:  NewClient(opts.Username, opts.Key),
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
	if err, ok := err.(*APIError); ok {
		switch {
		case err.Code == http.StatusNotFound:
			return gcerrors.NotFound
		case err.Code == http.StatusUnauthorized:
			return gcerrors.PermissionDenied
		case err.Code >= 400:
			return gcerrors.Unknown
		case err.Code >= 500:
			return gcerrors.Internal
		}
	}
	switch err {
	case errNotImplemented:
		return gcerrors.Unimplemented
	default:
		return gcerrors.Unknown
	}
}

func (b *bucket) As(i interface{}) bool {
	p, ok := i.(**http.Client)
	if !ok {
		return false
	}
	*p = b.client
	return true
}

func (b *bucket) ErrorAs(err error, i interface{}) bool {
	switch v := err.(type) {
	case *APIError:
		if p, ok := i.(**APIError); ok {
			*p = v
			return true
		}
	}
	return false
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
	url := b.baseURL.JoinPath("api/v2/file", key, "download")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{Status: resp.Status, Code: resp.StatusCode}
	}

	return &reader{
		r: resp.Body,
		attrs: driver.ReaderAttributes{
			ContentType: resp.Header.Get("Content-Type"),
			Size:        resp.ContentLength,
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

func (r *reader) As(i interface{}) bool {
	return false
}
