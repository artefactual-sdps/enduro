package activities

import (
	"context"
	"errors"
	"fmt"
	"os"

	"go.artefactual.dev/tools/temporal"
)

const RemovePathsName = "remove-paths"

type RemovePathsParams struct {
	Paths []string
}
type RemovePathsResult struct{}

type RemovePaths struct{}

func NewRemovePaths() *RemovePaths {
	return &RemovePaths{}
}

func (a *RemovePaths) Execute(ctx context.Context, params *RemovePathsParams) (*RemovePathsResult, error) {
	var e error

	for _, path := range params.Paths {
		if err := os.RemoveAll(path); err != nil {
			e = errors.Join(e, fmt.Errorf("error removing path: %v", err))
		}
	}

	if e != nil {
		return nil, temporal.NewNonRetryableError(e)
	}

	return &RemovePathsResult{}, nil
}
