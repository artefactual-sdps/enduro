// Package ssblob provides a blob implementation for the Archivematica Storage
// Service. Use OpenBucket to construct a *blob.Bucket.
package ssblob

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"gocloud.dev/blob"
	"gocloud.dev/blob/driver"
	"gocloud.dev/gcerrors"
)

type NotImplemented error

var errNotImpl NotImplemented

type bucket struct {
	Options
}

type Options struct {
	Context context.Context
	URL     string `json:"string"`
	Base    string
}

func openBucket(opts *Options) (driver.Bucket, error) {
	return &bucket{
		*opts,
	}, nil
}

func OpenBucket(opts *Options) (*blob.Bucket, error) {
	drv, err := openBucket(opts)
	if err != nil {
		return nil, err
	}
	return blob.NewBucket(drv), nil
}

func (b *bucket) ErrorCode(error) gcerrors.ErrorCode {
	return gcerrors.Unknown
}

func (b *bucket) As(i interface{}) bool {
	return false
}

func (b *bucket) ErrorAs(error, interface{}) bool {
	return false
}

func (b *bucket) Attributes(ctx context.Context, key string) (*driver.Attributes, error) {
	return nil, errNotImpl
}

func (b *bucket) ListPaged(ctx context.Context, opts *driver.ListOptions) (*driver.ListPage, error) {
	return nil, errNotImpl
}

func (b *bucket) NewRangeReader(ctx context.Context, key string, offset, length int64, opts *driver.ReaderOptions) (driver.Reader, error) {
	client := http.Client{}
	bu, err := url.Parse(b.URL)
	if err != nil {
		return nil, err
	}
	if b.Base == "" {
		b.Base = "."
	}
	path, err := url.JoinPath(b.Base, key)
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
	resp, err := client.Do(req)
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
	return nil, errNotImpl
}

func (b *bucket) Copy(ctx context.Context, dstKey, srcKey string, opts *driver.CopyOptions) error {
	return errNotImpl
}

func (b *bucket) Delete(ctx context.Context, key string) error {
	return errNotImpl
}

func (b *bucket) SignedURL(ctx context.Context, key string, opts *driver.SignedURLOptions) (string, error) {
	return "", errNotImpl
}

func (b *bucket) Close() error {
	return nil
}

// Reader should be able to read an AIP object from Storage Service using *http.Client.
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
