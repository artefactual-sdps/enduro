package datatypes

import (
	"database/sql"

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
