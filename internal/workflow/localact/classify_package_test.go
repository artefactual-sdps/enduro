package localact_test

import (
	"testing"

	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/workflow/localact"
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
		params localact.ClassifyPackageActivityParams
		want   localact.ClassifyPackageActivityResult
	}
	for _, tt := range []test{
		{
			name: "Returns an unknown package type",
			params: localact.ClassifyPackageActivityParams{
				Path: fs.NewDir(t, "enduro-test").Path(),
			},
			want: localact.ClassifyPackageActivityResult{Type: enums.PackageTypeUnknown},
		},
		{
			name:   "Returns a bagit package type",
			params: localact.ClassifyPackageActivityParams{Path: testBag(t)},
			want:   localact.ClassifyPackageActivityResult{Type: enums.PackageTypeBagIt},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			enc, err := env.ExecuteLocalActivity(
				localact.ClassifyPackageActivity,
				tt.params,
			)
			assert.NilError(t, err)

			var res localact.ClassifyPackageActivityResult
			_ = enc.Get(&res)
			assert.DeepEqual(t, res, tt.want)
		})
	}
}
