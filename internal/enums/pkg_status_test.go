package enums_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/enums"
)

func TestStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		str string
		val enums.PackageStatus
	}{
		{
			str: "new",
			val: enums.PackageStatusNew,
		},
		{
			str: "in progress",
			val: enums.PackageStatusInProgress,
		},
		{
			str: "done",
			val: enums.PackageStatusDone,
		},
		{
			str: "error",
			val: enums.PackageStatusError,
		},
		{
			str: "queued",
			val: enums.PackageStatusQueued,
		},
		{
			str: "abandoned",
			val: enums.PackageStatusAbandoned,
		},
		{
			str: "pending",
			val: enums.PackageStatusPending,
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("Status_%s", tc.str), func(t *testing.T) {
			s := enums.NewPackageStatus(tc.str)
			assert.Assert(t, s != enums.PackageStatusUnknown)
			assert.Equal(t, s, tc.val)

			assert.Equal(t, s.String(), tc.str)

			b, err := json.Marshal(s)
			assert.NilError(t, err)
			assert.DeepEqual(t, b, []byte("\""+tc.str+"\""))

			json.Unmarshal([]byte("\""+tc.str+"\""), &s)
			assert.Assert(t, s != enums.PackageStatusUnknown)
			assert.Equal(t, s, tc.val)
		})
	}
}

func TestStatusUnknown(t *testing.T) {
	s := enums.NewPackageStatus("?")

	assert.Equal(t, s, enums.PackageStatusUnknown)
	assert.Equal(t, s.String(), enums.PackageStatusUnknown.String())
}
