package activities_test

import (
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
	"go.artefactual.dev/tools/temporal"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	watcherfake "github.com/artefactual-sdps/enduro/internal/watcher/fake"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

func TestBundleActivity(t *testing.T) {
	watcherName := "watcher"
	sourceDir := fs.NewDir(t, "enduro-bundle-test",
		fs.FromDir("../../testdata"),
	)
	destDir := fs.NewDir(t, "enduro-bundle-test")

	type test struct {
		name     string
		params   *activities.BundleActivityParams
		msvc     func(*gomock.Controller) *watcherfake.MockService
		watchRec func(*watcherfake.MockWatcherMockRecorder)
		wantFs   fs.Manifest
		wantErr  string
	}
	for _, tt := range []test{
		{
			name: "Bundles a single file",
			params: &activities.BundleActivityParams{
				WatcherName: watcherName,
				TransferDir: destDir.Path(),
				Key:         "small.txt",
				TempFile:    sourceDir.Join("single_file_transfer", "small.txt"),
			},
			wantFs: fs.Expected(t, fs.WithMode(0o755),
				fs.WithDir("objects", fs.WithMode(0o755),
					fs.WithFile(
						"small.txt", "I am a small file.\n", fs.WithMode(0o755),
					),
				),
				fs.WithDir("metadata", fs.WithMode(0o755)),
			),
		},
		{
			name: "Bundles a local standard transfer directory",
			params: &activities.BundleActivityParams{
				WatcherName: watcherName,
				TransferDir: destDir.Path(),
				Key:         "small",
				TempFile:    sourceDir.Join("standard_transfer", "small"),
				IsDir:       true,
			},
			msvc: func(ctrl *gomock.Controller) *watcherfake.MockService {
				svc := watcherfake.NewMockService(ctrl)
				watcher := watcherfake.NewMockWatcher(ctrl)

				svc.EXPECT().
					ByName(watcherName).
					Return(
						watcher,
						nil,
					)

				watcher.EXPECT().
					Path().
					Return(sourceDir.Join("standard_transfer"))

				return svc
			},
			watchRec: func(watcher *watcherfake.MockWatcherMockRecorder) {
				watcher.Path().Return(sourceDir.Join("standard_transfer"))
			},
			wantFs: fs.Expected(t, fs.WithMode(0o755),
				fs.WithFile("small.txt", "I am a small file.\n", fs.WithMode(0o644)),
			),
		},
		{
			name: "Bundles a zipped standard transfer",
			params: &activities.BundleActivityParams{
				WatcherName:      watcherName,
				TransferDir:      destDir.Path(),
				Key:              "small.zip",
				TempFile:         sourceDir.Join("zipped_transfer", "small.zip"),
				StripTopLevelDir: true,
			},
			wantFs: fs.Expected(t, fs.WithMode(0o775),
				fs.WithFile("small.txt", "I am a small file.\n", fs.WithMode(0o664)),
			),
		},
		{
			name: "Bundles a tarred and gzipped bag transfer",
			params: &activities.BundleActivityParams{
				WatcherName:      watcherName,
				TransferDir:      destDir.Path(),
				Key:              "small_bag.tgz",
				TempFile:         sourceDir.Join("gzipped_bag", "small_bag.tgz"),
				StripTopLevelDir: true,
			},
			wantFs: fs.Expected(t, fs.WithMode(0o775),
				fs.WithFile(
					"small.txt", "I am a small file.\n",
					fs.WithMode(0o664),
				),
				fs.WithDir(
					"metadata",
					fs.WithMode(0o775),
					fs.WithFile(
						"checksum.sha256",
						"4450c8a88130a3b397bfc659245c4f0f87a8c79d017a60bdb1bd32f4b51c8133  ../objects/small.txt\n",
						fs.WithMode(0o664),
					),
					fs.WithDir(
						"submissionDocumentation",
						fs.WithMode(0o775),
						fs.WithFile(
							"bag-info.txt",
							`Bag-Software-Agent: bagit.py v1.8.1 <https://github.com/LibraryOfCongress/bagit-python>
Bagging-Date: 2023-12-12
Payload-Oxum: 19.1
`,
							fs.WithMode(0o664),
						),
						fs.WithFile(
							"bagit.txt",
							`BagIt-Version: 0.97
Tag-File-Character-Encoding: UTF-8
`,
							fs.WithMode(0o664),
						),
						fs.WithFile(
							"manifest-sha256.txt",
							`4450c8a88130a3b397bfc659245c4f0f87a8c79d017a60bdb1bd32f4b51c8133  data/small.txt
`,
							fs.WithMode(0o664),
						),
						fs.WithFile(
							"tagmanifest-sha256.txt",
							`ac3f0fa6e7763ba403c1bca2b6e785a51bfcd5102fe7cbc1cfcf05be77ffdf24 manifest-sha256.txt
fd696a4957ed3f8329860c7191e518162b99c942b26b42291386da69bb3c2bc8 bag-info.txt
e91f941be5973ff71f1dccbdd1a32d598881893a7f21be516aca743da38b1689 bagit.txt
`,
							fs.WithMode(0o664),
						),
					),
				),
			),
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var wsvc *watcherfake.MockService

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			ctrl := gomock.NewController(t)
			if tt.msvc != nil {
				wsvc = tt.msvc(ctrl)
			}

			env.RegisterActivityWithOptions(
				activities.NewBundleActivity(logr.Discard(), wsvc).Execute,
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

			if tt.params.StripTopLevelDir {
				assert.Equal(t, res.FullPathBeforeStrip, filepath.Dir(res.FullPath))
			} else {
				assert.Equal(t, res.FullPathBeforeStrip, res.FullPath)
			}

			p, err := filepath.Rel(destDir.Path(), res.FullPath)
			assert.NilError(t, err)
			assert.Equal(t, res.RelPath, p)
		})
	}
}
