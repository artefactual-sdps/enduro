package datatypes

import (
	"database/sql"

	"github.com/artefactual-sdps/enduro/internal/enums"
)

// PreservationAction represents a preservation action in the preservation_action table.
type PreservationAction struct {
	ID          int                            `db:"id"`
	WorkflowID  string                         `db:"workflow_id"`
	Type        enums.PreservationActionType   `db:"type"`
	Status      enums.PreservationActionStatus `db:"status"`
	StartedAt   sql.NullTime                   `db:"started_at"`
	CompletedAt sql.NullTime                   `db:"completed_at"`
	PackageID   int                            `db:"package_id"`
}
