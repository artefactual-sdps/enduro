package childwf

import "time"

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
