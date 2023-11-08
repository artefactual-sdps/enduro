package fformat_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/sfa/fformat"
)

func TestSiegfriedEmbed(t *testing.T) {
	t.Parallel()

	sf := fformat.NewSiegfriedEmbed()

	got, err := sf.Identify("fformat.go")
	assert.NilError(t, err)
	assert.DeepEqual(t, got, &fformat.FileFormat{
		Namespace:  "pronom",
		ID:         "x-fmt/111",
		CommonName: "Plain Text File",
		MIMEType:   "text/plain",
		Basis:      "text match ASCII",
		Warning:    "match on text only; extension mismatch",
	})

	_, err = sf.Identify("foobar.txt")
	assert.Error(t, err, "open foobar.txt: no such file or directory")
}

func BenchmarkSiegfried(b *testing.B) {
	b.Run("SiegfriedEmbed", func(b *testing.B) {
		sf := fformat.NewSiegfriedEmbed()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sf.Identify("fformat.go")
		}
	})
}
