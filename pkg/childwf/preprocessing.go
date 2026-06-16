package childwf

import (
	"time"

	"github.com/google/uuid"
)

type PreprocessingParams struct {
	// User contains non-sensitive information about the user who initiated the
	// workflow, when available.
	User *User

	// Relative path to the shared path.
	RelativePath string

	// SIPID is the identifier of the SIP being processed.
	SIPID uuid.UUID

	// BatchID is the identifier of the batch being processed. If the SIP is not
	// part of a batch, this will equal uuid.Nil.
	BatchID uuid.UUID

	// SIPName is the original filename of the SIP being processed.
	SIPName string
}

type PreprocessingResult struct {
	// Outcome is an integer indicating if the workflow completed successfully,
	// or with errors.
	Outcome Outcome

	// CustomMetadata is opaque metadata to carry to later child workflows.
	CustomMetadata CustomMetadata

	// Relative path to the shared path.
	RelativePath string

	// Tasks is a log of the tasks performed by preprocessing.
	Tasks []*Task
}

// NewTask creates a new preprocessing task and appends it to the result.
func (r *PreprocessingResult) NewTask(t time.Time, name string) *Task {
	return r.result().newTask(t, name)
}

// ValidationError marks the preprocessing result as a content error and
// completes task as a validation failure.
func (r *PreprocessingResult) ValidationError(t time.Time, task *Task, msg ...string) {
	r.result().validationError(t, task, msg...)
}

// SystemError marks the preprocessing result as a system error and completes
// task as a system failure.
func (r *PreprocessingResult) SystemError(t time.Time, task *Task, msg ...string) {
	r.result().systemError(t, task, msg...)
}

func (r *PreprocessingResult) result() result {
	return result{outcome: &r.Outcome, tasks: &r.Tasks}
}
