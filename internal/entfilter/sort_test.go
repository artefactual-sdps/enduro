package entfilter_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/entfilter"
)

func TestSort(t *testing.T) {
	got := entfilter.NewSort().AddCol("id", false).AddCol("date", true)

	assert.DeepEqual(t, got, entfilter.Sort{
		{Name: "id", Desc: false},
		{Name: "date", Desc: true},
	})
}
