package childwf

// Outcome is a status code indicating the overall result of the child
// workflow, with 0 being reserved for success, 1 indicating a critical system
// error, and 2 indicating a non-critical content error. Further outcome codes
// may be added in the future.
type Outcome int

const (
	OutcomeSuccess Outcome = iota
	OutcomeSystemError
	OutcomeContentError
)

type Result struct {
	// Outcome indicates the overall result of the child workflow. A value of 0
	// indicates success, while non-zero values indicate various error states.
	Outcome Outcome

	// Message provides additional context about the result of the child
	// workflow execution.
	Message string
}
