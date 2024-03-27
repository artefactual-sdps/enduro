package datatypes

import (
	"database/sql"

	"github.com/artefactual-sdps/enduro/internal/enums"
)

// PreservationTask represents a preservation action task in the
// preservation_task table.
type PreservationTask struct {
	ID                   uint                         `db:"id"`
	TaskID               string                       `db:"task_id"`
	Name                 string                       `db:"name"`
	Status               enums.PreservationTaskStatus `db:"status"`
	StartedAt            sql.NullTime                 `db:"started_at"`
	CompletedAt          sql.NullTime                 `db:"completed_at"`
	Note                 string
	PreservationActionID uint `db:"preservation_action_id"`
}
