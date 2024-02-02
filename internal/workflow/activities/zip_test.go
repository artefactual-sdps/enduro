package activities_test

import (
	"archive/zip"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

func TestZipActivity(t *testing.T) {
	transferName := "my_transfer"
	contents := tfs.WithDir(transferName,
		tfs.WithDir("subdir",
			tfs.WithFile("abc.txt", "Testing A-B-C"),
		),
		tfs.WithFile("123.txt", "Testing 1-2-3!"),
	)
	td := tfs.NewDir(t, "enduro-zip-test", contents)
	restrictedDir := tfs.NewDir(t, "enduro-zip-restricted", tfs.WithMode(0o555))

	type test struct {
		name    string
		params  activities.ZipActivityParams
		want    map[string]int64
		wantErr string
	}
	for _, tc := range []test{
		{
			name:   "Zips a directory",
			params: activities.ZipActivityParams{SourceDir: td.Join(transferName)},
			want: map[string]int64{
				"my_transfer/123.txt":        14,
				"my_transfer/subdir/abc.txt": 13,
			},
		},
		{
			name:    "Errors when SourceDir is missing",
			wantErr: "ZipActivity: missing source dir",
		},
		{
			name: "Errors when dest is not writable",
			params: activities.ZipActivityParams{
				SourceDir: td.Join(transferName),
				DestPath:  restrictedDir.Join(transferName + ".zip"),
			},
			wantErr: fmt.Sprintf("ZipActivity: create: open %s: permission denied", restrictedDir.Join(transferName+".zip")),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				activities.NewZipActivity(logr.Discard()).Execute,
				temporalsdk_activity.RegisterOptions{
					Name: activities.ZipActivityName,
				},
			)

			fut, err := env.ExecuteActivity(activities.ZipActivityName, tc.params)
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			assert.NilError(t, err)

			var res activities.ZipActivityResult
			_ = fut.Get(&res)
			assert.DeepEqual(t, res, activities.ZipActivityResult{Path: td.Join(transferName + ".zip")})

			// Confirm the zip has the expected contents.
			rc, err := zip.OpenReader(td.Join(transferName + ".zip"))
			assert.NilError(t, err)
			t.Cleanup(func() { rc.Close() })

			files := make(map[string]int64, len(rc.File))
			for _, f := range rc.File {
				files[f.Name] = f.FileInfo().Size()
			}
			assert.DeepEqual(t, files, tc.want)
		})
	}
}
