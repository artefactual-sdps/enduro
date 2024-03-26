package activities

import (
	"context"

	"github.com/artefactual-sdps/enduro/internal/watcher"
)

type DisposeOriginalActivity struct {
	wsvc watcher.Service
}

type DisposeOriginalActivityResult struct{}

func NewDisposeOriginalActivity(wsvc watcher.Service) *DisposeOriginalActivity {
	return &DisposeOriginalActivity{wsvc: wsvc}
}

func (a *DisposeOriginalActivity) Execute(
	ctx context.Context,
	watcherName, completedDir, key string,
) (*DisposeOriginalActivityResult, error) {
	return &DisposeOriginalActivityResult{}, a.wsvc.Dispose(ctx, watcherName, key)
}
