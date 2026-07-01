package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/httplog/v3"
	"github.com/go-logr/logr"
)

// logHandler returns a middleware that logs HTTP requests and responses using
// the provided logr.Logger in Elastic Common Schema (ECS) format. If debug
// parameter is true, request and response bodies are also logged.
//
// Success response statuses (100-399 and 429) are logged at level V(2).
// Error response statuses (400-428, 430-599) at level V(0) (highest).
// OPTIONS requests are not logged.
func logHandler(logger logr.Logger, debug bool) func(http.Handler) http.Handler {
	return httplog.RequestLogger(loggerAdapter(logger), &httplog.Options{
		// Level defines the verbosity of the request logs:
		// slog.LevelDebug - log all responses (incl. OPTIONS)
		// slog.LevelInfo  - log responses (excl. OPTIONS)
		// slog.LevelWarn  - log 4xx and 5xx responses only (except for 429)
		// slog.LevelError - log 5xx responses only
		Level: slog.LevelInfo,

		// Set log output to Elastic Common Schema (ECS) format.
		Schema: httplog.SchemaECS.Concise(false),

		// RecoverPanics recovers from panics occurring in the underlying HTTP
		// handlers and middlewares. It returns HTTP 500 unless response status
		// was already set.
		//
		// NOTE: Panics are logged as errors automatically, regardless of this
		// setting.
		RecoverPanics: true,

		// Optionally, filter out some request logs.
		// Skip: func(req *http.Request, respStatus int) bool {
		// 	return respStatus == 404 || respStatus == 405
		// },

		// Optionally, log selected request/response headers explicitly.
		// LogRequestHeaders:  []string{"Origin"},
		// LogResponseHeaders: []string{},

		// Optionally, enable logging of request/response body based on custom
		// conditions.
		// Useful for debugging payload issues in development.
		LogRequestBody:  func(req *http.Request) bool { return debug },
		LogResponseBody: func(req *http.Request) bool { return debug },
	})
}

// loggerAdapter converts a logr.Logger to a slog.Logger for use with httplog.
// The original logr.LogSink is used as the underlying log handler.
func loggerAdapter(l logr.Logger) *slog.Logger {
	// Setting the logr.Logger level to V(2) causes slog.LevelInfo messages to
	// be logged at V(2). Higher level messages (slog.LevelWarn and
	// slog.LevelError) are logged at V(0) (highest).
	h := logr.ToSlogHandler(l.V(2).WithName("api"))
	return slog.New(h)
}
