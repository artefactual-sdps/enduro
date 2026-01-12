package batch_test

import (
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/config"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	const testConfig = `# Config
[batch.poststorage]
namespace = "default"
taskQueue = "batch-post-storage"
workflowName = "batch-post-storage-workflow"
`

	t.Run("Decodes batch.poststorage TOML configs", func(t *testing.T) {
		t.Parallel()

		tmpDir := fs.NewDir(t, "", fs.WithFile("config.toml", testConfig))
		configFile := tmpDir.Join("config.toml")

		var c config.Configuration
		_, _, err := config.Read(&c, configFile)

		assert.NilError(t, err)

		// Test that batch.poststorage section is decoded correctly.
		assert.Equal(t, c.Batch.Poststorage.Namespace, "default")
		assert.Equal(t, c.Batch.Poststorage.TaskQueue, "batch-post-storage")
		assert.Equal(t, c.Batch.Poststorage.WorkflowName, "batch-post-storage-workflow")
	})

	t.Run("Handles missing batch.poststorage section", func(t *testing.T) {
		t.Parallel()

		tmpDir := fs.NewDir(t, "", fs.WithFile("config.toml", ""))
		configFile := tmpDir.Join("config.toml")

		var c config.Configuration
		_, _, err := config.Read(&c, configFile)

		assert.NilError(t, err)
		assert.Assert(t, c.Batch.Poststorage == nil)
	})
}
