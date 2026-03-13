package activities_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

func TestCountBagFilesActivity(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		params  *activities.CountBagFilesActivityParams
		want    *activities.CountBagFilesActivityResult
		wantErr string
	}
	for _, tc := range []test{
		{
			name: "count files in bag",
			params: &activities.CountBagFilesActivityParams{
				Path: "../../testdata/bag/small_bag",
			},
			want: &activities.CountBagFilesActivityResult{
				Count: 1,
			},
		},
		{
			name: "missing data directory",
			params: &activities.CountBagFilesActivityParams{
				Path: "../../testdata/missing",
			},
			wantErr: "count bag files: missing data dir: ../../testdata/missing/data",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			activity := activities.NewCountBagFilesActivity()
			got, err := activity.Execute(context.Background(), tc.params)
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, tc.want.Count, got.Count)
		})
	}
}
