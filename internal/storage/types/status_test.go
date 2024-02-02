package types_test

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func TestPackageStatus(t *testing.T) {
	t.Parallel()

	type test struct {
		code   string
		status types.PackageStatus
	}
	for _, tt := range []test{
		{
			code:   "unspecified",
			status: types.StatusUnspecified,
		},
		{
			code:   "in_review",
			status: types.StatusInReview,
		},
		{
			code:   "rejected",
			status: types.StatusRejected,
		},
		{
			code:   "stored",
			status: types.StatusStored,
		},
		{
			code:   "moving",
			status: types.StatusMoving,
		},
	} {
		t.Run(tt.code, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, types.NewPackageStatus(tt.code), tt.status)
			assert.Equal(t, tt.status.String(), tt.code)

			blob, err := json.Marshal(tt.status)
			assert.NilError(t, err)
			assert.DeepEqual(t, `"`+tt.code+`"`, string(blob))

			var st types.PackageStatus
			err = json.Unmarshal([]byte(`"`+tt.code+`"`), &st)
			assert.NilError(t, err)
			assert.Equal(t, st, tt.status)

			var ss types.PackageStatus
			err = ss.Scan(tt.code)
			assert.NilError(t, err)
			assert.Equal(t, ss, tt.status)

			assert.DeepEqual(t, ss.Values(), []string{
				types.StatusUnspecified.String(),
				types.StatusInReview.String(),
				types.StatusRejected.String(),
				types.StatusStored.String(),
				types.StatusMoving.String(),
			})

			v, err := ss.Value()
			assert.NilError(t, err)
			assert.Equal(t, v, tt.code)
		})
	}
}
