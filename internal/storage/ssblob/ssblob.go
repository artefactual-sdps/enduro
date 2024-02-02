// Package ssblob provides a blob implementation for the Archivematica Storage
// Service. Use OpenBucket to construct a *blob.Bucket.
package ssblob

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"

	"github.com/hashicorp/go-cleanhttp"
	"gocloud.dev/blob"
	"gocloud.dev/blob/driver"
	"gocloud.dev/gcerrors"
)

var (
	errNotImplemented = errors.New("not implemented")
	errNotFound       = errors.New("blob not found")
)

type bucket struct {
	Options Options
	client  *http.Client
}

type Options struct {
	URL      string
	Key      string
	Username string
	Client   *http.Client
}

func openBucket(opts *Options) (driver.Bucket, error) {
	// Will use the http client we pass with options if it is given.
	var cl *http.Client
	if reflect.ValueOf(opts.Client).IsZero() {
		cl = cleanhttp.DefaultPooledClient()
	} else {
		cl = opts.Client
	}
	return &bucket{
		Options: *opts,
		client:  cl,
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
	switch err {
	case errNotFound:
		return gcerrors.NotFound
	case errNotImplemented:
		return gcerrors.Unimplemented
	default:
		return gcerrors.Unknown
	}
}

func (b *bucket) As(i interface{}) bool {
	return false
}

func (b *bucket) ErrorAs(error, interface{}) bool {
	return false
}

func (b *bucket) Attributes(ctx context.Context, key string) (*driver.Attributes, error) {
	return nil, errNotImplemented
}

func (b *bucket) ListPaged(ctx context.Context, opts *driver.ListOptions) (*driver.ListPage, error) {
	return nil, errNotImplemented
}

func (b *bucket) NewRangeReader(ctx context.Context, key string, offset, length int64, opts *driver.ReaderOptions) (driver.Reader, error) {
	bu, err := url.Parse(b.Options.URL)
	if err != nil {
		return nil, err
	}
	if key != "" {
		b.Options.Key = key
	}
	path, err := url.JoinPath(b.Options.URL, b.Options.Key, b.Options.Username)
	if err != nil {
		return nil, err
	}
	says, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", bu.ResolveReference(says).String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	return &reader{
		r: resp.Body,
		attrs: driver.ReaderAttributes{
			ContentType: resp.Header.Get("Content-Type"),
		},
	}, nil
}

func (b *bucket) NewTypedWriter(ctx context.Context, key, contentType string, opts *driver.WriterOptions) (driver.Writer, error) {
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

func (r *reader) As(i interface{}) bool { return false }
