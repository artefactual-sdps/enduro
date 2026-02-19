package config_test

import (
	"testing"
	"time"

	"github.com/artefactual-sdps/temporal-activities/bagcreate"
	"github.com/google/uuid"
	"go.artefactual.dev/tools/bucket"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/a3m"
	"github.com/artefactual-sdps/enduro/internal/am"
	"github.com/artefactual-sdps/enduro/internal/api"
	"github.com/artefactual-sdps/enduro/internal/api/auth"
	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/pres"
	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/temporal"
)

const testConfig = `# Config
debug = true
debugListen = "127.0.0.1:9001"

[temporal]
address = "host:port"

[ingest.storage]
defaultPermanentLocationId = "f2cc963f-c14d-4eaa-b950-bd207189a1f1"

[api.auth.oidc.abac]
rolesMapping = '{"admin": ["*"], "operator": ["ingest:sips:list", "ingest:sips:workflows:list", "ingest:sips:read", "ingest:sips:upload"], "readonly": ["ingest:sips:list", "ingest:sips:workflows:list", "ingest:sips:read"]}'
`

func TestConfigRead(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		config  string
		want    config.Configuration
		wantErr string
	}
	for _, tc := range []test{
		{
			name:   "Loads TOML configs",
			config: testConfig,
			want: config.Configuration{
				Debug:       true,
				DebugListen: "127.0.0.1:9001",
				A3m: a3m.Config{
					Processing: a3m.ProcessingDefault,
				},
				AM: am.Config{
					Capacity:     20,
					PollInterval: 10 * time.Second,
				},
				API: api.Config{
					Listen: "127.0.0.1:9000",
					Auth: auth.Config{
						OIDC: &auth.OIDCConfig{
							ABAC: auth.OIDCABACConfig{
								RolesMapping: map[string][]string{
									"admin":    {"*"},
									"operator": {"ingest:sips:list", "ingest:sips:workflows:list", "ingest:sips:read", "ingest:sips:upload"},
									"readonly": {"ingest:sips:list", "ingest:sips:workflows:list", "ingest:sips:read"},
								},
							},
						},
					},
					CORSOrigin: "127.0.0.1:9000",
				},
				BagIt: bagcreate.Config{
					ChecksumAlgorithm: "sha512",
				},
				Ingest: ingest.Config{
					Storage: ingest.StorageConfig{
						DefaultPermanentLocationID: uuid.MustParse("f2cc963f-c14d-4eaa-b950-bd207189a1f1"),
					},
				},
				Preservation: pres.Config{
					TaskQueue: "a3m",
				},
				Storage: storage.Config{
					TaskQueue: "global",
				},
				Temporal: temporal.Config{
					Address:   "host:port",
					TaskQueue: "global",
				},
				Upload: ingest.UploadConfig{
					MaxSize: 4294967296,
				},
			},
		},
		{
			name: "Sets default values for missing config options",
			want: config.Configuration{
				DebugListen: "127.0.0.1:9001",
				A3m: a3m.Config{
					Processing: a3m.ProcessingDefault,
				},
				AM: am.Config{
					Capacity:     20,
					PollInterval: 10 * time.Second,
				},
				API: api.Config{
					Listen:     "127.0.0.1:9000",
					CORSOrigin: "127.0.0.1:9000",
				},
				BagIt: bagcreate.Config{
					ChecksumAlgorithm: "sha512",
				},
				Preservation: pres.Config{
					TaskQueue: "a3m",
				},
				Storage: storage.Config{
					TaskQueue: "global",
				},
				Temporal: temporal.Config{
					TaskQueue: "global",
				},
				Upload: ingest.UploadConfig{
					MaxSize: 4294967296,
				},
			},
		},
		{
			name:    "Returns error if config is invalid",
			config:  "debug = not-a-boolean",
			wantErr: "failed to read configuration file: While parsing config: toml: expected 'nan'",
		},
		{
			name: "Returns error if string to UUID hook fails",
			config: `[ingest.storage]
defaultPermanentLocationId = "not-a-uuid"`,
			wantErr: `failed to unmarshal configuration: 1 error(s) decoding:

* error decoding 'Ingest.Storage.DefaultPermanentLocationID': invalid UUID length: 10`,
		},
		{
			name: "Returns error if string to map hook fails",
			config: `[api.auth.oidc.abac]
rolesMapping = "not-a-json"`,
			wantErr: `failed to unmarshal configuration: 1 error(s) decoding:

* error decoding 'API.Auth.OIDC.ABAC.RolesMapping': invalid character 'o' in literal null (expecting 'u')`,
		},
		{
			name: "Returns error if validation fails",
			config: `[a3m.processing]
aipCompressionLevel = 10`,
			wantErr: "failed to validate the provided config: AipCompressionLevel: 10 is outside valid range (0 to 9)",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := fs.NewDir(t, "",
				fs.WithFile("enduro-config.toml", tc.config),
			)
			configFile := tmpDir.Join("enduro-config.toml")

			var c config.Configuration
			found, configFileUsed, err := config.Read(&c, configFile)
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, found, true)
			assert.Equal(t, configFileUsed, configFile)
			assert.DeepEqual(t, c, tc.want)
		})
	}
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		config  config.Configuration
		wantErr string
	}
	for _, tc := range []test{
		{
			name: "Returns error if bucket URL and other options are both provided",
			config: config.Configuration{
				InternalStorage: config.InternalStorageConfig{
					Bucket: bucket.Config{
						URL:    "s3blob://my-bucket",
						Bucket: "my-bucket",
						Region: "planet-earth",
					},
				},
			},
			wantErr: "the [internalStorage] URL option and the other configuration options are mutually exclusive",
		},
		{
			name: "Returns error if azure credentials are missing",
			config: config.Configuration{
				InternalStorage: config.InternalStorageConfig{
					Bucket: bucket.Config{
						URL: "azblob://my-bucket",
					},
				},
			},
			wantErr: "the [internalStorage] Azure credentials are undefined",
		},
		{
			name: "Validates if only URL is provided",
			config: config.Configuration{
				InternalStorage: config.InternalStorageConfig{
					Bucket: bucket.Config{
						URL: "s3blob://my-bucket",
					},
				},
			},
		},
		{
			name: "Validates if only bucket options are provided",
			config: config.Configuration{
				InternalStorage: config.InternalStorageConfig{
					Bucket: bucket.Config{
						Bucket: "my-bucket",
						Region: "planet-earth",
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.config.Validate()
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
		})
	}
}
