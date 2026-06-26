package ingest

import (
	"context"
	"time"
)

const failedIngestCleanupTimeout = 10 * time.Second

// withFailedIngestCleanupContext runs compensating cleanup after ingest state
// has been created but the follow-up operation failed. Cleanup is attempted for
// any failure, not only context cancellation, but it uses a fresh bounded
// context so a canceled request cannot prevent cleanup from running.
func withFailedIngestCleanupContext(ctx context.Context, cleanup func(context.Context) error) error {
	// Preserve request-scoped values for telemetry/logging while giving cleanup
	// a fresh, bounded deadline after ingest startup fails.
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), failedIngestCleanupTimeout)
	defer cancel()

	return cleanup(cleanupCtx)
}
