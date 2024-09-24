package persistence_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/persistence"
)

func TestOrder(t *testing.T) {
	got := persistence.NewSort().AddCol("id", false).AddCol("date", true)

	assert.DeepEqual(t, got, persistence.Sort{
		{Name: "id", Desc: false},
		{Name: "date", Desc: true},
	})
}
