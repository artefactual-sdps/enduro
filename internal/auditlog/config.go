package auditlog

import (
	"io"
	"log/slog"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	// Filepath specifies the location of the audit log file. If Filepath is
	// not set, audit logging will be disabled.
	Filepath string

	// MaxSize sets the maximum size of the audit log file in megabytes before
	// it is rotated (default: 100 MB).
	MaxSize int

	// Compress determines if the rotated log files are compressed using gzip
	// (default: false).
	Compress bool

	// Verbosity sets the minimum log level that will be logged. Valid levels
	// are: -4 = DEBUG, 0 = INFO (default), 4 = WARN, 8 = ERROR.
	Verbosity int
}

// NewFromConfig creates a new Auditlog instance from the given config, using
// lumberjack for log rotation.  If the Filepath in the config is not set,
// a nil Auditlog will be returned.
func NewFromConfig[T any](cfg Config, h EventHandler[T]) *Auditlog[T] {
	if cfg.Filepath == "" {
		return nil
	}

	logger, w := loggerFromConfig(cfg)
	return &Auditlog[T]{
		logger:  logger,
		w:       w,
		handler: h,
		stopCh:  make(chan struct{}),
	}
}

func loggerFromConfig(cfg Config) (*slog.Logger, io.WriteCloser) {
	w := &lumberjack.Logger{
		Filename: cfg.Filepath,
		MaxSize:  cfg.MaxSize,
		Compress: cfg.Compress,
	}

	logger := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.Level(cfg.Verbosity),
	}))

	return logger, w
}
