package storage

import (
	"context"
	"time"

	"github.com/google/uuid"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func ReadAIPLocalActivity(
	ctx context.Context,
	storagesvc Service,
	aipID uuid.UUID,
) (*goastorage.AIP, error) {
	return storagesvc.ReadAip(ctx, aipID)
}

type UpdateAIPLocationLocalActivityParams struct {
	AIPID      uuid.UUID
	LocationID uuid.UUID
}

func UpdateAIPLocationLocalActivity(
	ctx context.Context,
	storagesvc Service,
	params *UpdateAIPLocationLocalActivityParams,
) error {
	return storagesvc.UpdateAipLocationID(ctx, params.AIPID, params.LocationID)
}

type UpdateAIPStatusLocalActivityParams struct {
	AIPID  uuid.UUID
	Status enums.AIPStatus
}

func UpdateAIPStatusLocalActivity(
	ctx context.Context,
	storagesvc Service,
	params *UpdateAIPStatusLocalActivityParams,
) error {
	return storagesvc.UpdateAipStatus(ctx, params.AIPID, params.Status)
}

type DeleteFromMinIOLocationLocalActivityParams struct {
	AIPID uuid.UUID
}

func DeleteFromMinIOLocationLocalActivity(
	ctx context.Context,
	storagesvc Service,
	params *DeleteFromMinIOLocationLocalActivityParams,
) error {
	return storagesvc.DeleteAip(ctx, params.AIPID)
}

type ReadLocationInfoLocalActivityResult struct {
	Source enums.LocationSource
	Config types.LocationConfig
}

func ReadLocationInfoLocalActivity(
	ctx context.Context,
	storagesvc Service,
	locationID uuid.UUID,
) (*ReadLocationInfoLocalActivityResult, error) {
	l, err := storagesvc.ReadLocation(ctx, locationID)
	if err != nil {
		return nil, err
	}

	// The location config from the Goa type cannot unmarshal.
	config, err := ConvertGoaLocationConfigToLocationConfig(l.Config)
	if err != nil {
		return nil, err
	}

	return &ReadLocationInfoLocalActivityResult{
		Source: enums.LocationSource(l.Source),
		Config: config,
	}, nil
}

type CreateWorkflowLocalActivityParams struct {
	AIPID      uuid.UUID
	TemporalID string
	Type       enums.WorkflowType
}

func CreateWorkflowLocalActivity(
	ctx context.Context,
	storagesvc Service,
	params *CreateWorkflowLocalActivityParams,
) (int, error) {
	w := &types.Workflow{
		UUID:       uuid.New(),
		TemporalID: params.TemporalID,
		Type:       params.Type,
		Status:     enums.WorkflowStatusInProgress,
		StartedAt:  time.Now(),
		AIPUUID:    params.AIPID,
	}
	err := storagesvc.CreateWorkflow(ctx, w)
	if err != nil {
		return 0, err
	}

	return w.DBID, nil
}

type UpdateWorkflowStatusLocalActivityParams struct {
	DBID   int
	Status enums.WorkflowStatus
}

func UpdateWorkflowStatusLocalActivity(
	ctx context.Context,
	storagesvc Service,
	params *UpdateWorkflowStatusLocalActivityParams,
) error {
	_, err := storagesvc.UpdateWorkflow(ctx, params.DBID, func(w *types.Workflow) (*types.Workflow, error) {
		w.Status = params.Status
		return w, nil
	})

	return err
}

type CompleteWorkflowLocalActivityParams struct {
	DBID   int
	Status enums.WorkflowStatus
}

func CompleteWorkflowLocalActivity(
	ctx context.Context,
	storagesvc Service,
	params *CompleteWorkflowLocalActivityParams,
) error {
	_, err := storagesvc.UpdateWorkflow(ctx, params.DBID, func(w *types.Workflow) (*types.Workflow, error) {
		w.Status = params.Status
		w.CompletedAt = time.Now()
		return w, nil
	})

	return err
}

type CreateTaskLocalActivityParams struct {
	WorkflowDBID int
	Status       enums.TaskStatus
	Name         string
	Note         string
}

func CreateTaskLocalActivity(
	ctx context.Context,
	storagesvc Service,
	params *CreateTaskLocalActivityParams,
) (int, error) {
	t := &types.Task{
		UUID:         uuid.New(),
		Name:         params.Name,
		Status:       params.Status,
		StartedAt:    time.Now(),
		Note:         params.Note,
		WorkflowDBID: params.WorkflowDBID,
	}
	err := storagesvc.CreateTask(ctx, t)
	if err != nil {
		return 0, err
	}

	return t.DBID, nil
}

type CompleteTaskLocalActivityParams struct {
	DBID   int
	Status enums.TaskStatus
	Note   string
}

func CompleteTaskLocalActivity(
	ctx context.Context,
	storagesvc Service,
	params *CompleteTaskLocalActivityParams,
) error {
	_, err := storagesvc.UpdateTask(ctx, params.DBID, func(t *types.Task) (*types.Task, error) {
		t.Status = params.Status
		t.CompletedAt = time.Now()
		t.Note = params.Note
		return t, nil
	})

	return err
}

type CreateDeletionRequestLocalActivityParams struct {
	Requester    string
	RequesterISS string
	RequesterSub string
	Reason       string
	AIPUUID      uuid.UUID
	WorkflowDBID int
}

func CreateDeletionRequestLocalActivity(
	ctx context.Context,
	storagesvc Service,
	params *CreateDeletionRequestLocalActivityParams,
) (int, error) {
	dr := &types.DeletionRequest{
		UUID:         uuid.New(),
		Requester:    params.Requester,
		RequesterISS: params.RequesterISS,
		RequesterSub: params.RequesterSub,
		RequestedAt:  time.Now(),
		Reason:       params.Reason,
		Status:       enums.DeletionRequestStatusPending,
		AIPUUID:      params.AIPUUID,
		WorkflowDBID: params.WorkflowDBID,
	}
	err := storagesvc.CreateDeletionRequest(ctx, dr)
	if err != nil {
		return 0, err
	}

	return dr.DBID, nil
}

func UpdateDeletionRequestLocalActivity(
	ctx context.Context,
	storagesvc Service,
	dbID int,
	review DeletionDecisionSignal,
) error {
	_, err := storagesvc.UpdateDeletionRequest(
		ctx,
		dbID,
		func(dr *types.DeletionRequest) (*types.DeletionRequest, error) {
			dr.Reviewer = review.UserEmail
			dr.ReviewerISS = review.UserISS
			dr.ReviewerSub = review.UserSub
			dr.ReviewedAt = time.Now()
			dr.Status = review.Status
			return dr, nil
		},
	)

	return err
}

func CancelDeletionRequestLocalActivity(
	ctx context.Context,
	storagesvc Service,
	dbID int,
) error {
	_, err := storagesvc.UpdateDeletionRequest(
		ctx,
		dbID,
		func(dr *types.DeletionRequest) (*types.DeletionRequest, error) {
			dr.Status = enums.DeletionRequestStatusCanceled
			return dr, nil
		},
	)

	return err
}
