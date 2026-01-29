package childwf

type ResultStatus int

// ResultStatus is a status code indicating the overall result of the child
// workflow, with zero (0) being reserved for error, and one (1) indicating
// success. Additional status codes may be added in the future.
const (
	Error ResultStatus = iota
	Success
)

type Result struct {
	// Status indicates the overall result of the child workflow. An error
	// status (0) indicates that one or more events failed, but the child
	// workflow itself completed. If a critical error occurs during workflow
	// execution, the workflow should return a nil result and a standard error
	// instead.
	Status ResultStatus

	// Message provides additional context about the result of the child
	// workflow execution.
	Message string
}
