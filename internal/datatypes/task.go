package datatypes

import (
	"database/sql"
	"strings"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/enums"
)

// Task represents a workflow task in the task table.
type Task struct {
	ID           int
	UUID         uuid.UUID
	Name         string
	Status       enums.TaskStatus
	StartedAt    sql.NullTime
	CompletedAt  sql.NullTime
	Note         string
	WorkflowUUID uuid.UUID
}

// SystemError indicates that a system error occurred during task execution.
func (t *Task) SystemError(notes ...string) {
	t.Status = enums.TaskStatusError
	t.Note = "System error: " + strings.Join(notes, "\n\n")
}

// Failed indicates the task failed due to the content not meeting one or more
// policy requirements.
func (t *Task) Failed(notes ...string) {
	t.Status = enums.TaskStatusFailed
	t.Note = "Content error: " + strings.Join(notes, "\n\n")
}
