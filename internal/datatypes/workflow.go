package datatypes

import (
	"database/sql"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/enums"
)

type Workflow struct {
	ID          int
	UUID        uuid.UUID
	TemporalID  string
	Type        enums.WorkflowType
	Status      enums.WorkflowStatus
	StartedAt   sql.NullTime
	CompletedAt sql.NullTime
	SIPUUID     uuid.UUID
	Tasks       []*Task
}
