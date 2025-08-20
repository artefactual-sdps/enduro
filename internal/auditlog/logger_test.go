package auditlog_test

import (
	"encoding/json"
	"os"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	auditlog "github.com/artefactual-sdps/enduro/internal/auditlog"
)

type entry struct {
	// Ignore the time field because it's non-deterministic.

	Level string `json:"level"`
	Msg   string `json:"msg"`
	Key   string `json:"key"`
}

func TestNewFromConfig(t *testing.T) {
	t.Parallel()

	t.Run("creates a logger with the given config", func(t *testing.T) {
		t.Parallel()

		d := fs.NewDir(t, "enduro-test")
		cfg := auditlog.Config{
			Filepath: d.Join("enduro_audit.log"),
			MaxSize:  10,
		}

		logger := auditlog.NewFromConfig(cfg)
		if logger == nil {
			t.Fatal("Expected logger to be created")
		}

		logger.Info("Test audit log entry", "key", "value")

		r, err := os.Open(d.Join("enduro_audit.log"))
		if err != nil {
			if os.IsNotExist(err) {
				t.Fatalf("Missing expected file: %s", d.Join("enduro_audit.log"))
			}
			t.Fatal(err)
		}
		defer r.Close()

		var e entry
		if err := json.NewDecoder(r).Decode(&e); err != nil {
			t.Fatal(err)
		}

		assert.DeepEqual(t, e, entry{
			Level: "INFO",
			Msg:   "Test audit log entry",
			Key:   "value",
		})
	})

	t.Run("returns nil if Filepath is empty", func(t *testing.T) {
		t.Parallel()

		logger := auditlog.NewFromConfig(auditlog.Config{})
		if logger != nil {
			t.Fatal("Expected logger to be nil")
		}
	})
}
