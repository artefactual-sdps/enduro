package types_test

import (
	"context"
	"io"
	"testing"

	"github.com/artefactual-sdps/enduro/internal/storage/types"
	"gotest.tools/v3/assert"
)

// fix tests to work for archivematica storage service...
func TestBucket(t *testing.T) {
	bucket, err := types.OpenBucket(nil)
	assert.NilError(t, err)
	defer bucket.Close()

	r, err := bucket.NewReader(context.Background(), "", nil)
	assert.NilError(t, err)

	n, err := io.ReadAll(r)
	assert.NilError(t, err)
	assert.Assert(t, len(n) > 0)
	// change content type
	assert.Equal(t, r.ContentType(), "image/jpeg")
}
