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

		// Valid SFTP config.
		cfg = types.LocationConfig{
			Value: &types.SFTPConfig{
				Address:   "sftp:22",
				Username:  "user",
				Password:  "secret",
				Directory: "upload",
			},
		}
		blob, err = json.Marshal(cfg)
		assert.NilError(t, err)
		assert.DeepEqual(t, string(blob), `{"sftp":{"address":"sftp:22","username":"user","password":"secret","directory":"upload"}}`)
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

		// Invalid SFTP config.
		cfg = types.LocationConfig{
			Value: &types.SFTPConfig{
				Address:   "sftp:22",
				Username:  "user",
				Directory: "upload",
			},
		}
		blob, err = json.Marshal(cfg)
		assert.NilError(t, err)
		assert.DeepEqual(t, string(blob), `{"sftp":{"address":"sftp:22","username":"user","password":"","directory":"upload"}}`)
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

		// SFTP Config
		blob = []byte(`{"sftp":{"address":"sftp:22","username":"user","password":"secret","directory":"upload"}}`)
		cfg = types.LocationConfig{}
		err = json.Unmarshal(blob, &cfg)
		assert.NilError(t, err)
		assert.DeepEqual(t, cfg, types.LocationConfig{
			Value: &types.SFTPConfig{
				Address:   "sftp:22",
				Username:  "user",
				Password:  "secret",
				Directory: "upload",
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
