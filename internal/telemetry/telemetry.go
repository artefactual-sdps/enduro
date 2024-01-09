package telemetry

import "context"

type ShutdownProvider func(ctx context.Context) error
