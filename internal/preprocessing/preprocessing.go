package preprocessing

import "errors"

type Config struct {
	// Enable preprocessing child workflow.
	Enabled bool
	// Extract SIP in preprocessing.
	Extract bool
	// Local path shared between workers.
	SharedPath string
	// Temporal configuration.
	Temporal Temporal
}

type Temporal struct {
	Namespace    string
	TaskQueue    string
	WorkflowName string
}

type WorkflowParams struct {
	// Relative path to the shared path.
	RelativePath string
}

type Outcome int

const (
	OutcomeSuccess Outcome = iota
	OutcomeSystemError
	OutcomeContentError
)

type WorkflowResult struct {
	// Outcome is an integer indicating if the workflow completed successfully,
	// or with errors.
	Outcome Outcome

	// Relative path to the shared path.
	RelativePath string

	// PreservationTasks is a log of the tasks performed by preprocessing.
	PreservationTasks []Task
}

// Validate implements config.ConfigurationValidator.
func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.SharedPath == "" {
		return errors.New("sharedPath is required in the [preprocessing] configuration")
	}
	if c.Temporal.Namespace == "" || c.Temporal.TaskQueue == "" || c.Temporal.WorkflowName == "" {
		return errors.New(
			"namespace, taskQueue and workflowName are required in the [preprocessing.temporal] configuration",
		)
	}
	return nil
}
