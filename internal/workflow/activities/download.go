package activities

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	temporal_tools "go.artefactual.dev/tools/temporal"
	"go.opentelemetry.io/otel/trace"

	"github.com/artefactual-sdps/enduro/internal/telemetry"
	"github.com/artefactual-sdps/enduro/internal/watcher"
)

// DownloadActivity downloads the blob into the processing directory.
type DownloadActivity struct {
	tracer trace.Tracer
	wsvc   watcher.Service
}

type DownloadActivityParams struct {
	Key             string
	WatcherName     string
	DestinationPath string
}

type DownloadActivityResult struct {
	Path string
}

func NewDownloadActivity(tracer trace.Tracer, wsvc watcher.Service) *DownloadActivity {
	return &DownloadActivity{
		tracer: tracer,
		wsvc:   wsvc,
	}
}

func (a *DownloadActivity) Execute(
	ctx context.Context,
	params *DownloadActivityParams,
) (*DownloadActivityResult, error) {
	logger := temporal_tools.GetLogger(ctx)
	logger.V(1).Info("Executing DownloadActivity",
		"Key", params.Key,
		"WatcherName", params.WatcherName,
	)

	destDir, err := os.MkdirTemp(params.DestinationPath, "enduro")
	if err != nil {
		return &DownloadActivityResult{}, temporal_tools.NewNonRetryableError(fmt.Errorf("make temp dir: %v", err))
	}

	dest := filepath.Clean(filepath.Join(destDir, params.Key))

	ctx, span := a.tracer.Start(ctx, "download")
	if err := a.wsvc.Download(ctx, dest, params.WatcherName, params.Key); err != nil {
		telemetry.RecordError(span, err)
		return &DownloadActivityResult{}, temporal_tools.NewNonRetryableError(fmt.Errorf("download: %v", err))
	}
	span.End()

	return &DownloadActivityResult{Path: dest}, nil
}
