package datatypes

import (
	"database/sql"
	"strings"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/enums"
)

// Task represents a workflow task in the task table.
type Task struct {
	ID           int              `db:"id"`
	UUID         uuid.UUID        `db:"uuid"`
	Name         string           `db:"name"`
	Status       enums.TaskStatus `db:"status"`
	StartedAt    sql.NullTime     `db:"started_at"`
	CompletedAt  sql.NullTime     `db:"completed_at"`
	Note         string           `db:"note"`
	WorkflowUUID uuid.UUID        `db:"workflow_uuid"`
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
