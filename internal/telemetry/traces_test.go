package telemetry

import (
	"testing"

	"github.com/go-logr/logr"
	"gotest.tools/v3/assert"
)

func TestTracerProvider(t *testing.T) {
	t.Parallel()

	t.Run("Invalid address fails fast", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			Traces: TracesConfig{
				Enabled: true,
				Address: "invalid-endpoint", // missing port
			},
		}

		_, _, err := TracerProvider(t.Context(), logr.Discard(), cfg, "enduro", "dev")
		assert.Assert(t, err != nil, "expected error for invalid address")
	})

	t.Run("Valid address returns provider and shutdown", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			Traces: TracesConfig{
				Enabled: true,
				Address: "localhost:4317",
			},
		}

		tp, shutdown, err := TracerProvider(t.Context(), logr.Discard(), cfg, "enduro", "dev")
		assert.NilError(t, err)
		assert.Assert(t, tp != nil, "expected tracer provider")
		assert.Assert(t, shutdown != nil, "expected shutdown func")
	})
}
