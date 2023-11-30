package config_test

import (
	"testing"
	"time"

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
`

func TestConfig(t *testing.T) {
	t.Run("Loads toml configs", func(t *testing.T) {
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
	})

	t.Run("Sets default configs", func(t *testing.T) {
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
		assert.Equal(t, c.API.Listen, "127.0.0.1:9000")
		assert.Equal(t, c.DebugListen, "127.0.0.1:9001")
		assert.Equal(t, c.Preservation.TaskQueue, temporal.A3mWorkerTaskQueue)
		assert.Equal(t, c.Storage.TaskQueue, temporal.GlobalTaskQueue)
		assert.Equal(t, c.Temporal.TaskQueue, temporal.GlobalTaskQueue)
	})
}
