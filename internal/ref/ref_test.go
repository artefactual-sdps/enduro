package ref_test

import (
	"testing"

	"github.com/artefactual-sdps/enduro/internal/ref"
	"gotest.tools/v3/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	s := "string"

	p1 := ref.New(s)
	assert.Equal(t, s, *p1)
	assert.Assert(t, &s != p1)

	p2 := ref.New(s)
	assert.Equal(t, s, *p2)
	assert.Assert(t, &s != p2)
}

func TestDeref(t *testing.T) {
	t.Parallel()

	s := "string"

	assert.Equal(t, s, ref.Deref(&s))
}
