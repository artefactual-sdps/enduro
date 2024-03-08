package activities

import (
	"context"

	"github.com/artefactual-sdps/enduro/internal/watcher"
)

type DeleteOriginalActivity struct {
	wsvc watcher.Service
}

type DeleteOriginalActivityResult struct{}

func NewDeleteOriginalActivity(wsvc watcher.Service) *DeleteOriginalActivity {
	return &DeleteOriginalActivity{wsvc: wsvc}
}

func (a *DeleteOriginalActivity) Execute(ctx context.Context, watcherName, key string) (*DeleteOriginalActivityResult, error) {
	return &DeleteOriginalActivityResult{}, a.wsvc.Delete(ctx, watcherName, key)
}
