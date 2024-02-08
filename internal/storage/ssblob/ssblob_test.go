package ssblob_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
				assert.Equal(t, r.URL.Path, "/api/v2/file/2db707f3-3cd2-44b7-9012-9b68eb10d207/download")

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

	t.Run("Gives access to underlying HTTP client", func(t *testing.T) {
		t.Parallel()

		b := setUpTest(t, nil, nil)

		var client *http.Client
		assert.Equal(t, b.As(&client), true)
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

	t.Run("Returns an error if the URL is invalid", func(t *testing.T) {
		b, err := ssblob.OpenBucket(&ssblob.Options{
			URL: string([]byte{0x7f}), // DEL character is rejected.
		})
		assert.ErrorContains(t, err, "net/url: invalid control character in URL")
		assert.Assert(t, b == nil)
	})
}
