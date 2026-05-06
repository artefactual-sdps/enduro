package childwf

import "time"

type PostStorageParams struct {
	AIPUUID string

	// CustomMetadata is opaque metadata returned by earlier child workflows.
	CustomMetadata CustomMetadata
}

type PostStorageResult struct {
	// Outcome indicates the overall result of the child workflow. A value of 0
	// indicates success, while non-zero values indicate various error states.
	Outcome Outcome

	// CustomMetadata is opaque metadata to carry to later child workflows.
	CustomMetadata CustomMetadata

	// Tasks is a log of the tasks performed by poststorage.
	Tasks []*Task
}

// NewTask creates a new poststorage task and appends it to the result.
func (r *PostStorageResult) NewTask(t time.Time, name string) *Task {
	return r.result().newTask(t, name)
}

// ValidationError marks the poststorage result as a content error and completes
// task as a validation failure.
func (r *PostStorageResult) ValidationError(t time.Time, task *Task, msg ...string) {
	r.result().validationError(t, task, msg...)
}

// SystemError marks the poststorage result as a system error and completes task
// as a system failure.
func (r *PostStorageResult) SystemError(t time.Time, task *Task, msg ...string) {
	r.result().systemError(t, task, msg...)
}

func (r *PostStorageResult) result() result {
	return result{outcome: &r.Outcome, tasks: &r.Tasks}
}
