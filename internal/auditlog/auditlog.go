package auditlog

import (
	"context"
	"io"
	"log/slog"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger represents a structured audit event logger.
type Logger struct {
	l *slog.Logger
	w io.WriteCloser
}

// New creates a new audit logger using the provided writer w and logger l.
func New(w io.WriteCloser, l *slog.Logger) *Logger {
	return &Logger{l: l, w: w}
}

// NewFromConfig creates a new audit logger from the provided config. The
// logger writes JSON entries to cfg.Filepath and does log rotation. If
// cfg.Filepath is not set, a no-op logger will be returned.
func NewFromConfig(cfg Config) *Logger {
	if cfg.Filepath == "" {
		return &Logger{}
	}

	w := rotatingWriter(cfg)

	return New(w, slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{})))
}

// Close releases any resources held by the logger.
func (l *Logger) Close() error {
	if l.w == nil {
		return nil
	}
	return l.w.Close()
}

// Log logs an audit event.
func (l *Logger) Log(ctx context.Context, ev *Event) {
	if l.l == nil {
		return
	}
	l.l.Log(ctx, ev.Level, ev.Msg, ev.args()...)
}

// rotatingWriter creates a new io.WriteCloser that writes to a log file that is
// rotated after reaching MaxSize.
func rotatingWriter(cfg Config) io.WriteCloser {
	w := &lumberjack.Logger{
		Filename: cfg.Filepath,
		MaxSize:  cfg.MaxSize,
		Compress: cfg.Compress,
	}

	return w
}

// Event represents a audit log event.
type Event struct {
	Level      slog.Level
	Msg        string
	Type       string
	ResourceID string
	User       string
}

// args returns a slice of key/value pairs to be written to the audit log.
func (e Event) args() []any {
	return []any{
		"type", e.Type,
		"resourceID", e.ResourceID,
		"user", e.User,
	}
}
