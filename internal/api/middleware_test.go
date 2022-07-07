package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr/funcr"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestRescoverMiddleware(t *testing.T) {
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

	assert.Assert(t, cmp.Contains(logged, "\"msg\"=\"panic error\""))
	assert.Assert(t, cmp.Contains(logged, "\"error\"=\"panic: opsie"))
}

func TestVersionHeaderMiddleware(t *testing.T) {
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
