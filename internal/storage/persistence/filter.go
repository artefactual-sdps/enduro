package persistence

import (
	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/storage/enums"
)

type DeletionRequestFilter struct {
	AIPUUID *uuid.UUID
	Status  *enums.DeletionRequestStatus
}

type WorkflowFilter struct {
	AIPUUID *uuid.UUID
	Status  *enums.WorkflowStatus
	Type    *enums.WorkflowType
}
