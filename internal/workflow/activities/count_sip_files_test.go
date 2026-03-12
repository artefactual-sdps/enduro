package activities_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

func TestCountSIPFilesActivity(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		params  *activities.CountSIPFilesActivityParams
		want    *activities.CountSIPFilesActivityResult
		wantErr string
	}
	for _, tc := range []test{
		{
			name: "count files in BagIt Bag",
			params: &activities.CountSIPFilesActivityParams{
				Path:    "../../testdata/bag/small_bag",
				SIPType: enums.SIPTypeBagIt,
			},
			want: &activities.CountSIPFilesActivityResult{
				Count: 1,
			},
		},
		{
			name: "count files in a standard transfer",
			params: &activities.CountSIPFilesActivityParams{
				Path:    "../../testdata/standard_transfer/small",
				SIPType: enums.SIPTypeUnknown,
			},
			want: &activities.CountSIPFilesActivityResult{
				Count: 1,
			},
		},
		{
			name: "missing data directory",
			params: &activities.CountSIPFilesActivityParams{
				Path:    "../../testdata/missing",
				SIPType: enums.SIPTypeBagIt,
			},
			wantErr: "count SIP files: directory not found: ../../testdata/missing/data",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			activity := activities.NewCountSIPFilesActivity()
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
