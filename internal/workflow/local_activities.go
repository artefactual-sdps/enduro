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

type saveLocationMovePreservationActionLocalActivityParams struct {
	SIPID       int
	LocationID  uuid.UUID
	WorkflowID  string
	Type        enums.PreservationActionType
	Status      enums.PreservationActionStatus
	StartedAt   time.Time
	CompletedAt time.Time
}

type saveLocationMovePreservationActionLocalActivityResult struct{}

func saveLocationMovePreservationActionLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	params *saveLocationMovePreservationActionLocalActivityParams,
) (*saveLocationMovePreservationActionLocalActivityResult, error) {
	paID, err := createPreservationActionLocalActivity(ctx, ingestsvc, &createPreservationActionLocalActivityParams{
		WorkflowID:  params.WorkflowID,
		Type:        params.Type,
		Status:      params.Status,
		StartedAt:   params.StartedAt,
		CompletedAt: params.CompletedAt,
		SIPID:       params.SIPID,
	})
	if err != nil {
		return &saveLocationMovePreservationActionLocalActivityResult{}, err
	}

	actionStatusToTaskStatus := map[enums.PreservationActionStatus]enums.PreservationTaskStatus{
		enums.PreservationActionStatusUnspecified: enums.PreservationTaskStatusUnspecified,
		enums.PreservationActionStatusDone:        enums.PreservationTaskStatusDone,
		enums.PreservationActionStatusInProgress:  enums.PreservationTaskStatusInProgress,
		enums.PreservationActionStatusError:       enums.PreservationTaskStatusError,
	}

	pt := datatypes.PreservationTask{
		TaskID:               uuid.NewString(),
		Name:                 "Move AIP",
		Status:               actionStatusToTaskStatus[params.Status],
		Note:                 fmt.Sprintf("Moved to location %s", params.LocationID),
		PreservationActionID: paID,
	}
	pt.StartedAt.Time = params.StartedAt
	pt.CompletedAt.Time = params.CompletedAt

	return &saveLocationMovePreservationActionLocalActivityResult{}, ingestsvc.CreatePreservationTask(ctx, &pt)
}

type createPreservationActionLocalActivityParams struct {
	WorkflowID  string
	Type        enums.PreservationActionType
	Status      enums.PreservationActionStatus
	StartedAt   time.Time
	CompletedAt time.Time
	SIPID       int
}

func createPreservationActionLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	params *createPreservationActionLocalActivityParams,
) (int, error) {
	pa := datatypes.PreservationAction{
		WorkflowID: params.WorkflowID,
		Type:       params.Type,
		Status:     params.Status,
		SIPID:      params.SIPID,
	}
	if !params.StartedAt.IsZero() {
		pa.StartedAt = sql.NullTime{Time: params.StartedAt, Valid: true}
	}
	if !params.CompletedAt.IsZero() {
		pa.CompletedAt = sql.NullTime{Time: params.CompletedAt, Valid: true}
	}

	if err := ingestsvc.CreatePreservationAction(ctx, &pa); err != nil {
		return 0, err
	}

	return pa.ID, nil
}

type setPreservationActionStatusLocalActivityResult struct{}

func setPreservationActionStatusLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	ID int,
	status enums.PreservationActionStatus,
) (*setPreservationActionStatusLocalActivityResult, error) {
	return &setPreservationActionStatusLocalActivityResult{}, ingestsvc.SetPreservationActionStatus(ctx, ID, status)
}

type completePreservationActionLocalActivityParams struct {
	PreservationActionID int
	Status               enums.PreservationActionStatus
	CompletedAt          time.Time
}

type completePreservationActionLocalActivityResult struct{}

func completePreservationActionLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	params *completePreservationActionLocalActivityParams,
) (*completePreservationActionLocalActivityResult, error) {
	return &completePreservationActionLocalActivityResult{}, ingestsvc.CompletePreservationAction(
		ctx,
		params.PreservationActionID,
		params.Status,
		params.CompletedAt,
	)
}

type createPreservationTaskLocalActivityParams struct {
	Ingestsvc        ingest.Service
	RNG              io.Reader
	PreservationTask datatypes.PreservationTask
}

func createPreservationTaskLocalActivity(
	ctx context.Context,
	params *createPreservationTaskLocalActivityParams,
) (int, error) {
	pt := params.PreservationTask
	if pt.TaskID == "" {
		id, err := uuid.NewRandomFromReader(params.RNG)
		if err != nil {
			return 0, err
		}
		pt.TaskID = id.String()
	}

	if err := params.Ingestsvc.CreatePreservationTask(ctx, &pt); err != nil {
		return 0, err
	}

	return pt.ID, nil
}

type completePreservationTaskLocalActivityParams struct {
	ID          int
	Status      enums.PreservationTaskStatus
	CompletedAt time.Time
	Note        *string
}

type completePreservationTaskLocalActivityResult struct{}

func completePreservationTaskLocalActivity(
	ctx context.Context,
	ingestsvc ingest.Service,
	params *completePreservationTaskLocalActivityParams,
) (*completePreservationTaskLocalActivityResult, error) {
	return &completePreservationTaskLocalActivityResult{}, ingestsvc.CompletePreservationTask(
		ctx,
		params.ID,
		params.Status,
		params.CompletedAt,
		params.Note,
	)
}
