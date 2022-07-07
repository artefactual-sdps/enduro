package api

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/go-logr/logr"
)

func recoverMiddleware(logger logr.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if r := recover(); r != nil {
					var msg string
					switch x := r.(type) {
					case string, error:
						msg = fmt.Sprintf("panic: %s", x)
					default:
						msg = "unknown panic"
					}
					const size = 64 << 10 // 64KB
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					lines := strings.Split(string(buf), "\n")
					stack := lines[3:]
					err := fmt.Errorf("%s\n%s", msg, strings.Join(stack, "\n"))
					logger.Error(err, "panic error")
				}
			}()
			h.ServeHTTP(w, r)
		})
	}
}

func versionHeaderMiddleware(version string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Enduro-Version", version)
			h.ServeHTTP(w, r)
		})
	}
}
