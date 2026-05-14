package activities_test

import (
	"path/filepath"
	"testing"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

func TestCalcFileChecksumActivity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		params  activities.CalcFileChecksumActivityParams
		want    activities.CalcFileChecksumActivityResult
		wantErr string
	}{
		{
			name: "Calculates file checksum",
			params: activities.CalcFileChecksumActivityParams{
				Path: filepath.Join("..", "..", "testdata", "zipped_transfer", "small.zip"),
			},
			want: activities.CalcFileChecksumActivityResult{
				Algo: "SHA-256",
				Hash: "3ee958afa3bacb8698d0b7a549a6192d2bdf9feb53b409941a00a9d69da2ae6c",
			},
		},
		{
			name: "Fails when file is not found",
			params: activities.CalcFileChecksumActivityParams{
				Path: filepath.Join("..", "..", "testdata", "zipped_transfer", "missing.zip"),
			},
			wantErr: "calculate file checksum: file not found",
		},
		{
			name: "Fails when path is a directory",
			params: activities.CalcFileChecksumActivityParams{
				Path: filepath.Join("..", "..", "testdata", "zipped_transfer"),
			},
			wantErr: "calculate file checksum: not a file",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				activities.NewCalcFileChecksumActivity().Execute,
				temporalsdk_activity.RegisterOptions{Name: activities.CalcFileChecksumActivityName},
			)

			future, err := env.ExecuteActivity(
				activities.CalcFileChecksumActivityName,
				&tt.params,
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
			var res activities.CalcFileChecksumActivityResult
			future.Get(&res)
			assert.DeepEqual(t, res, tt.want)
		})
	}
}
