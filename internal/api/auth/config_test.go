package auth_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	t.Run("Passes validation", func(t *testing.T) {
		t.Parallel()

		cfg := &auth.Config{
			Enabled: false,
		}

		err := cfg.Validate()
		assert.NilError(t, err)
	})

	t.Run("Fails validation", func(t *testing.T) {
		t.Parallel()

		cfg := &auth.Config{
			Enabled: true,
		}

		err := cfg.Validate()
		assert.Error(t, err, "Missing OIDC configuration with API auth. enabled.")
	})
}
