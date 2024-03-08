package activities

import (
	"context"
	"fmt"
	"os"
)

// CleanUpActivity removes the contents that we've created in the TS location.
type CleanUpActivity struct{}

func NewCleanUpActivity() *CleanUpActivity {
	return &CleanUpActivity{}
}

type CleanUpActivityParams struct {
	FullPath string
}

type CleanUpActivityResult struct{}

func (a *CleanUpActivity) Execute(ctx context.Context, params *CleanUpActivityParams) (*CleanUpActivityResult, error) {
	if params == nil || params.FullPath == "" {
		return &CleanUpActivityResult{}, fmt.Errorf("error processing parameters: missing or empty")
	}

	if err := os.RemoveAll(params.FullPath); err != nil {
		return &CleanUpActivityResult{}, fmt.Errorf("error removing transfer directory: %v", err)
	}

	return &CleanUpActivityResult{}, nil
}
