package api

import (
	"net/http"
	"net/url"
	"unicode/utf8"

	"github.com/go-logr/logr"
)

func sameOriginChecker(logger logr.Logger) func(r *http.Request) bool {
	return func(r *http.Request) bool {
		origin := r.Header["Origin"]
		if len(origin) == 0 {
			return true
		}
		u, err := url.Parse(origin[0])
		if err != nil {
			logger.V(1).Info("WebSocket client rejected (origin parse error)", "err", err)
			return false
		}
		eq := equalASCIIFold(u.Host, r.Host)
		if !eq {
			logger.V(1).Info("WebSocket client rejected (origin and host not equal)", "origin-host", u.Host, "request-host", r.Host)
		}
		return eq
	}
}

// equalASCIIFold returns true if s is equal to t with ASCII case folding as
// defined in RFC 4790.
func equalASCIIFold(s, t string) bool {
	for s != "" && t != "" {
		sr, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		tr, size := utf8.DecodeRuneInString(t)
		t = t[size:]
		if sr == tr {
			continue
		}
		if 'A' <= sr && sr <= 'Z' {
			sr = sr + 'a' - 'A'
		}
		if 'A' <= tr && tr <= 'Z' {
			tr = tr + 'a' - 'A'
		}
		if sr != tr {
			return false
		}
	}
	return s == t
}
