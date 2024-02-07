package ssblob_test

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage/ssblob"
)

func TestBucket(t *testing.T) {
	t.Parallel()

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Username") != "test@example.com" && r.Header.Get("ApiKey") != "api_key_example" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	type test struct {
		name, want, wantErr string
		ssblob.Options
		handler http.Handler
	}
	for _, tt := range []test{
		{
			name: "Download a package from the AMSS",
			want: "hello AMSS",
			Options: ssblob.Options{
				Key:      "api_key_example",
				Username: "test@example.com",
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/zip")
				zipWriter := zip.NewWriter(w)
				defer zipWriter.Close()
				zw, _ := zipWriter.Create("Storage-Service-AIP")
				zw.Write([]byte("hello AMSS"))
			}),
		},
		{
			name:    "Return an error when the server fails",
			wantErr: fmt.Sprint(http.StatusInternalServerError),
			Options: ssblob.Options{
				Key:      "api_key_example",
				Username: "test@example.com",
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/zip")
			}),
		},
		{
			name:    "Return an error when request has no auth",
			wantErr: fmt.Sprint(http.StatusUnauthorized),
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			}),
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(
				middleware(tt.handler),
			)
			defer srv.Close()

			opts := ssblob.Options{
				URL:      srv.URL,
				Username: tt.Options.Username,
				Key:      tt.Options.Key,
			}

			bucket, err := ssblob.OpenBucket(&opts)
			if err != nil {
				assert.Error(t, err, tt.wantErr)
				return
			}
			defer bucket.Close()

			r, err := bucket.NewReader(context.Background(), "", nil)
			// We check if the header of the response is an http error code.
			if err != nil {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			defer r.Close()

			n, err := io.ReadAll(r)
			if err != nil {
				assert.Error(t, err, tt.wantErr)
				return
			}

			zr, err := zip.NewReader(bytes.NewReader(n), int64(len(n)))
			if err != nil {
				assert.Error(t, err, tt.wantErr)
				return
			}

			file, err := zr.Open("Storage-Service-AIP")
			if err != nil {
				assert.Error(t, err, tt.wantErr)
				return
			}
			defer file.Close()

			bytes, err := io.ReadAll(file)
			if err != nil {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.DeepEqual(t, tt.want, string(bytes))
			assert.Equal(t, r.ContentType(), "application/zip")
		})
	}
}
