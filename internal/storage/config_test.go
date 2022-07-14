package storage_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage"
)

func TestConfigValidator(t *testing.T) {
	t.Parallel()

	t.Run("Finds unnamed locations", func(t *testing.T) {
		t.Parallel()
		cfg := storage.Config{
			Locations: []storage.LocationConfig{
				{}, // Unnamed.
			},
		}
		err := cfg.Validate()

		assert.Error(t, err, "location name is undefined")
	})

	t.Run("Finds duplicated locations", func(t *testing.T) {
		t.Parallel()
		cfg := storage.Config{
			Locations: []storage.LocationConfig{
				{Name: "loc1"},
				{Name: "loc1"},
			},
		}
		err := cfg.Validate()

		assert.Error(t, err, "location with name loc1 already defined")
	})
}
