package types

import (
	"time"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/storage/enums"
)

type Task struct {
	DBID         int
	UUID         uuid.UUID
	Name         string
	Status       enums.TaskStatus
	StartedAt    time.Time
	CompletedAt  time.Time
	Note         string
	WorkflowDBID int
}
