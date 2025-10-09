package workflow

import (
	"context"
	"database/sql"
	"fmt"
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
	FailedAs    enums.SIPFailedAs
	FailedKey   string
}

type updateSIPLocalActivityResult struct{}

func updateSIPLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	params *updateSIPLocalActivityParams,
) (*updateSIPLocalActivityResult, error) {
	_, err := ingestsvc.UpdateSIP(
		ctx,
		params.UUID,
		func(s *datatypes.SIP) (*datatypes.SIP, error) {
			s.Name = params.Name
			s.Status = params.Status
			s.FailedAs = params.FailedAs
			s.FailedKey = params.FailedKey

			if !params.Status.IsValid() {
				return nil, fmt.Errorf("invalid status: %s", params.Status)
			}

			if params.FailedAs != "" && !params.FailedAs.IsValid() {
				return nil, fmt.Errorf("invalid failed as: %s", params.FailedAs)
			}

			if params.AIPUUID != "" {
				aipUUID, err := uuid.Parse(params.AIPUUID)
				if err != nil {
					return nil, fmt.Errorf("invalid AIP UUID: %s", params.AIPUUID)
				}
				s.AIPID = uuid.NullUUID{Valid: true, UUID: aipUUID}
			}

			if !params.CompletedAt.IsZero() {
				s.CompletedAt = sql.NullTime{Valid: true, Time: params.CompletedAt}
			}

			return s, nil
		},
	)
	if err != nil {
		return nil, err
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
	// RNG is the source of randomness for generating the workflow UUID.
	RNG         io.Reader
	TemporalID  string
	Type        enums.WorkflowType
	Status      enums.WorkflowStatus
	StartedAt   time.Time
	CompletedAt time.Time
	SIPUUID     uuid.UUID
}

type createWorkflowLocalActivityResult struct {
	ID   int
	UUID uuid.UUID
}

func createWorkflowLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	params *createWorkflowLocalActivityParams,
) (*createWorkflowLocalActivityResult, error) {
	wfUUID, err := uuid.NewRandomFromReader(params.RNG)
	if err != nil {
		return nil, fmt.Errorf("generate workflow UUID: %w", err)
	}

	w := datatypes.Workflow{
		UUID:       wfUUID,
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
		return nil, err
	}

	return &createWorkflowLocalActivityResult{ID: w.ID, UUID: w.UUID}, nil
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
	Task      *datatypes.Task
}

func createTaskLocalActivity(
	ctx context.Context,
	params *createTaskLocalActivityParams,
) (int, error) {
	task := params.Task

	uid, err := uuid.NewRandomFromReader(params.RNG)
	if err != nil {
		return 0, fmt.Errorf("generate task UUID: %v", err)
	}
	task.UUID = uid

	task.StartedAt = sql.NullTime{
		Time:  time.Now().UTC(),
		Valid: true,
	}

	if err := params.Ingestsvc.CreateTask(ctx, params.Task); err != nil {
		return 0, err
	}

	return params.Task.ID, nil
}

type completeTaskLocalActivityParams struct {
	ID     int
	Status enums.TaskStatus
	Note   *string
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
		time.Now().UTC(),
		params.Note,
	)
}
