package types_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "gocloud.dev/blob/memblob"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

var (
	handler = http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})
	srv     = httptest.NewServer(handler)
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
		assert.Equal(t, string(blob), `{"s3":{"bucket":"perma-aips-1","region":"eu-west-1"}}`)
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
		assert.Equal(t, string(blob), `{"sftp":{"address":"sftp:22","username":"user","password":"secret","directory":"upload"}}`)
		assert.Equal(t, cfg.Value.Valid(), true)

		// Valid SS config.
		cfg = types.LocationConfig{
			Value: &types.SSConfig{
				URL:    srv.URL,
				APIKey: "secret",
			},
		}
		testSSConfig := fmt.Sprintf(`{"ss":{"url":"%s","username":"","api_key":"secret"}}`, srv.URL)
		blob, err = json.Marshal(cfg)
		assert.NilError(t, err)
		assert.Equal(t, string(blob), testSSConfig)
		assert.Equal(t, cfg.Value.Valid(), true)

		// Invalid S3 config.
		cfg = types.LocationConfig{
			Value: &types.S3Config{
				Bucket: "perma-aips-1",
			},
		}
		blob, err = json.Marshal(cfg)
		assert.NilError(t, err)
		assert.Equal(t, string(blob), `{"s3":{"bucket":"perma-aips-1","region":""}}`)
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
		assert.Equal(t, string(blob), `{"sftp":{"address":"sftp:22","username":"user","password":"","directory":"upload"}}`)
		assert.Equal(t, cfg.Value.Valid(), false)

		// Invalid SS config.
		cfg = types.LocationConfig{
			Value: &types.SSConfig{
				URL: srv.URL,
			},
		}
		blob, err = json.Marshal(cfg)
		assert.NilError(t, err)
		testSSConfig = fmt.Sprintf(`{"ss":{"url":"%s","username":"","api_key":""}}`, srv.URL)
		assert.Equal(t, string(blob), testSSConfig)
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

		// SS config.
		cfg = types.LocationConfig{
			Value: &types.SSConfig{
				URL:    srv.URL,
				APIKey: "secret",
			},
		}
		testSSConfig := fmt.Sprintf(`{"ss":{"url":"%s","username":"","api_key":"secret"}}`, srv.URL)
		blob = []byte(testSSConfig)
		cfg = types.LocationConfig{}
		err = json.Unmarshal(blob, &cfg)
		assert.NilError(t, err)
		assert.DeepEqual(t, cfg, types.LocationConfig{
			Value: &types.SSConfig{
				URL:    srv.URL,
				APIKey: "secret",
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

func TestURLConfig(t *testing.T) {
	t.Parallel()

	t.Run("Encodes a URL config", func(t *testing.T) {
		t.Parallel()

		cfg := types.LocationConfig{
			Value: &types.URLConfig{
				URL: "mem:///test-bucket",
			},
		}
		blob, err := json.Marshal(cfg)

		assert.NilError(t, err)
		assert.Equal(t, string(blob), `{"url":{"url":"mem:///test-bucket"}}`)
		assert.Equal(t, cfg.Value.Valid(), true)
	})

	t.Run("Decodes a URL config", func(t *testing.T) {
		t.Parallel()

		blob := []byte(`{"url":{"url":"mem:///test-bucket"}}`)
		cfg := types.LocationConfig{}
		err := json.Unmarshal(blob, &cfg)

		assert.NilError(t, err)
		assert.DeepEqual(t, cfg, types.LocationConfig{
			Value: &types.URLConfig{
				URL: "mem:///test-bucket",
			},
		})
	})

	t.Run("Opens a URL Config bucket", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c := types.URLConfig{URL: "mem:///test-bucket"}

		b, err := c.OpenBucket(ctx)
		assert.NilError(t, err)
		defer b.Close()

		y, err := b.IsAccessible(ctx)
		assert.NilError(t, err)
		assert.Equal(t, y, true)
	})

	t.Run("Errors if URL Config is invalid", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c := types.URLConfig{URL: "foo:///test-bucket"}
		_, err := c.OpenBucket(ctx)
		assert.ErrorContains(t, err,
			`open bucket by URL: open blob.Bucket: no driver registered for "foo"`,
		)
	})

	t.Run("Opens a SS Config bucket", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c := types.URLConfig{URL: "mem:///test-bucket"}

		b, err := c.OpenBucket(ctx)
		assert.NilError(t, err)
		defer b.Close()

		y, err := b.IsAccessible(ctx)
		assert.NilError(t, err)
		assert.Equal(t, y, true)
	})

	t.Run("Errors if SS Config is invalid", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c := types.URLConfig{URL: "foo:///test-bucket"}
		_, err := c.OpenBucket(ctx)
		assert.ErrorContains(t, err,
			`open bucket by URL: open blob.Bucket: no driver registered for "foo"`,
		)
	})
}
func TestSSConfig(t *testing.T) {
	t.Parallel()

	t.Run("Encodes a SS config", func(t *testing.T) {
		t.Parallel()

		cfg := types.LocationConfig{
			Value: &types.SSConfig{
				URL:      srv.URL,
				Username: "test@example.com",
				APIKey:   "api_key_example",
			},
		}
		blob, err := json.Marshal(cfg)

		str := fmt.Sprintf(`{"ss":{"url":"%s","username":"test@example.com","api_key":"api_key_example"}}`, srv.URL)
		assert.NilError(t, err)
		assert.Equal(t, string(blob), str)
		assert.Equal(t, cfg.Value.Valid(), true)
	})

	t.Run("Decodes a SS config", func(t *testing.T) {
		t.Parallel()

		str := fmt.Sprintf(`{"ss":{"url":"%s","api_key":"api_key_example"}}`, srv.URL)
		blob := []byte(str)
		cfg := types.LocationConfig{}
		err := json.Unmarshal(blob, &cfg)

		assert.NilError(t, err)
		assert.DeepEqual(t, cfg, types.LocationConfig{
			Value: &types.SSConfig{
				URL:    srv.URL,
				APIKey: "api_key_example",
			},
		})
	})

	t.Run("Opens a SS Config bucket", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c := types.SSConfig{URL: srv.URL, APIKey: "api_key_example"}

		b, err := c.OpenBucket(ctx)
		assert.NilError(t, err)
		defer b.Close()
	})
}
