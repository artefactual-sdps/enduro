package datatypes

import (
	"database/sql"
	"strings"

	"github.com/artefactual-sdps/enduro/internal/enums"
)

// Task represents a workflow task in the task table.
type Task struct {
	ID          int              `db:"id"`
	TaskID      string           `db:"task_id"`
	Name        string           `db:"name"`
	Status      enums.TaskStatus `db:"status"`
	StartedAt   sql.NullTime     `db:"started_at"`
	CompletedAt sql.NullTime     `db:"completed_at"`
	Note        string           `db:"note"`
	WorkflowID  int              `db:"workflow_id"`
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
