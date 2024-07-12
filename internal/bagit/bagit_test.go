package bagit_test

import (
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/bagit"
)

func TestBagitIs(t *testing.T) {
	t.Parallel()

	t.Run("Not a bag", func(t *testing.T) {
		t.Parallel()
		td := fs.NewDir(t, "enduro-test")
		assert.Equal(t, bagit.Is(td.Path()), false)
	})

	t.Run("Is a bag", func(t *testing.T) {
		t.Parallel()
		td := fs.NewDir(t, "enduro-test", fs.WithFile("bagit.txt", ""))
		assert.Equal(t, bagit.Is(td.Path()), true)
	})
}

func TestBagitComplete(t *testing.T) {
	assert.NilError(t, bagit.Complete("./tests/test-bagged-transfer"))
	assert.ErrorContains(
		t,
		bagit.Complete("./tests/test-bagged-transfer-with-invalid-oxum"),
		"Payload-Oxum validation failed. Expected 1 files and 7 bytes but found 2 files and 7 bytes",
	)
	assert.ErrorContains(
		t,
		bagit.Complete("./tests/test-bagged-transfer-with-missing-manifest"),
		"Bag validation failed: tests/test-bagged-transfer-with-missing-manifest/manifest-sha256.txt does not exist",
	)
	assert.ErrorContains(
		t,
		bagit.Complete("./tests/test-bagged-transfer-with-unexpected-files"),
		"Bag validation failed: data/dos.txt exists on filesystem but is not in the manifest",
	)
	assert.ErrorContains(
		t,
		bagit.Complete("./tests/nobag"),
		"- ERROR - input ./tests/nobag directory does not exist",
	)
}
