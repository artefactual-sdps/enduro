package upload_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/upload"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	t.Run("Returns error if config is invalid", func(t *testing.T) {
		t.Parallel()

		c := upload.Config{
			URL:    "s3blob://my-bucket",
			Bucket: "my-bucket",
			Region: "planet-earth",
		}
		err := c.Validate()
		assert.ErrorContains(t, err, "URL and rest of the [upload] configuration options are mutually exclusive")
	})

	t.Run("Validates if only URL is provided", func(t *testing.T) {
		t.Parallel()

		c := upload.Config{
			URL: "s3blob://my-bucket",
		}
		err := c.Validate()
		assert.NilError(t, err)
	})

	t.Run("Validates if only bucket options are provided", func(t *testing.T) {
		t.Parallel()

		c := upload.Config{
			Bucket: "my-bucket",
			Region: "planet-earth",
		}
		err := c.Validate()
		assert.NilError(t, err)
	})
}
