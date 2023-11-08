package activities_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mholt/archiver/v3"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/sfa/activities"
)

func TestExtractPackage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		processingDir *fs.Dir
		zipFolders    []string
		wantErr       string
	}{
		{
			name:          "Succeeds",
			processingDir: fs.NewDir(t, "", fs.WithDir("folder")),
			zipFolders:    []string{"folder"},
		},
		{
			name:          "Fails without entry",
			processingDir: fs.NewDir(t, ""),
			wantErr:       "no entry found in extracted package directory",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			zip := archiver.NewZip()
			zipPath := filepath.Join(tt.processingDir.Path(), "package.zip")
			var zipFolders []string
			for _, f := range tt.zipFolders {
				zipFolders = append(zipFolders, filepath.Join(tt.processingDir.Path(), f))
			}
			err := zip.Archive(zipFolders, zipPath)
			assert.NilError(t, err)

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				activities.NewExtractPackage().Execute,
				temporalsdk_activity.RegisterOptions{Name: activities.ExtractPackageName},
			)

			future, err := env.ExecuteActivity(
				activities.ExtractPackageName,
				&activities.ExtractPackageParams{
					Path: zipPath,
					Key:  filepath.Base(zipPath),
				},
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
			var res activities.ExtractPackageResult
			future.Get(&res)
			assert.Assert(t, strings.HasPrefix(res.Path, tt.processingDir.Path()+"/package-"))
			assert.Assert(t, strings.HasSuffix(res.Path, "/folder"))

			_, err = os.Stat(res.Path)
			assert.NilError(t, err)
		})
	}
}
