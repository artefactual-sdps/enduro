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

var bagPathOps = []fs.PathOp{
	fs.WithDir("data",
		fs.WithMode(activities.ModeDir),
		fs.WithFile(
			"small.txt", "I am a small file.\n",
			fs.WithMode(activities.ModeFile),
		),
	),
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
}

func TestUnarchiveActivity(t *testing.T) {
	type test struct {
		name    string
		params  *activities.UnarchiveParams
		want    *activities.UnarchiveResult
		wantFs  fs.Manifest
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Unarchives a zipped standard transfer",
			params: &activities.UnarchiveParams{
				SourcePath: filepath.Join("zipped_transfer", "small.zip"),
			},
			want: &activities.UnarchiveResult{
				DestPath: filepath.Join("zipped_transfer", "extract"),
				IsDir:    true,
			},
			wantFs: fs.Expected(t, fs.WithMode(activities.ModeDir),
				fs.WithDir("small", fs.WithMode(activities.ModeDir),
					fs.WithFile("small.txt",
						"I am a small file.\n",
						fs.WithMode(activities.ModeFile),
					),
				),
			),
		},
		{
			name: "Unarchives a zipped transfer and strips the top-level dir",
			params: &activities.UnarchiveParams{
				SourcePath:       filepath.Join("zipped_transfer", "small.zip"),
				StripTopLevelDir: true,
			},
			want: &activities.UnarchiveResult{
				DestPath: filepath.Join("zipped_transfer", "extract"),
				IsDir:    true,
			},
			wantFs: fs.Expected(t, fs.WithMode(activities.ModeDir),
				fs.WithFile("small.txt",
					"I am a small file.\n",
					fs.WithMode(activities.ModeFile),
				),
			),
		},
		{
			name: "Unarchives a tarred and gzipped bag transfer",
			params: &activities.UnarchiveParams{
				SourcePath: filepath.Join("gzipped_bag", "small_bag.tgz"),
			},
			want: &activities.UnarchiveResult{
				DestPath: filepath.Join("gzipped_bag", "extract"),
				IsDir:    true,
			},
			wantFs: fs.Expected(t, fs.WithMode(activities.ModeDir),
				fs.WithDir("small_bag", append(
					[]fs.PathOp{fs.WithMode(activities.ModeDir)},
					bagPathOps...,
				)...),
			),
		},
		{
			name: "Unarchives a tgz bag transfer and strips top-level dir",
			params: &activities.UnarchiveParams{
				SourcePath:       filepath.Join("gzipped_bag", "small_bag.tgz"),
				StripTopLevelDir: true,
			},
			want: &activities.UnarchiveResult{
				DestPath: filepath.Join("gzipped_bag", "extract"),
				IsDir:    true,
			},
			wantFs: fs.Expected(t, append(
				[]fs.PathOp{fs.WithMode(activities.ModeDir)},
				bagPathOps...,
			)...),
		},
		{
			name: "Returns a directory path unaltered",
			params: &activities.UnarchiveParams{
				SourcePath: filepath.Join("bag", "small_bag"),
			},
			want: &activities.UnarchiveResult{
				DestPath: filepath.Join("bag", "small_bag"),
				IsDir:    true,
			},
		},
		{
			name: "Returns a non-archive file path unaltered",
			params: &activities.UnarchiveParams{
				SourcePath: filepath.Join("single_file_transfer", "small.txt"),
			},
			want: &activities.UnarchiveResult{
				DestPath: filepath.Join("single_file_transfer", "small.txt"),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				activities.NewUnarchiveActivity(logr.Discard()).Execute,
				temporalsdk_activity.RegisterOptions{
					Name: activities.UnarchiveActivityName,
				},
			)

			// New source dir for each test run.
			sourceDir := fs.NewDir(t, "enduro-test-unarchive",
				fs.FromDir("../../testdata"),
			)
			tt.params.SourcePath = sourceDir.Join(tt.params.SourcePath)

			enc, err := env.ExecuteActivity(activities.UnarchiveActivityName, tt.params)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				assert.Assert(t, temporal.NonRetryableError(err))
				return
			}
			assert.NilError(t, err)

			var res activities.UnarchiveResult
			_ = enc.Get(&res)

			if tt.want != nil {
				tt.want.DestPath = sourceDir.Join(tt.want.DestPath)
				assert.DeepEqual(t, &res, tt.want)
			}
			if tt.wantFs != (fs.Manifest{}) {
				assert.Assert(t, fs.Equal(res.DestPath, tt.wantFs))
			}
		})
	}
}
