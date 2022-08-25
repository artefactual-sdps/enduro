package activities

import (
	"context"

	"github.com/artefactual-sdps/enduro/internal/watcher"
)

type DeleteOriginalActivity struct {
	wsvc watcher.Service
}

func NewDeleteOriginalActivity(wsvc watcher.Service) *DeleteOriginalActivity {
	return &DeleteOriginalActivity{wsvc: wsvc}
}

func (a *DeleteOriginalActivity) Execute(ctx context.Context, watcherName, key string) error {
	return a.wsvc.Delete(ctx, watcherName, key)
}
