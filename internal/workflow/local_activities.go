package workflow

import (
	"context"
	"database/sql"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
)

type createSIPLocalActivityParams struct {
	UUID   uuid.UUID
	Name   string
	Status enums.SIPStatus
}

func createSIPLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	params *createSIPLocalActivityParams,
) (int, error) {
	col := &datatypes.SIP{
		UUID:   params.UUID,
		Name:   params.Name,
		Status: params.Status,
	}

	if err := ingestsvc.CreateSIP(ctx, col); err != nil {
		return 0, err
	}

	return col.ID, nil
}

type updateSIPLocalActivityParams struct {
	UUID        uuid.UUID
	Name        string
	AIPUUID     string
	CompletedAt time.Time
	Status      enums.SIPStatus
}

type updateSIPLocalActivityResult struct{}

func updateSIPLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	params *updateSIPLocalActivityParams,
) (*updateSIPLocalActivityResult, error) {
	err := ingestsvc.UpdateSIP(
		ctx,
		params.UUID,
		params.Name,
		params.AIPUUID,
		params.Status,
		params.CompletedAt,
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
	sipUUID uuid.UUID,
	startedAt time.Time,
) (*setStatusInProgressLocalActivityResult, error) {
	return &setStatusInProgressLocalActivityResult{}, ingestsvc.SetStatusInProgress(ctx, sipUUID, startedAt)
}

type setStatusLocalActivityResult struct{}

func setStatusLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	sipUUID uuid.UUID,
	status enums.SIPStatus,
) (*setStatusLocalActivityResult, error) {
	return &setStatusLocalActivityResult{}, ingestsvc.SetStatus(ctx, sipUUID, status)
}

type createWorkflowLocalActivityParams struct {
	TemporalID  string
	Type        enums.WorkflowType
	Status      enums.WorkflowStatus
	StartedAt   time.Time
	CompletedAt time.Time
	SIPUUID     uuid.UUID
}

func createWorkflowLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	params *createWorkflowLocalActivityParams,
) (int, error) {
	w := datatypes.Workflow{
		TemporalID: params.TemporalID,
		Type:       params.Type,
		Status:     params.Status,
		SIPUUID:    params.SIPUUID,
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
