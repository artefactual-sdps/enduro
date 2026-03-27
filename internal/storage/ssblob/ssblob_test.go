package ssblob_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.artefactual.dev/ssclient"
	"gocloud.dev/blob"
	"gocloud.dev/gcerrors"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage/ssblob"
)

// setUpTest creates a ssblob bucket configured against a fake SS API.
func setUpTest(t *testing.T, h http.HandlerFunc, opts *ssblob.Options) *blob.Bucket {
	t.Helper()

	if opts == nil {
		opts = &ssblob.Options{}
	}

	srv := httptest.NewServer(h)
	t.Cleanup(func() { srv.Close() })
	opts.URL = srv.URL

	b, err := ssblob.OpenBucket(opts)
	assert.NilError(t, err)
	t.Cleanup(func() { b.Close() })

	return b
}

func TestBucket(t *testing.T) {
	t.Parallel()

	t.Run("Downloads an AIP from SS", func(t *testing.T) {
		t.Parallel()

		b := setUpTest(t,
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Header.Get("Authorization"), "ApiKey test:test")
				assert.Equal(t, r.URL.Path, "/api/v2/file/2db707f3-3cd2-44b7-9012-9b68eb10d207/download/")

				w.Header().Set("Content-Type", "text/plain")
				w.Header().Set("Content-Disposition", "attachment; filename=\"hello.txt\"")
				_, err := w.Write([]byte("Hello World!"))
				assert.NilError(t, err)
			},
			&ssblob.Options{
				Username: "test",
				Key:      "test",
			},
		)

		r, err := b.NewReader(context.Background(), "2db707f3-3cd2-44b7-9012-9b68eb10d207", nil)
		assert.NilError(t, err)
		defer r.Close()

		blob, err := io.ReadAll(r)
		assert.NilError(t, err)
		assert.DeepEqual(t, string(blob), "Hello World!")
		assert.Equal(t, r.ContentType(), "text/plain")
	})

	t.Run("Exposes the underlying Storage Service client", func(t *testing.T) {
		t.Parallel()

		b := setUpTest(t, nil, nil)

		var client *ssclient.Client
		assert.Equal(t, b.As(&client), true)
		assert.Assert(t, client != nil)
	})

	t.Run("Rejects unsupported As types", func(t *testing.T) {
		t.Parallel()

		b := setUpTest(t, nil, nil)

		var s string
		assert.Equal(t, b.As(&s), false)
	})

	t.Run("Returns an error if the AIP is not found", func(t *testing.T) {
		t.Parallel()

		b := setUpTest(t, func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "AIP not found.", http.StatusNotFound)
		}, nil)

		r, err := b.NewReader(context.Background(), "2db707f3-3cd2-44b7-9012-9b68eb10d207", nil)
		assert.Equal(t, gcerrors.Code(err), gcerrors.NotFound)
		assert.Assert(t, r == nil)

		apiErr := &ssblob.APIError{}
		assert.Equal(t, b.ErrorAs(err, &apiErr), true)
		assert.Equal(t, apiErr.Code, http.StatusNotFound)
	})

	t.Run("Returns an internal error on server failure", func(t *testing.T) {
		t.Parallel()

		b := setUpTest(t, func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "server error", http.StatusInternalServerError)
		}, nil)

		r, err := b.NewReader(context.Background(), "2db707f3-3cd2-44b7-9012-9b68eb10d207", nil)
		assert.Equal(t, gcerrors.Code(err), gcerrors.Internal)
		assert.Assert(t, r == nil)

		apiErr := &ssblob.APIError{}
		assert.Equal(t, b.ErrorAs(err, &apiErr), true)
		assert.Equal(t, apiErr.Code, http.StatusInternalServerError)
	})

	t.Run("Matches wrapped API errors", func(t *testing.T) {
		t.Parallel()

		b := setUpTest(t, nil, nil)

		wrapped := fmt.Errorf("wrapped: %w", &ssblob.APIError{
			Status: http.StatusText(http.StatusNotFound),
			Code:   http.StatusNotFound,
		})

		apiErr := &ssblob.APIError{}
		assert.Equal(t, b.ErrorAs(wrapped, &apiErr), true)
		assert.Equal(t, apiErr.Code, http.StatusNotFound)
	})

	t.Run("Rejects ranged reads", func(t *testing.T) {
		t.Parallel()

		b := setUpTest(t, func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("unexpected request")
		}, nil)

		r, err := b.NewRangeReader(context.Background(), "2db707f3-3cd2-44b7-9012-9b68eb10d207", 1, 10, nil)
		assert.Equal(t, gcerrors.Code(err), gcerrors.Unimplemented)
		assert.Assert(t, r == nil)
	})

	t.Run("Rejects unsupported operations", func(t *testing.T) {
		t.Parallel()

		b := setUpTest(t, func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("unexpected request")
		}, nil)

		ctx := context.Background()

		attrs, err := b.Attributes(ctx, "2db707f3-3cd2-44b7-9012-9b68eb10d207")
		assert.Assert(t, attrs == nil)
		assert.Equal(t, gcerrors.Code(err), gcerrors.Unimplemented)

		objs, nextPageToken, err := b.ListPage(ctx, nil, 10, nil)
		assert.Assert(t, objs == nil)
		assert.Assert(t, nextPageToken == nil)
		assert.Assert(t, err != nil)

		iter := b.List(nil)
		obj, err := iter.Next(ctx)
		assert.Assert(t, obj == nil)
		assert.Assert(t, err != nil)

		err = b.WriteAll(ctx, "2db707f3-3cd2-44b7-9012-9b68eb10d207", []byte("hello"), nil)
		assert.Equal(t, gcerrors.Code(err), gcerrors.Unimplemented)

		err = b.Copy(ctx, "dst", "src", nil)
		assert.Equal(t, gcerrors.Code(err), gcerrors.Unimplemented)

		err = b.Delete(ctx, "2db707f3-3cd2-44b7-9012-9b68eb10d207")
		assert.Equal(t, gcerrors.Code(err), gcerrors.Unimplemented)

		signedURL, err := b.SignedURL(ctx, "2db707f3-3cd2-44b7-9012-9b68eb10d207", nil)
		assert.Equal(t, signedURL, "")
		assert.Equal(t, gcerrors.Code(err), gcerrors.Unimplemented)
	})

	t.Run("Rejects zero-length ranged reads", func(t *testing.T) {
		t.Parallel()

		b := setUpTest(t, func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("unexpected request")
		}, nil)

		r, err := b.NewRangeReader(context.Background(), "2db707f3-3cd2-44b7-9012-9b68eb10d207", 0, 0, nil)
		assert.Equal(t, gcerrors.Code(err), gcerrors.Unimplemented)
		assert.Assert(t, r == nil)
	})

	t.Run("Returns an error if the request is unauthorized", func(t *testing.T) {
		t.Parallel()

		b := setUpTest(t, func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "User is unauthorized.", http.StatusUnauthorized)
		}, nil)

		r, err := b.NewReader(context.Background(), "2db707f3-3cd2-44b7-9012-9b68eb10d207", nil)
		assert.Equal(t, gcerrors.Code(err), gcerrors.PermissionDenied)
		assert.Assert(t, r == nil)

		apiErr := &ssblob.APIError{}
		assert.Equal(t, b.ErrorAs(err, &apiErr), true)
		assert.Equal(t, apiErr.Code, http.StatusUnauthorized)
	})

	t.Run("Returns an error if the package is not available", func(t *testing.T) {
		t.Parallel()

		b := setUpTest(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			err := fmt.Errorf(`{"message":"package is not available"}`)
			_, writeErr := w.Write([]byte(err.Error()))
			assert.NilError(t, writeErr)
		}, nil)

		r, err := b.NewReader(context.Background(), "2db707f3-3cd2-44b7-9012-9b68eb10d207", nil)
		assert.Assert(t, r == nil)

		apiErr := &ssblob.APIError{}
		assert.Equal(t, b.ErrorAs(err, &apiErr), true)
		assert.Equal(t, apiErr.Code, http.StatusAccepted)
		assert.Equal(t, apiErr.Error(), http.StatusText(http.StatusAccepted))
	})

	t.Run("Exposes reader and API error behavior", func(t *testing.T) {
		t.Parallel()

		apiErr := &ssblob.APIError{Cause: errors.New("boom")}
		assert.Equal(t, apiErr.Error(), "boom")
		assert.Equal(t, errors.Unwrap(apiErr).Error(), "boom")

		b := setUpTest(t,
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				_, err := w.Write([]byte("Hello World!"))
				assert.NilError(t, err)
			},
			nil,
		)

		r, err := b.NewReader(context.Background(), "2db707f3-3cd2-44b7-9012-9b68eb10d207", nil)
		assert.NilError(t, err)
		defer r.Close()

		var s string
		assert.Equal(t, r.As(&s), false)
	})

	t.Run("Returns an error if the URL is invalid", func(t *testing.T) {
		b, err := ssblob.OpenBucket(&ssblob.Options{
			URL: string([]byte{0x7f}), // DEL character is rejected.
		})
		assert.ErrorContains(t, err, "net/url: invalid control character in URL")
		assert.Assert(t, b == nil)
	})
}
