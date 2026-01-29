package childwf

import (
	"errors"
	"fmt"
)

type Config struct {
	// Enabled toggles whether to run the child workflow.
	Enabled bool

	// Namespace is the Temporal namespace of the child workflow.
	Namespace string

	// TaskQueue is the Temporal task queue to use for child workflow tasks.
	TaskQueue string

	// WorkflowName is the Temporal workflow name for the child workflow.
	WorkflowName string
}

type Configs []Config

func (c Configs) GetByName(name string) *Config {
	for _, cw := range c {
		if cw.WorkflowName == name {
			return &cw
		}
	}
	return nil
}

func (c Configs) IsEnabled(name string) bool {
	cw := c.GetByName(name)
	if cw == nil {
		return false
	}
	return cw.Enabled
}

func (c Configs) Validate() error {
	var errs error
	for i, cfg := range c {
		if !cfg.Enabled {
			continue
		}
		if cfg.Namespace == "" {
			errs = errors.Join(errs, fmt.Errorf("child workflows[%d]: namespace is required", i))
		}
		if cfg.TaskQueue == "" {
			errs = errors.Join(errs, fmt.Errorf("child workflows[%d]: task queue is required", i))
		}
		if cfg.WorkflowName == "" {
			errs = errors.Join(errs, fmt.Errorf("child workflows[%d]: workflow name is required", i))
		}
	}

	return errs
}
