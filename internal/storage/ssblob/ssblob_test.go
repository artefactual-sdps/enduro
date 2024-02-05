package ssblob_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage/ssblob"
)

func TestBucket(t *testing.T) {
	t.Parallel()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write([]byte("hello"))
	})
	srv := httptest.NewServer(
		handler,
	)
	opts := ssblob.Options{
		URL: srv.URL,
	}
	t.Run("Basic download from the amss", func(t *testing.T) {
		t.Parallel()

		bucket, err := ssblob.OpenBucket(&opts)
		assert.NilError(t, err)
		defer bucket.Close()

		r, err := bucket.NewReader(context.Background(), "64273703-f1f6-4588-85bd-5facc852a1be", nil)
		assert.NilError(t, err)

		n, err := io.ReadAll(r)
		assert.NilError(t, err)
		assert.Assert(t, len(n) > 0)
		// change content type
		assert.Equal(t, r.ContentType(), "image/jpeg")
	})
}
