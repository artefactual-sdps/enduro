package api

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/httplog/v3"
)

// fileLogger wraps slog.Logger and the underlying log file descriptor.
// fileLogger implements the io.Closer interface, and Close() must be called to
// close the file descriptor when the logger is no longer needed.
type fileLogger struct {
	*slog.Logger
	cfg  LogConfig
	f    *os.File
	name string
}

// NewFileLogger creates a new fileLogger that writes logs to a file at
// cfg.Path. If cfg.Path is empty, the returned fileLogger will silently
// discard all log messages.
func NewFileLogger(cfg LogConfig) (*fileLogger, error) {
	fl := fileLogger{}
	if cfg.Path != "" {
		f, err := openLogFile(cfg.Path)
		if err != nil {
			return nil, err
		}
		fl.f = f
	}

	fl.cfg = cfg
	fl.Logger = slog.New(newHandler(cfg, fl.f))

	return &fl, nil
}

// Close closes the underlying log file descriptor, and implements the io.Closer
// interface. Close() should be called when the logger is no longer needed to
// avoid resource leaks. If the log file is stdout or stderr, Close() does
// nothing. Any log messages sent after Close() is called will be silently
// discarded.
func (fl *fileLogger) Close() error {
	if fl.f == nil || fl.f == os.Stdout || fl.f == os.Stderr {
		return nil
	}
	return fl.f.Close()
}

// WithName returns a Logger instance with the specified name element added
// to the Logger's name. Successive calls with WithName append additional
// suffixes to the Logger's name, separated by a dot. The log name is written to
// each produced message in a "service.name" field to identify the service
// that generated the log message. If name is empty, WithName returns the
// receiver fileLogger unchanged.
func (fl *fileLogger) WithName(name string) *fileLogger {
	if name == "" {
		return fl
	}

	if fl.name != "" && name != "" {
		name = fl.name + "." + name
	}

	return &fileLogger{
		Logger: slog.New(newHandler(fl.cfg, fl.f)).With("service.name", name),
		cfg:    fl.cfg,
		f:      fl.f,
		name:   name,
	}
}

// openLogFile opens the log file at the given path for appending, creating it
// if it does not already exist. If the path is "stdout" or "stderr" (case-
// insensitive), it returns the corresponding standard output or standard error
// file descriptor. If the path is empty, it returns an error.
func openLogFile(path string) (*os.File, error) {
	switch p := strings.ToLower(path); p {
	case "stderr":
		return os.Stderr, nil
	case "stdout":
		return os.Stdout, nil
	default:
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o640)
		if err != nil {
			return nil, err
		}
		return f, nil
	}
}

func newHandler(cfg LogConfig, f *os.File) slog.Handler {
	if cfg.Path == "" {
		return slog.DiscardHandler
	}

	var h slog.Handler
	switch cfg.Format {
	case LogFormatText:
		h = slog.NewTextHandler(f, &slog.HandlerOptions{Level: cfg.Level})
	default:
		h = slog.NewJSONHandler(f, &slog.HandlerOptions{Level: cfg.Level})
	}
	return h
}

// requestLogger returns a middleware that logs HTTP requests and responses
// using the provided slog.Logger in Elastic Common Schema (ECS) format.
func requestLogger(logger *slog.Logger, level slog.Level) func(http.Handler) http.Handler {
	return httplog.RequestLogger(logger, &httplog.Options{
		// Level defines the verbosity of the request logs:
		// slog.LevelDebug - log all responses (incl. OPTIONS)
		// slog.LevelInfo  - log responses (excl. OPTIONS)
		// slog.LevelWarn  - log 4xx and 5xx responses only (except for 429)
		// slog.LevelError - log 5xx responses only
		Level: level,

		// Set log output to Elastic Common Schema (ECS) format.
		Schema: httplog.SchemaECS,

		// RecoverPanics recovers from panics occurring in the underlying HTTP
		// handlers and middlewares. It returns HTTP 500 unless response status
		// was already set.
		//
		// NOTE: Panics are logged as errors automatically, regardless of this
		// setting.
		RecoverPanics: true,
	})
}
