package datatypes

import (
	"database/sql"

	"github.com/artefactual-sdps/enduro/internal/enums"
)

// Workflow represents a Workflow in the workflow table.
type Workflow struct {
	ID          int                  `db:"id"`
	WorkflowID  string               `db:"workflow_id"`
	Type        enums.WorkflowType   `db:"type"`
	Status      enums.WorkflowStatus `db:"status"`
	StartedAt   sql.NullTime         `db:"started_at"`
	CompletedAt sql.NullTime         `db:"completed_at"`
	SIPID       int                  `db:"sip_id"`
}
