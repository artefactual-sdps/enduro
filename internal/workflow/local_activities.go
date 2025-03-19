package workflow

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	temporalsdk_activity "go.temporal.io/sdk/activity"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
)

type createSIPLocalActivityParams struct {
	Key    string
	Status enums.SIPStatus
}

func createSIPLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	params *createSIPLocalActivityParams,
) (int, error) {
	info := temporalsdk_activity.GetInfo(ctx)

	col := &datatypes.SIP{
		Name:       params.Key,
		WorkflowID: info.WorkflowExecution.ID,
		RunID:      info.WorkflowExecution.RunID,
		Status:     params.Status,
	}

	if err := ingestsvc.Create(ctx, col); err != nil {
		return 0, err
	}

	return col.ID, nil
}

type updateSIPLocalActivityParams struct {
	SIPID    int
	Key      string
	AIPUUID  string
	StoredAt time.Time
	Status   enums.SIPStatus
}

type updateSIPLocalActivityResult struct{}

func updateSIPLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	params *updateSIPLocalActivityParams,
) (*updateSIPLocalActivityResult, error) {
	info := temporalsdk_activity.GetInfo(ctx)

	err := ingestsvc.UpdateWorkflowStatus(
		ctx,
		params.SIPID,
		params.Key,
		info.WorkflowExecution.ID,
		info.WorkflowExecution.RunID,
		params.AIPUUID,
		params.Status,
		params.StoredAt,
	)
	if err != nil {
		return &updateSIPLocalActivityResult{}, err
	}

	return &updateSIPLocalActivityResult{}, nil
}

type setStatusInProgressLocalActivityResult struct{}

func setStatusInProgressLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	sipID int,
	startedAt time.Time,
) (*setStatusInProgressLocalActivityResult, error) {
	return &setStatusInProgressLocalActivityResult{}, ingestsvc.SetStatusInProgress(ctx, sipID, startedAt)
}

type setStatusLocalActivityResult struct{}

func setStatusLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	sipID int,
	status enums.SIPStatus,
) (*setStatusLocalActivityResult, error) {
	return &setStatusLocalActivityResult{}, ingestsvc.SetStatus(ctx, sipID, status)
}

type setLocationIDLocalActivityResult struct{}

func setLocationIDLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	sipID int,
	locationID uuid.UUID,
) (*setLocationIDLocalActivityResult, error) {
	return &setLocationIDLocalActivityResult{}, ingestsvc.SetLocationID(ctx, sipID, locationID)
}

type saveLocationMoveWorkflowLocalActivityParams struct {
	SIPID       int
	LocationID  uuid.UUID
	WorkflowID  string
	Type        enums.WorkflowType
	Status      enums.WorkflowStatus
	StartedAt   time.Time
	CompletedAt time.Time
}

type saveLocationMoveWorkflowLocalActivityResult struct{}

func saveLocationMoveWorkflowLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	params *saveLocationMoveWorkflowLocalActivityParams,
) (*saveLocationMoveWorkflowLocalActivityResult, error) {
	wID, err := createWorkflowLocalActivity(ctx, ingestsvc, &createWorkflowLocalActivityParams{
		WorkflowID:  params.WorkflowID,
		Type:        params.Type,
		Status:      params.Status,
		StartedAt:   params.StartedAt,
		CompletedAt: params.CompletedAt,
		SIPID:       params.SIPID,
	})
	if err != nil {
		return &saveLocationMoveWorkflowLocalActivityResult{}, err
	}

	actionStatusToTaskStatus := map[enums.WorkflowStatus]enums.TaskStatus{
		enums.WorkflowStatusUnspecified: enums.TaskStatusUnspecified,
		enums.WorkflowStatusDone:        enums.TaskStatusDone,
		enums.WorkflowStatusInProgress:  enums.TaskStatusInProgress,
		enums.WorkflowStatusError:       enums.TaskStatusError,
	}

	task := datatypes.Task{
		TaskID:     uuid.NewString(),
		Name:       "Move AIP",
		Status:     actionStatusToTaskStatus[params.Status],
		Note:       fmt.Sprintf("Moved to location %s", params.LocationID),
		WorkflowID: wID,
	}
	task.StartedAt.Time = params.StartedAt
	task.CompletedAt.Time = params.CompletedAt

	return &saveLocationMoveWorkflowLocalActivityResult{}, ingestsvc.CreateTask(ctx, &task)
}

type createWorkflowLocalActivityParams struct {
	WorkflowID  string
	Type        enums.WorkflowType
	Status      enums.WorkflowStatus
	StartedAt   time.Time
	CompletedAt time.Time
	SIPID       int
}

func createWorkflowLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	params *createWorkflowLocalActivityParams,
) (int, error) {
	w := datatypes.Workflow{
		WorkflowID: params.WorkflowID,
		Type:       params.Type,
		Status:     params.Status,
		SIPID:      params.SIPID,
	}
	if !params.StartedAt.IsZero() {
		w.StartedAt = sql.NullTime{Time: params.StartedAt, Valid: true}
	}
	if !params.CompletedAt.IsZero() {
		w.CompletedAt = sql.NullTime{Time: params.CompletedAt, Valid: true}
	}

	if err := ingestsvc.CreateWorkflow(ctx, &w); err != nil {
		return 0, err
	}

	return w.ID, nil
}

type setWorkflowStatusLocalActivityResult struct{}

func setWorkflowStatusLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	ID int,
	status enums.WorkflowStatus,
) (*setWorkflowStatusLocalActivityResult, error) {
	return &setWorkflowStatusLocalActivityResult{}, ingestsvc.SetWorkflowStatus(ctx, ID, status)
}

type completeWorkflowLocalActivityParams struct {
	WorkflowID  int
	Status      enums.WorkflowStatus
	CompletedAt time.Time
}

type completeWorkflowLocalActivityResult struct{}

func completeWorkflowLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	params *completeWorkflowLocalActivityParams,
) (*completeWorkflowLocalActivityResult, error) {
	return &completeWorkflowLocalActivityResult{}, ingestsvc.CompleteWorkflow(
		ctx,
		params.WorkflowID,
		params.Status,
		params.CompletedAt,
	)
}

type createTaskLocalActivityParams struct {
	Ingestsvc ingest.Service
	RNG       io.Reader
	Task      datatypes.Task
}

func createTaskLocalActivity(
	ctx context.Context,
	params *createTaskLocalActivityParams,
) (int, error) {
	task := params.Task
	if task.TaskID == "" {
		id, err := uuid.NewRandomFromReader(params.RNG)
		if err != nil {
			return 0, err
		}
		task.TaskID = id.String()
	}

	if err := params.Ingestsvc.CreateTask(ctx, &task); err != nil {
		return 0, err
	}

	return task.ID, nil
}

type completeTaskLocalActivityParams struct {
	ID          int
	Status      enums.TaskStatus
	CompletedAt time.Time
	Note        *string
}

type completeTaskLocalActivityResult struct{}

func completeTaskLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	params *completeTaskLocalActivityParams,
) (*completeTaskLocalActivityResult, error) {
	return &completeTaskLocalActivityResult{}, ingestsvc.CompleteTask(
		ctx,
		params.ID,
		params.Status,
		params.CompletedAt,
		params.Note,
	)
}
