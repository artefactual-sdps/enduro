package persistence

import (
	"context"

	"github.com/google/uuid"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

type (
	WorkflowUpdater        func(*types.Workflow) (*types.Workflow, error)
	TaskUpdater            func(*types.Task) (*types.Task, error)
	DeletionRequestUpdater func(*types.DeletionRequest) (*types.DeletionRequest, error)
)

type Storage interface {
	// AIP.
	CreateAIP(ctx context.Context, aip *goastorage.AIP) (*goastorage.AIP, error)
	ListAIPs(ctx context.Context, payload *goastorage.ListAipsPayload) (*goastorage.AIPs, error)
	ReadAIP(ctx context.Context, aipID uuid.UUID) (*goastorage.AIP, error)
	UpdateAIPStatus(ctx context.Context, aipID uuid.UUID, status enums.AIPStatus) error
	UpdateAIPLocationID(ctx context.Context, aipID, locationID uuid.UUID) error
	ListWorkflows(ctx context.Context, f *WorkflowFilter) (goastorage.AIPWorkflowCollection, error)

	// Location.
	CreateLocation(
		ctx context.Context,
		location *goastorage.Location,
		config *types.LocationConfig,
	) (*goastorage.Location, error)
	ListLocations(ctx context.Context) (goastorage.LocationCollection, error)
	ReadLocation(ctx context.Context, locationID uuid.UUID) (*goastorage.Location, error)
	LocationAIPs(ctx context.Context, locationID uuid.UUID) (goastorage.AIPCollection, error)

	// Workflow.
	CreateWorkflow(context.Context, *types.Workflow) error
	UpdateWorkflow(context.Context, int, WorkflowUpdater) (*types.Workflow, error)

	// Task.
	CreateTask(context.Context, *types.Task) error
	UpdateTask(context.Context, int, TaskUpdater) (*types.Task, error)

	// DeletionRequest.
	CreateDeletionRequest(context.Context, *types.DeletionRequest) error
	UpdateDeletionRequest(context.Context, int, DeletionRequestUpdater) (*types.DeletionRequest, error)
	ReadAipPendingDeletionRequest(context.Context, uuid.UUID) (*types.DeletionRequest, error)
}

type WorkflowFilter struct {
	AIPUUID *uuid.UUID
	Status  *enums.WorkflowStatus
	Type    *enums.WorkflowType
}
