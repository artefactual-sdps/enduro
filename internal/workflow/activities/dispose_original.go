package activities

import (
	"context"

	"github.com/artefactual-sdps/enduro/internal/watcher"
)

type DisposeOriginalActivity struct {
	wsvc watcher.Service
}

func NewDisposeOriginalActivity(wsvc watcher.Service) *DisposeOriginalActivity {
	return &DisposeOriginalActivity{wsvc: wsvc}
}

func (a *DisposeOriginalActivity) Execute(ctx context.Context, watcherName, completedDir, key string) error {
	return a.wsvc.Dispose(ctx, watcherName, key)
}
