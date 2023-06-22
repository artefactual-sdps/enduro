package activities

import (
	"context"
	"fmt"
	"os"

	"github.com/artefactual-sdps/enduro/internal/temporal"
	"github.com/artefactual-sdps/enduro/internal/watcher"
)

// DownloadActivity downloads the blob into the processing directory.
type DownloadActivity struct {
	wsvc watcher.Service
}

func NewDownloadActivity(wsvc watcher.Service) *DownloadActivity {
	return &DownloadActivity{
		wsvc: wsvc,
	}
}

func tempFile(pattern string) (*os.File, error) {
	if pattern == "" {
		pattern = "blob-*"
	}
	return os.CreateTemp("", pattern)
}

func (a *DownloadActivity) Execute(ctx context.Context, watcherName, key string) (string, error) {
	file, err := tempFile("blob-*")
	if err != nil {
		return "", temporal.NonRetryableError(fmt.Errorf(
			"error creating temporary file in processing directory: %v", err))
	}
	defer file.Close() //#nosec G307 -- Errors returned by Close() here do not require specific handling.

	if err := a.wsvc.Download(ctx, file, watcherName, key); err != nil {
		return "", temporal.NonRetryableError(fmt.Errorf("error downloading blob: %v", err))
	}

	return file.Name(), nil
}
