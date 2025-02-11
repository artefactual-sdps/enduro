package activities_test

import (
	"testing"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

func testBag(t *testing.T) string {
	d := fs.NewDir(t, "enduro-test", fs.WithFile("bagit.txt", `
BagIt-Version: 0.97
Tag-File-Character-Encoding: UTF-8
`))

	return d.Path()
}

func TestClassifyPackageActivity(t *testing.T) {
	t.Parallel()

	type test struct {
		name   string
		params activities.ClassifyPackageActivityParams
		want   activities.ClassifyPackageActivityResult
	}
	for _, tt := range []test{
		{
			name: "Returns an unknown package type",
			params: activities.ClassifyPackageActivityParams{
				Path: fs.NewDir(t, "enduro-test").Path(),
			},
			want: activities.ClassifyPackageActivityResult{Type: enums.SIPTypeUnknown},
		},
		{
			name:   "Returns a bagit package type",
			params: activities.ClassifyPackageActivityParams{Path: testBag(t)},
			want:   activities.ClassifyPackageActivityResult{Type: enums.SIPTypeBagIt},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				activities.NewClassifyPackageActivity().Execute,
				temporalsdk_activity.RegisterOptions{
					Name: activities.ClassifyPackageActivityName,
				},
			)
			enc, err := env.ExecuteActivity(
				activities.ClassifyPackageActivityName,
				tt.params,
			)
			assert.NilError(t, err)

			var res activities.ClassifyPackageActivityResult
			_ = enc.Get(&res)
			assert.DeepEqual(t, res, tt.want)
		})
	}
}
