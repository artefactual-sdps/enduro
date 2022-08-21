package types_test

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func TestLocationConfig(t *testing.T) {
	t.Parallel()

	t.Run("Encoding", func(t *testing.T) {
		t.Parallel()

		// Invalid config.
		cfg := types.LocationConfig{}
		blob, err := json.Marshal(cfg)
		assert.DeepEqual(t, blob, []byte(nil))
		assert.Error(t, err, "json: error calling MarshalJSON for type types.LocationConfig: unsupported config type: <nil>")

		// Valid S3 config.
		cfg = types.LocationConfig{
			Value: &types.S3Config{
				Bucket: "perma-aips-1",
				Region: "eu-west-1",
			},
		}
		blob, err = json.Marshal(cfg)
		assert.NilError(t, err)
		assert.DeepEqual(t, string(blob), `{"s3":{"bucket":"perma-aips-1","region":"eu-west-1"}}`)
		assert.Equal(t, cfg.Value.Valid(), true)

		// Invalid S3 config.
		cfg = types.LocationConfig{
			Value: &types.S3Config{
				Bucket: "perma-aips-1",
			},
		}
		blob, err = json.Marshal(cfg)
		assert.NilError(t, err)
		assert.DeepEqual(t, string(blob), `{"s3":{"bucket":"perma-aips-1","region":""}}`)
		assert.Equal(t, cfg.Value.Valid(), false)
	})

	t.Run("Decoding", func(t *testing.T) {
		t.Parallel()

		// S3 config.
		blob := []byte(`{"s3":{"bucket":"perma-aips-1","region":"eu-west-1"}}`)
		cfg := types.LocationConfig{}
		err := json.Unmarshal(blob, &cfg)
		assert.NilError(t, err)
		assert.DeepEqual(t, cfg, types.LocationConfig{
			Value: &types.S3Config{
				Bucket: "perma-aips-1",
				Region: "eu-west-1",
			},
		})

		// Unknown config.
		blob = []byte(`{"xxxxxx":{"bucket":"perma-aips-1","region":"eu-west-1"}}`)
		cfg = types.LocationConfig{}
		err = json.Unmarshal(blob, &cfg)
		assert.Error(t, err, "undefined configuration document")
		assert.DeepEqual(t, cfg, types.LocationConfig{})
	})
}
