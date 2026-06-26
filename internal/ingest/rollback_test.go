package ingest_test

import (
	"context"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func assertRollbackCleanupContext(t *testing.T, ctx context.Context) {
	t.Helper()

	assert.NilError(t, ctx.Err())
	deadline, ok := ctx.Deadline()
	assert.Assert(t, ok)
	remaining := time.Until(deadline)
	assert.Assert(t, remaining > 0)
	assert.Assert(t, remaining <= 2*time.Second)
}
