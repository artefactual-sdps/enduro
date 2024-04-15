package activities_test

import (
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
	"go.artefactual.dev/tools/temporal"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

func TestBundleActivity(t *testing.T) {
	t.Parallel()

	sourceDir := fs.NewDir(t, "enduro-bundle-test",
		fs.FromDir("../../testdata"),
	)
	destDir := fs.NewDir(t, "enduro-bundle-test")

	type test struct {
		name    string
		params  *activities.BundleActivityParams
		wantFs  fs.Manifest
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Bundles a single file",
			params: &activities.BundleActivityParams{
				SourcePath:  sourceDir.Join("single_file_transfer", "small.txt"),
				TransferDir: destDir.Path(),
			},
			wantFs: fs.Expected(t, fs.WithMode(activities.ModeDir),
				fs.WithDir("objects", fs.WithMode(activities.ModeDir),
					fs.WithFile(
						"small.txt", "I am a small file.\n", fs.WithMode(activities.ModeFile),
					),
				),
				fs.WithDir("metadata", fs.WithMode(activities.ModeDir)),
			),
		},
		{
			name: "Bundles a local standard transfer directory",
			params: &activities.BundleActivityParams{
				SourcePath:  sourceDir.Join("standard_transfer", "small"),
				TransferDir: destDir.Path(),
				IsDir:       true,
			},
			wantFs: fs.Expected(t, fs.WithMode(activities.ModeDir),
				fs.WithFile("small.txt", "I am a small file.\n", fs.WithMode(activities.ModeFile)),
			),
		},
		{
			name: "Bundles a BagIt transfer",
			params: &activities.BundleActivityParams{
				SourcePath:  sourceDir.Join("bag", "small_bag"),
				TransferDir: destDir.Path(),
				IsDir:       true,
			},
			wantFs: fs.Expected(t, fs.WithMode(activities.ModeDir),
				fs.WithFile(
					"small.txt", "I am a small file.\n",
					fs.WithMode(activities.ModeFile),
				),
				fs.WithDir(
					"metadata",
					fs.WithMode(activities.ModeDir),
					fs.WithFile(
						"checksum.sha256",
						"4450c8a88130a3b397bfc659245c4f0f87a8c79d017a60bdb1bd32f4b51c8133  ../objects/small.txt\n",
						fs.WithMode(activities.ModeFile),
					),
					fs.WithDir(
						"submissionDocumentation",
						fs.WithMode(activities.ModeDir),
						fs.WithFile(
							"bag-info.txt",
							`Bag-Software-Agent: bagit.py v1.8.1 <https://github.com/LibraryOfCongress/bagit-python>
Bagging-Date: 2023-12-12
Payload-Oxum: 19.1
`,
							fs.WithMode(activities.ModeFile),
						),
						fs.WithFile(
							"bagit.txt",
							`BagIt-Version: 0.97
Tag-File-Character-Encoding: UTF-8
`,
							fs.WithMode(activities.ModeFile),
						),
						fs.WithFile(
							"manifest-sha256.txt",
							`4450c8a88130a3b397bfc659245c4f0f87a8c79d017a60bdb1bd32f4b51c8133  data/small.txt
`,
							fs.WithMode(activities.ModeFile),
						),
						fs.WithFile(
							"tagmanifest-sha256.txt",
							`ac3f0fa6e7763ba403c1bca2b6e785a51bfcd5102fe7cbc1cfcf05be77ffdf24 manifest-sha256.txt
fd696a4957ed3f8329860c7191e518162b99c942b26b42291386da69bb3c2bc8 bag-info.txt
e91f941be5973ff71f1dccbdd1a32d598881893a7f21be516aca743da38b1689 bagit.txt
`,
							fs.WithMode(activities.ModeFile),
						),
					),
				),
			),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				activities.NewBundleActivity(logr.Discard()).Execute,
				temporalsdk_activity.RegisterOptions{
					Name: activities.BundleActivityName,
				},
			)

			enc, err := env.ExecuteActivity(activities.BundleActivityName, tt.params)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				assert.Assert(t, temporal.NonRetryableError(err))
				return
			}
			assert.NilError(t, err)

			var res activities.BundleActivityResult
			_ = enc.Get(&res)

			assert.Assert(t, res.FullPath != "")
			if tt.wantFs != (fs.Manifest{}) {
				assert.Assert(t, fs.Equal(res.FullPath, tt.wantFs))
			}
			assert.NilError(t, err)

			rp, err := filepath.Rel(tt.params.TransferDir, res.FullPath)
			assert.NilError(t, err)
			assert.Assert(t, len(rp) > 0)
		})
	}
}
