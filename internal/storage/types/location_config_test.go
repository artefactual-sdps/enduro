package types_test

import (
	"context"
	"encoding/json"
	"testing"

	_ "gocloud.dev/blob/memblob"
	"gocloud.dev/gcerrors"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func TestLocationConfigEncoding(t *testing.T) {
	t.Parallel()

	type test struct {
		config      types.LocationConfig
		want        string
		wantErr     string
		wantInvalid bool
	}
	for name, tt := range map[string]test{
		"Encodes valid S3 config": {
			config: types.LocationConfig{
				Value: &types.S3Config{
					Bucket: "perma-aips-1",
					Region: "eu-west-1",
				},
			},
			want: `{"s3":{"bucket":"perma-aips-1","region":"eu-west-1"}}`,
		},
		"Encodes valid SFTP config": {
			config: types.LocationConfig{
				Value: &types.SFTPConfig{
					Address:   "sftp:22",
					Username:  "user",
					Password:  "secret",
					Directory: "upload",
				},
			},
			want: `{"sftp":{"address":"sftp:22","username":"user","password":"secret","directory":"upload"}}`,
		},
		"Encodes valid URL config": {
			config: types.LocationConfig{
				Value: &types.URLConfig{
					URL: "mem://",
				},
			},
			want: `{"url":{"url":"mem://"}}`,
		},
		"Encodes valid SS config": {
			config: types.LocationConfig{
				Value: &types.SSConfig{
					URL:      "http://127.0.0.1:62081",
					Username: "test",
					APIKey:   "secret",
				},
			},
			want: `{"ss":{"url":"http://127.0.0.1:62081","username":"test","api_key":"secret"}}`,
		},
		"Rejects invalid S3 config": {
			config: types.LocationConfig{
				Value: &types.S3Config{
					Bucket: "perma-aips-1",
				},
			},
			want:        `{"s3":{"bucket":"perma-aips-1","region":""}}`,
			wantInvalid: true,
		},
		"Rejects invalid SFTP config": {
			config: types.LocationConfig{
				Value: &types.SFTPConfig{},
			},
			want:        `{"sftp":{"address":"","username":"","password":"","directory":""}}`,
			wantInvalid: true,
		},
		"Rejects invalid URL config": {
			config: types.LocationConfig{
				Value: &types.URLConfig{},
			},
			want:        `{"url":{"url":""}}`,
			wantInvalid: true,
		},
		"Rejects invalid SS config": {
			config: types.LocationConfig{
				Value: &types.SSConfig{},
			},
			want:        `{"ss":{"url":"","username":"","api_key":""}}`,
			wantInvalid: true,
		},
		"Rejects invalid config": {
			config:  types.LocationConfig{},
			wantErr: "json: error calling MarshalJSON for type types.LocationConfig: unsupported config type: <nil>",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			blob, err := json.Marshal(tt.config)

			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				assert.Assert(t, blob == nil)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, string(blob), tt.want)
			assert.Equal(t, tt.config.Value.Valid(), !tt.wantInvalid)
		})
	}
}

func TestLocationConfigDecoding(t *testing.T) {
	t.Parallel()

	type test struct {
		blob    string
		want    types.LocationConfig
		extra   func(c types.LocationConfig)
		wantErr string
	}
	for name, tt := range map[string]test{
		"Decodes S3 config": {
			blob: `{"s3":{"bucket":"perma-aips-1","region":"eu-west-1"}}`,
			want: types.LocationConfig{
				Value: &types.S3Config{
					Bucket: "perma-aips-1",
					Region: "eu-west-1",
				},
			},
			extra: func(c types.LocationConfig) {
				b, err := c.Value.OpenBucket(context.Background())
				assert.NilError(t, err)
				defer b.Close()

				// Use it even though we know it's not accessible.
				_, err = b.IsAccessible(context.Background())
				assert.ErrorContains(t, err, "static credentials are empty")
				assert.Equal(t, gcerrors.Code(err), gcerrors.Unknown)
			},
		},
		"Decodes SFTP config": {
			blob: `{"sftp":{"address":"sftp:22","username":"user","password":"secret","directory":"upload"}}`,
			want: types.LocationConfig{
				Value: &types.SFTPConfig{
					Address:   "sftp:22",
					Username:  "user",
					Password:  "secret",
					Directory: "upload",
				},
			},
		},
		"Decodes URL config": {
			blob: `{"url": {"url": "mem://test-bucket"}}`,
			want: types.LocationConfig{
				Value: &types.URLConfig{
					URL: "mem://test-bucket",
				},
			},
			extra: func(c types.LocationConfig) {
				b, err := c.Value.OpenBucket(context.Background())
				assert.NilError(t, err)
				defer b.Close()

				y, err := b.IsAccessible(context.Background())
				assert.NilError(t, err)
				assert.Equal(t, y, true)
			},
		},
		"Decodes SS config": {
			blob: `{"ss":{"url":"http://127.0.0.1:62081","username":"test","api_key":"secret"}}`,
			want: types.LocationConfig{
				Value: &types.SSConfig{
					URL:      "http://127.0.0.1:62081",
					Username: "test",
					APIKey:   "secret",
				},
			},
			extra: func(c types.LocationConfig) {
				b, err := c.Value.OpenBucket(context.Background())
				assert.NilError(t, err)
				b.Close()
			},
		},
		"Rejects URL config if invalid": {
			blob: `{"url": {"url": "foo://test-bucket"}}`,
			want: types.LocationConfig{
				Value: &types.URLConfig{
					URL: "foo://test-bucket",
				},
			},
			extra: func(c types.LocationConfig) {
				_, err := c.Value.OpenBucket(context.Background())
				assert.Error(t, err, `open bucket by URL: open blob.Bucket: no driver registered for "foo" for URL "foo://test-bucket"; available schemes: mem, s3, sftp`)
			},
		},
		"Rejects unknown config": {
			blob:    `{"xxxxxx":{}}`,
			wantErr: "undefined configuration document",
		},
		"Rejects unexpected JSON": {
			blob:    `[1, 2, 3]`,
			wantErr: "undefined configuration format",
		},
		"Rejects multiple configs (1)": {
			blob: `{
				"s3":{"bucket":"perma-aips-1","region":"eu-west-1"},
				"sftp":{"address":"sftp:22","username":"user","password":"secret","directory":"upload"}
			}`,
			wantErr: "multiple config values have been assigned",
		},
		"Rejects multiple configs (2)": {
			blob: `{
				"sftp":{"address":"sftp:22","username":"user","password":"secret","directory":"upload"},
				"s3":{"bucket":"perma-aips-1","region":"eu-west-1"}
			}`,
			wantErr: "multiple config values have been assigned",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := types.LocationConfig{}
			err := json.Unmarshal([]byte(tt.blob), &cfg)

			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				assert.DeepEqual(t, cfg, types.LocationConfig{})
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, cfg, tt.want)
			assert.Equal(t, cfg.Value.Valid(), true)

			if tt.extra != nil {
				tt.extra(cfg)
			}
		})
	}
}
