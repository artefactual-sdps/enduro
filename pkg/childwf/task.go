package childwf

import (
	"fmt"
	"time"
)

type Task struct {
	// Name is the name of the task.
	Name string

	// Message is a human readable description of the task operations or result.
	Message string

	// Outcome indicates the completion state of the task.
	Outcome TaskOutcome

	// StartedAt is the timestamp of the task initiation.
	StartedAt time.Time

	// CompletedAt is the timestamp of the task completion.
	CompletedAt time.Time
}

// NewTask creates a new Task with the given name and start time.
func NewTask(t time.Time, name string) *Task {
	return &Task{
		Name:      name,
		Outcome:   TaskOutcomeUnspecified,
		StartedAt: t,
	}
}

// Complete sets the task's completion time, outcome, and message.
func (e *Task) Complete(t time.Time, outcome TaskOutcome, msg string, a ...any) *Task {
	e.CompletedAt = t
	e.Outcome = outcome
	e.Message = fmt.Sprintf(msg, a...)

	return e
}

// Succeed sets the task's outcome to TaskOutcomeSuccess and updates the completion time and message.
func (e *Task) Succeed(t time.Time, msg string, a ...any) *Task {
	return e.Complete(t, TaskOutcomeSuccess, msg, a...)
}

// IsSuccess returns true if the task's outcome is TaskOutcomeSuccess.
func (e *Task) IsSuccess() bool {
	return e.Outcome == TaskOutcomeSuccess
}
