package api

import (
	"bytes"
	"errors"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

func recoverMiddleware(logger logr.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					// Don't recover if the request is aborted, otherwise the
					// request can't detect the error.
					if rec == http.ErrAbortHandler {
						panic(rec)
					}

					// Prepare error message and log it.
					b := strings.Builder{}
					switch x := rec.(type) {
					case string:
						b.WriteString("panic: ")
						b.WriteString(x)
					case error:
						b.WriteString("panic: ")
						b.WriteString(x.Error())
					default:
						b.WriteString("unknown panic")
					}
					const size = 64 << 10 // 64KB
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					lines := bytes.Split(buf, []byte{'\n'})
					if len(lines) > 3 {
						b.WriteByte('\n')
						for _, line := range lines[3:] {
							b.Write(line)
							b.WriteByte('\n')
						}
					}
					logger.Error(errors.New(b.String()), "Panic error recovered.")

					// Skip write header on upgrade connection.
					if r.Header.Get("Connection") != "Upgrade" {
						w.WriteHeader(http.StatusInternalServerError)
					}
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

// writeTimeout sets the write deadline for writing the response. A zero value
// means no timeout.
func writeTimeout(h http.Handler, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := http.NewResponseController(w)
		var deadline time.Time
		if timeout != 0 {
			deadline = time.Now().Add(timeout)
		}
		_ = rc.SetWriteDeadline(deadline)
		h.ServeHTTP(w, r)
	})
}
