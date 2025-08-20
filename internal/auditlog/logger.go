package auditlog

import (
	"log/slog"

	"gopkg.in/natefinch/lumberjack.v2"
)

// NewFromConfig creates a new *slog.Logger from the specified configuration
// that writes logs in JSON format to cfg.Filepath and does log rotation. If
// cfg.Filepath is not set, a nil Logger is returned.
func NewFromConfig(cfg Config) *slog.Logger {
	if cfg.Filepath == "" {
		return nil
	}

	// Use lumberjack.Logger for log rotation.
	l := &lumberjack.Logger{
		Filename: cfg.Filepath,
		MaxSize:  cfg.MaxSize,
		Compress: cfg.Compress,
	}

	return slog.New(slog.NewJSONHandler(l, &slog.HandlerOptions{
		Level: slog.Level(cfg.Verbosity),
	}))
}
