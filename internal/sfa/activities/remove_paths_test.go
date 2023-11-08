package activities_test

import (
	"os"
	"path/filepath"
	"testing"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/sfa/activities"
)

func TestRemovePaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dir     *fs.Dir
		paths   []string
		want    activities.RemovePathsResult
		wantErr string
	}{
		{
			name:  "Succeeds with single path",
			dir:   fs.NewDir(t, "", fs.WithDir("folder")),
			paths: []string{"folder"},
			want:  activities.RemovePathsResult{},
		},
		{
			name:  "Succeeds with multiple paths",
			dir:   fs.NewDir(t, "", fs.WithDir("folder1"), fs.WithDir("folder2")),
			paths: []string{"folder1", "folder2"},
			want:  activities.RemovePathsResult{},
		},
		{
			name:    "Fails with single path",
			dir:     fs.NewDir(t, "", fs.WithDir("folder"), fs.WithMode(0o000)),
			paths:   []string{"folder"},
			wantErr: "error removing path",
		},
		{
			name:    "Fails with multiple paths",
			dir:     fs.NewDir(t, "", fs.WithDir("folder1"), fs.WithDir("folder2"), fs.WithMode(0o000)),
			paths:   []string{"folder1", "folder2"},
			wantErr: "error removing path",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var paths []string
			for _, path := range tt.paths {
				paths = append(paths, filepath.Join(tt.dir.Path(), path))
			}

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				activities.NewRemovePaths().Execute,
				temporalsdk_activity.RegisterOptions{Name: activities.RemovePathsName},
			)

			future, err := env.ExecuteActivity(
				activities.RemovePathsName,
				&activities.RemovePathsParams{Paths: paths},
			)

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("error is nil, expecting: %q", tt.wantErr)
				} else {
					assert.ErrorContains(t, err, tt.wantErr)
				}

				return
			}

			assert.NilError(t, err)
			var res activities.RemovePathsResult
			future.Get(&res)
			assert.DeepEqual(t, res, tt.want)

			for _, path := range paths {
				_, err = os.Stat(path)
				assert.ErrorContains(t, err, "no such file or directory")
			}
		})
	}
}
