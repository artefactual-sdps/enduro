package api_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"

	"github.com/artefactual-sdps/enduro/internal/api"
)

func TestNewLogger(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		cfg            api.LogConfig
		want           map[string]string
		loggerDisabled bool
		wantErr        string
	}

	for _, tc := range []testCase{
		{
			name: "logs to a file",
			cfg: api.LogConfig{
				Path: filepath.Join(t.TempDir(), "test.log"),
			},
			want: map[string]string{
				"level":        "INFO",
				"msg":          "Test log message",
				"service.name": "enduro.api",
			},
		},
		{
			name: "logs at debug level",
			cfg: api.LogConfig{
				Path:  filepath.Join(t.TempDir(), "test.log"),
				Level: slog.LevelDebug,
			},
			want: map[string]string{
				"level":        "DEBUG",
				"msg":          "Test log message",
				"service.name": "enduro.api",
			},
		},
		{
			name:           "discards log messages when log path is empty",
			cfg:            api.LogConfig{Path: ""},
			loggerDisabled: true,
		},
		{
			name: "logs to stderr",
			cfg:  api.LogConfig{Path: "stderr"},
		},
		{
			name: "logs to stdout",
			cfg:  api.LogConfig{Path: "stdout"},
		},
		{
			name: "errors on invalid log path",
			cfg: api.LogConfig{
				Path: filepath.Join(t.TempDir(), "nonexistent", "test.log"),
			},
			wantErr: "no such file or directory",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger, err := api.NewFileLogger(tc.cfg)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			assert.NilError(t, err)

			logger = logger.WithName("enduro").WithName("api").WithName("")

			enabled := logger.Enabled(context.Background(), tc.cfg.Level)
			if tc.loggerDisabled {
				assert.Assert(t, !enabled, "logger should be disabled")
			} else {
				assert.Assert(t, enabled, "logger should be enabled")
			}

			// If the log path is empty, stdout or stderr, we can't read the log
			// file to verify its contents.
			if tc.cfg.Path == "" || tc.cfg.Path == "stdout" || tc.cfg.Path == "stderr" {
				assert.NilError(t, logger.Close())
				return
			}

			logger.Log(context.Background(), tc.cfg.Level, "Test log message")

			data, err := os.ReadFile(tc.cfg.Path)
			assert.NilError(t, err)

			var msg map[string]any
			err = json.Unmarshal(data, &msg)
			assert.NilError(t, err)

			if tc.want != nil {
				for k, v := range tc.want {
					assert.Equal(t, msg[k], v)
				}
			}

			assert.Equal(t, strings.Count(string(data), "service.name"), 1)
			assert.NilError(t, logger.Close())
		})
	}
}

func TestNewLoggerWithDebug(t *testing.T) {
	t.Parallel()

	cfg := api.LogConfig{
		Path:   filepath.Join(t.TempDir(), "test.log"),
		Format: api.LogFormatText,
	}

	logger, err := api.NewFileLogger(cfg)
	assert.NilError(t, err)
	assert.Assert(t, logger != nil)
	logger = logger.WithName("enduro.api.test")

	logger.Info("Test log message")

	data, err := os.ReadFile(cfg.Path)
	assert.NilError(t, err)

	s := string(data)
	assert.Assert(t, cmp.Contains(s, `level=INFO`))
	assert.Assert(t, cmp.Contains(s, `msg="Test log message"`))
	assert.Assert(t, cmp.Contains(s, `service.name=enduro.api.test`))

	assert.NilError(t, logger.Close())
}

func TestLogToAClosedFile(t *testing.T) {
	t.Parallel()

	cfg := api.LogConfig{
		Path: filepath.Join(t.TempDir(), "test.log"),
	}

	logger, err := api.NewFileLogger(cfg)

	assert.NilError(t, err)
	assert.NilError(t, logger.Close())

	// Logging to a closed file should not panic or return an error.
	logger.Log(context.Background(), slog.LevelInfo, "Test log message")
}
