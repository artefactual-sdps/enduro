package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-logr/logr/funcr"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestRescoverMiddleware(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	var logged string
	logger := funcr.New(
		func(prefix, args string) { logged = args },
		funcr.Options{},
	)

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) { panic("opsie") })
	mw := recoverMiddleware(logger)

	mw(handler).ServeHTTP(w, req)

	assert.Assert(t, cmp.Contains(logged, "\"msg\"=\"Panic error recovered.\""))
	assert.Assert(t, cmp.Contains(logged, "\"error\"=\"panic: opsie"))
}

func TestVersionHeaderMiddleware(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	var continued bool
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) { continued = true })
	mw := versionHeaderMiddleware("v1.2.3")

	mw(handler).ServeHTTP(w, req)
	resp := w.Result()

	assert.Equal(t, resp.Header.Get("X-Enduro-Version"), "v1.2.3")
	assert.Equal(t, continued, true)
}

func TestWriteTimeout(t *testing.T) {
	t.Parallel()

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Microsecond * 100)
		w.Write([]byte("Hi there!"))
	})

	t.Run("Sets a write timeout", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(writeTimeout(h, time.Microsecond))
		defer ts.Close()

		_, err := ts.Client().Get(ts.URL)
		assert.ErrorIs(t, err, io.EOF)
	})

	t.Run("Sets an unlimited write timeout", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(writeTimeout(h, 0))
		defer ts.Close()

		resp, err := ts.Client().Get(ts.URL)
		assert.NilError(t, err)

		blob, err := io.ReadAll(resp.Body)
		assert.NilError(t, err)
		assert.Equal(t, string(blob), "Hi there!")
	})
}

func TestReadTimeout(t *testing.T) {
	t.Parallel()

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestTimeout)
		}
	})

	for _, tt := range []struct {
		name     string
		timeout  time.Duration
		respCode int
	}{
		{
			name:     "Sets a read timeout",
			timeout:  time.Millisecond * 10,
			respCode: http.StatusRequestTimeout,
		},
		{
			name:     "Sets an unlimited read timeout",
			timeout:  0,
			respCode: http.StatusOK,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewServer(readTimeout(h, tt.timeout))
			defer ts.Close()

			pr, pw := io.Pipe()
			defer pr.Close()
			go func() {
				time.Sleep(time.Millisecond * 100)
				pw.Write([]byte("body"))
				pw.Close()
			}()

			resp, err := ts.Client().Post(ts.URL, "application/octet-stream", pr)
			assert.NilError(t, err)
			assert.Equal(t, resp.StatusCode, tt.respCode)
		})
	}
}
