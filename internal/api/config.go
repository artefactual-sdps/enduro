package api

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/artefactual-sdps/enduro/internal/auth"
)

type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)

type Config struct {
	// Listen defines the TCP address in host:port form on which the API server
	// will listen for incoming requests. The default value is "127.0.0.1:9000".
	Listen string

	// Auth defines the authentication configuration for the API server.
	Auth auth.Config

	// CORSOrigin defines the allowed origin for Cross-Origin Resource Sharing
	// (CORS) requests, by setting the `Access-Control-Allow-Origin` header for
	// API responses. If not set, CORSOrigin will default to the value of
	// Listen, disallowing all cross-origin requests. See
	// https://pkg.go.dev/goa.design/plugins/cors/dsl?utm_source=godoc#Origin
	// for detailed information on the allowed CORSOrigin values.
	CORSOrigin string

	// Log defines the logging configuration for the API server.
	Log LogConfig
}

func (c *Config) Validate() error {
	c.setDefaults()

	return errors.Join(
		c.Auth.Validate(),
		c.Log.Validate(),
	)
}

func (c *Config) setDefaults() {
	// Default to the API URI to disallow all cross-origin requests.
	if c.CORSOrigin == "" {
		c.CORSOrigin = c.Listen
	}

	c.Log.setDefaults()
}

type LogConfig struct {
	// Path defines the path to the log file.
	//
	// If Path is not set, logging is disabled.
	// If Path is set to "stdout" or "stderr" (case insensitive), the logs
	// will be written to the standard output or standard error streams,
	// respectively.
	// If Path is set to a file path, the logs will be written to that file.
	Path string

	// Level defines the verbosity of the API logs. The default Level is
	// slog.LevelInfo.
	//
	// slog.LevelDebug - log all responses (incl. OPTIONS)
	// slog.LevelInfo  - log responses (excl. OPTIONS)
	// slog.LevelWarn  - log 4xx and 5xx responses only (except for 429)
	// slog.LevelError - log 5xx responses only
	Level slog.Level

	// Format sets the encoding of log messages to support different reader
	// contexts. Valid option are LogFormatJSON and LogFormatText.
	//
	// The default LogFormatJSON encodes messages as line separated JSON, and is
	// better for analytics and log aggregation.
	// LogFormatText is intended for human readability and returns log data
	// as simple "key=value" pairs.
	Format LogFormat
}

func (l *LogConfig) Validate() error {
	l.setDefaults()

	if l.Format != LogFormatJSON && l.Format != LogFormatText {
		return fmt.Errorf(
			"unsupported log format: %q, supported formats are %q, %q",
			l.Format,
			LogFormatJSON,
			LogFormatText,
		)
	}

	return nil
}

func (l *LogConfig) setDefaults() {
	if l.Format == "" {
		l.Format = LogFormatJSON
	}
}
