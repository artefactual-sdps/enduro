package types

import (
	"time"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/storage/enums"
)

type Workflow struct {
	DBID        int
	UUID        uuid.UUID
	TemporalID  string
	Type        enums.WorkflowType
	Status      enums.WorkflowStatus
	StartedAt   time.Time
	CompletedAt time.Time
	AIPUUID     uuid.UUID
}
