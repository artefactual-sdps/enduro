package preprocessing

import (
	"time"

	"github.com/artefactual-sdps/enduro/internal/enums"
)

type Task struct {
	// Name is the name of the task.
	Name string

	// Message is a human readable description of the task operations or result.
	Message string

	// Outcome indicates the completion state of the task.
	Outcome enums.PreprocessingTaskOutcome

	// StartedAt is the timestamp of the task initiation.
	StartedAt time.Time

	// StartedAt is the timestamp of the task completion.
	CompletedAt time.Time
}
