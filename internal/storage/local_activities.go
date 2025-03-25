package storage

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

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

type DeleteFromLocationLocalActivityParams struct {
	AIPID uuid.UUID
}

func DeleteFromLocationLocalActivity(
	ctx context.Context,
	storagesvc Service,
	params *DeleteFromLocationLocalActivityParams,
) error {
	return storagesvc.DeleteAip(ctx, params.AIPID)
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
	if err != nil {
		return err
	}

	return nil
}

type CreateTaskLocalActivityParams struct {
	WorkflowDBID int
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
		Status:       enums.TaskStatusInProgress,
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
	if err != nil {
		return err
	}

	return nil
}
