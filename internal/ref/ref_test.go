package ref_test

import (
	"testing"

	"go.artefactual.dev/tools/ref"
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

func TestDerefZero(t *testing.T) {
	t.Parallel()

	t.Run("Returns the underlying value of the pointer", func(t *testing.T) {
		s := "string"
		assert.Equal(t, ref.DerefZero(&s), "string")
	})

	t.Run("Returns the default value if the pointer is nil", func(t *testing.T) {
		assert.Equal(t, ref.DerefZero[string](nil), "")
	})
}
