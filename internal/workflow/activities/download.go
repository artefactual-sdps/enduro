package activities

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	temporal_tools "go.artefactual.dev/tools/temporal"

	"github.com/artefactual-sdps/enduro/internal/watcher"
)

// DownloadActivity downloads the blob into the processing directory.
type DownloadActivity struct {
	logger logr.Logger
	wsvc   watcher.Service
}

type DownloadActivityParams struct {
	Key         string
	WatcherName string
}

type DownloadActivityResult struct {
	Path string
}

func NewDownloadActivity(logger logr.Logger, wsvc watcher.Service) *DownloadActivity {
	return &DownloadActivity{
		logger: logger,
		wsvc:   wsvc,
	}
}

func (a *DownloadActivity) Execute(ctx context.Context, params *DownloadActivityParams) (*DownloadActivityResult, error) {
	a.logger.V(1).Info("Executing DownloadActivity",
		"Key", params.Key,
		"WatcherName", params.WatcherName,
	)

	destDir, err := os.MkdirTemp("", "enduro")
	if err != nil {
		return nil, temporal_tools.NewNonRetryableError(fmt.Errorf("make temp dir: %v", err))
	}

	dest := filepath.Clean(filepath.Join(destDir, params.Key))
	if err := a.wsvc.Download(ctx, dest, params.WatcherName, params.Key); err != nil {
		return nil, temporal_tools.NewNonRetryableError(fmt.Errorf("download: %v", err))
	}

	return &DownloadActivityResult{Path: dest}, nil
}
