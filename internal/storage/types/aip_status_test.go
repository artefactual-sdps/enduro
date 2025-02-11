package types_test

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func TestAIPStatus(t *testing.T) {
	t.Parallel()

	type test struct {
		code   string
		status types.AIPStatus
	}
	for _, tt := range []test{
		{
			code:   "unspecified",
			status: types.AIPStatusUnspecified,
		},
		{
			code:   "in_review",
			status: types.AIPStatusInReview,
		},
		{
			code:   "rejected",
			status: types.AIPStatusRejected,
		},
		{
			code:   "stored",
			status: types.AIPStatusStored,
		},
		{
			code:   "moving",
			status: types.AIPStatusMoving,
		},
	} {
		t.Run(tt.code, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, types.NewAIPStatus(tt.code), tt.status)
			assert.Equal(t, tt.status.String(), tt.code)

			blob, err := json.Marshal(tt.status)
			assert.NilError(t, err)
			assert.DeepEqual(t, `"`+tt.code+`"`, string(blob))

			var st types.AIPStatus
			err = json.Unmarshal([]byte(`"`+tt.code+`"`), &st)
			assert.NilError(t, err)
			assert.Equal(t, st, tt.status)

			var ss types.AIPStatus
			err = ss.Scan(tt.code)
			assert.NilError(t, err)
			assert.Equal(t, ss, tt.status)

			assert.DeepEqual(t, ss.Values(), []string{
				types.AIPStatusUnspecified.String(),
				types.AIPStatusInReview.String(),
				types.AIPStatusRejected.String(),
				types.AIPStatusStored.String(),
				types.AIPStatusMoving.String(),
			})

			v, err := ss.Value()
			assert.NilError(t, err)
			assert.Equal(t, v, tt.code)
		})
	}
}
