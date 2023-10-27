package activities

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-logr/logr"
)

// CleanUpActivity removes the contents that we've created in the TS location.
type CleanUpActivity struct {
	logger logr.Logger
}

func NewCleanUpActivity(logger logr.Logger) *CleanUpActivity {
	return &CleanUpActivity{logger: logger}
}

type CleanUpActivityParams struct {
	Paths []string
}

func (a *CleanUpActivity) Execute(ctx context.Context, params *CleanUpActivityParams) error {
	if params == nil || params.Paths == nil {
		a.logger.V(2).Info("CleanUpActivity: no paths to clean up.")
		return nil
	}

	a.logger.V(2).Info("Executing CleanUpActivity", "Paths", strings.Join(params.Paths, ","))

	for _, path := range params.Paths {
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("clean up: %v", err)
		}
	}

	return nil
}
