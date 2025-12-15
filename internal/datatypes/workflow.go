package datatypes

import (
	"database/sql"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/enums"
)

// Workflow represents a workflow execution associated with a SIP.
//
// Workflows track the execution of processing pipelines (e.g., ingest, move)
// and their associated tasks. A SIP may have multiple workflows over its
// lifecycle.
type Workflow struct {
	ID          int
	UUID        uuid.UUID
	TemporalID  string
	Type        enums.WorkflowType
	Status      enums.WorkflowStatus
	StartedAt   sql.NullTime
	CompletedAt sql.NullTime
	SIPUUID     uuid.UUID

	// Tasks contains the workflow's tasks, or nil if they were not loaded.
	Tasks []*Task
}
