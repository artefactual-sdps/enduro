package api

import (
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"gotest.tools/v3/assert"
)

func TestSameOriginChecker(t *testing.T) {
	t.Parallel()

	t.Run("Undefined host passes", func(t *testing.T) {
		t.Parallel()
		check := sameOriginChecker(logr.Discard())
		req := httptest.NewRequest("GET", "http://example.com/foo", nil)
		assert.Equal(t, check(req), true)
	})

	t.Run("Same host passes", func(t *testing.T) {
		t.Parallel()
		check := sameOriginChecker(logr.Discard())
		req := httptest.NewRequest("GET", "http://example.com/foo", nil)
		req.Header.Add("Origin", "http://example.com")
		assert.Equal(t, check(req), true)
	})

	t.Run("Host mismatch fails to pass", func(t *testing.T) {
		t.Parallel()
		check := sameOriginChecker(logr.Discard())
		req := httptest.NewRequest("GET", "http://example.com/foo", nil)
		req.Header.Add("Origin", "http://example.net")
		assert.Equal(t, check(req), false)
	})
}
