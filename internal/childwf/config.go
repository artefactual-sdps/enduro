package childwf

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/artefactual-sdps/enduro/internal/enums"
)

type Config struct {
	// Type of the child workflow.
	Type enums.ChildWorkflowType

	// Namespace is the Temporal namespace of the child workflow.
	Namespace string

	// TaskQueue is the Temporal task queue to use for child workflow tasks.
	TaskQueue string

	// WorkflowName is the Temporal workflow name for the child workflow.
	WorkflowName string

	// SharedPath is a filesystem path shared between Enduro and the child
	// workflow (Preprocessing only).
	SharedPath string

	// Extract the SIP in the child workflow (Preprocessing only).
	Extract bool
}

func (c Config) Validate() error {
	errs := c.missingFields()

	if c.Type != "" && !c.Type.IsValid() {
		errs = errors.Join(errs, fmt.Errorf("invalid type: %s", c.Type))
	}

	return errs
}

func (c Config) missingFields() error {
	missing := make([]string, 0)

	if c.Type == "" {
		missing = append(missing, "type")
	}
	if c.Namespace == "" {
		missing = append(missing, "namespace")
	}
	if c.TaskQueue == "" {
		missing = append(missing, "taskQueue")
	}
	if c.WorkflowName == "" {
		missing = append(missing, "workflowName")
	}

	// The preprocessing workflow requires SharedPath to be set.
	if c.Type == enums.ChildWorkflowTypePreprocessing && c.SharedPath == "" {
		missing = append(missing, "sharedPath")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required value(s): %s", strings.Join(missing, ", "))
	}

	return nil
}

type Configs []Config

func (c Configs) ByType(t enums.ChildWorkflowType) *Config {
	for _, cfg := range c {
		if cfg.Type == t {
			return &cfg
		}
	}

	return nil
}

func (c Configs) Validate() error {
	var (
		types []enums.ChildWorkflowType
		errs  error
	)

	for i, cfg := range c {
		if err := cfg.Validate(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("child workflow[%d]: %w", i, err))
		}

		// Don't do duplicate check for empty types, as they are already
		// reported by the Validate() method above.
		if cfg.Type == "" {
			continue
		}

		// Ensure there are no duplicate child workflow types.
		if slices.Contains(types, cfg.Type) {
			errs = errors.Join(errs, fmt.Errorf("child workflow[%d]: duplicate type: %s", i, cfg.Type))
		} else {
			types = append(types, cfg.Type)
		}
	}

	return errs
}
