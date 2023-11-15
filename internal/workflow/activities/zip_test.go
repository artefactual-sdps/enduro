package activities_test

import (
	"os"
	"testing"

	"github.com/go-logr/logr"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

func TestZipActivity(t *testing.T) {
	t.Run("Zips a directory", func(t *testing.T) {
		t.Parallel()

		transferName := "my_transfer"
		contents := tfs.WithDir(transferName,
			tfs.WithDir("subdir",
				tfs.WithFile("abc.txt", "Testing A-B-C"),
			),
			tfs.WithFile("123.txt", "Testing 1-2-3"),
		)
		td := tfs.NewDir(t, "enduro-zip-test", contents)

		ts := &temporalsdk_testsuite.WorkflowTestSuite{}
		env := ts.NewTestActivityEnvironment()
		env.RegisterActivityWithOptions(
			activities.NewZipActivity(logr.Discard()).Execute,
			temporalsdk_activity.RegisterOptions{
				Name: activities.ZipActivityName,
			},
		)

		fut, err := env.ExecuteActivity(activities.ZipActivityName,
			activities.ZipActivityParams{SourceDir: td.Join(transferName)},
		)
		assert.NilError(t, err)

		var res activities.ZipActivityResult
		_ = fut.Get(&res)
		assert.DeepEqual(t, res, activities.ZipActivityResult{Path: td.Join(transferName + ".zip")})

		// Confirm that a zip file was created with some contents.
		i, err := os.Lstat(td.Join(transferName + ".zip"))
		assert.NilError(t, err)
		assert.Assert(t, i.Size() > int64(0))
	})
}
