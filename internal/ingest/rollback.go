package ingest

import (
	"context"
	"time"
)

const rollbackCleanupTimeout = time.Second

// withRollbackCleanupContext runs cleanup with request-scoped values preserved
// but request cancellation ignored, bounded by the rollback cleanup timeout.
func withRollbackCleanupContext(ctx context.Context, cleanup func(context.Context) error) error {
	// Preserve request-scoped values for telemetry/logging while giving
	// rollback cleanup a fresh, bounded deadline after ingest startup fails.
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), rollbackCleanupTimeout)
	defer cancel()

	return cleanup(cleanupCtx)
}
