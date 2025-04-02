package config_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/a3m"
	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/temporal"
)

const testConfig = `# Config
debug = true
debugListen = "127.0.0.1:9001"

[temporal]
address = "host:port"

[storage]
defaultPermanentLocationId = "f2cc963f-c14d-4eaa-b950-bd207189a1f1"

[api.auth.oidc.abac]
rolesMapping = '{"admin": ["*"], "operator": ["ingest:sips:list", "ingest:sips:workflows:list", "ingest:sips:read", "ingest:sips:upload"], "readonly": ["ingest:sips:list", "ingest:sips:workflows:list", "ingest:sips:read"]}'
`

func TestConfig(t *testing.T) {
	t.Parallel()

	t.Run("Loads toml configs", func(t *testing.T) {
		t.Parallel()

		tmpDir := fs.NewDir(
			t, "",
			fs.WithFile(
				"enduro-config.toml",
				testConfig,
			),
		)
		configFile := tmpDir.Join("enduro-config.toml")

		var c config.Configuration
		found, configFileUsed, err := config.Read(&c, configFile)

		assert.NilError(t, err)
		assert.Equal(t, found, true)
		assert.Equal(t, configFileUsed, configFile)
		assert.Equal(t, c.Temporal.Address, "host:port")

		// Test that a UUID config is decoded correctly.
		assert.Equal(t, c.Storage.DefaultPermanentLocationID, uuid.MustParse("f2cc963f-c14d-4eaa-b950-bd207189a1f1"))

		// Test that a map[string][]string config is decoded correctly.
		assert.DeepEqual(t, c.API.Auth.OIDC.ABAC.RolesMapping, map[string][]string{
			"admin": {"*"},
			"operator": {
				"ingest:sips:list",
				"ingest:sips:workflows:list",
				"ingest:sips:read",
				"ingest:sips:upload",
			},
			"readonly": {"ingest:sips:list", "ingest:sips:workflows:list", "ingest:sips:read"},
		})
	})

	t.Run("Sets default configs", func(t *testing.T) {
		t.Parallel()

		var c config.Configuration
		found, configFileUsed, err := config.Read(&c, "")

		assert.NilError(t, err)
		assert.Equal(t, found, false)
		assert.Equal(t, configFileUsed, "")

		// Zero value defaults.
		assert.Equal(t, c.Verbosity, 0)
		assert.Equal(t, c.Debug, false)
		assert.Equal(t, c.Database.DSN, "")

		// Valued defaults.
		assert.Equal(t, c.A3m.Capacity, 1)
		assert.Equal(t, c.A3m.Processing, a3m.ProcessingDefault)
		assert.Equal(t, c.AM.Capacity, 1)
		assert.Equal(t, c.AM.PollInterval, 10*time.Second)
		assert.Equal(t, c.AM.ZipPIP, false)
		assert.Equal(t, c.API.Listen, "127.0.0.1:9000")
		assert.Equal(t, c.BagIt.ChecksumAlgorithm, "sha512")
		assert.Equal(t, c.DebugListen, "127.0.0.1:9001")
		assert.Equal(t, c.Preservation.TaskQueue, temporal.A3mWorkerTaskQueue)
		assert.Equal(t, c.Storage.TaskQueue, temporal.GlobalTaskQueue)
		assert.Equal(t, c.Temporal.TaskQueue, temporal.GlobalTaskQueue)
		assert.Equal(t, c.ValidatePREMIS.Enabled, false)
		assert.Equal(t, c.ValidatePREMIS.XSDPath, "")
	})
}
