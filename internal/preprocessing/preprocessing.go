package preprocessing

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
