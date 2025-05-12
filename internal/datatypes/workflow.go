package datatypes

import (
	"database/sql"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/enums"
)

type Workflow struct {
	ID          int                  `db:"id"`
	TemporalID  string               `db:"temporal_id"`
	Type        enums.WorkflowType   `db:"type"`
	Status      enums.WorkflowStatus `db:"status"`
	StartedAt   sql.NullTime         `db:"started_at"`
	CompletedAt sql.NullTime         `db:"completed_at"`
	SIPUUID     uuid.UUID            `db:"sip_uuid"`
}
