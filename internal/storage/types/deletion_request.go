package types

import (
	"time"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/storage/enums"
)

type DeletionRequest struct {
	DBID         int
	UUID         uuid.UUID
	Requester    string
	RequesterISS string
	RequesterSub string
	Reviewer     string
	ReviewerISS  string
	ReviewerSub  string
	Reason       string
	Status       enums.DeletionRequestStatus
	RequestedAt  time.Time
	ReviewedAt   time.Time
	AIPUUID      uuid.UUID
	WorkflowDBID int
}
