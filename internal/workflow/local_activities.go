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
	"github.com/artefactual-sdps/enduro/internal/package_"
)

type createPackageLocalActivityParams struct {
	Key    string
	Status enums.PackageStatus
}

func createPackageLocalActivity(
	ctx context.Context,
	pkgsvc package_.Service,
	params *createPackageLocalActivityParams,
) (int, error) {
	info := temporalsdk_activity.GetInfo(ctx)

	col := &datatypes.Package{
		Name:       params.Key,
		WorkflowID: info.WorkflowExecution.ID,
		RunID:      info.WorkflowExecution.RunID,
		Status:     params.Status,
	}

	if err := pkgsvc.Create(ctx, col); err != nil {
		return 0, err
	}

	return col.ID, nil
}

type updatePackageLocalActivityParams struct {
	PackageID int
	Key       string
	SIPID     string
	StoredAt  time.Time
	Status    enums.PackageStatus
}

type updatePackageLocalActivityResult struct{}

func updatePackageLocalActivity(
	ctx context.Context,
	pkgsvc package_.Service,
	params *updatePackageLocalActivityParams,
) (*updatePackageLocalActivityResult, error) {
	info := temporalsdk_activity.GetInfo(ctx)

	err := pkgsvc.UpdateWorkflowStatus(
		ctx,
		params.PackageID,
		params.Key,
		info.WorkflowExecution.ID,
		info.WorkflowExecution.RunID,
		params.SIPID,
		params.Status,
		params.StoredAt,
	)
	if err != nil {
		return &updatePackageLocalActivityResult{}, err
	}

	return &updatePackageLocalActivityResult{}, nil
}

type setStatusInProgressLocalActivityResult struct{}

func setStatusInProgressLocalActivity(
	ctx context.Context,
	pkgsvc package_.Service,
	pkgID int,
	startedAt time.Time,
) (*setStatusInProgressLocalActivityResult, error) {
	return &setStatusInProgressLocalActivityResult{}, pkgsvc.SetStatusInProgress(ctx, pkgID, startedAt)
}

type setStatusLocalActivityResult struct{}

func setStatusLocalActivity(
	ctx context.Context,
	pkgsvc package_.Service,
	pkgID int,
	status enums.PackageStatus,
) (*setStatusLocalActivityResult, error) {
	return &setStatusLocalActivityResult{}, pkgsvc.SetStatus(ctx, pkgID, status)
}

type setLocationIDLocalActivityResult struct{}

func setLocationIDLocalActivity(
	ctx context.Context,
	pkgsvc package_.Service,
	pkgID int,
	locationID uuid.UUID,
) (*setLocationIDLocalActivityResult, error) {
	return &setLocationIDLocalActivityResult{}, pkgsvc.SetLocationID(ctx, pkgID, locationID)
}

type saveLocationMovePreservationActionLocalActivityParams struct {
	PackageID   int
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
	pkgsvc package_.Service,
	params *saveLocationMovePreservationActionLocalActivityParams,
) (*saveLocationMovePreservationActionLocalActivityResult, error) {
	paID, err := createPreservationActionLocalActivity(ctx, pkgsvc, &createPreservationActionLocalActivityParams{
		WorkflowID:  params.WorkflowID,
		Type:        params.Type,
		Status:      params.Status,
		StartedAt:   params.StartedAt,
		CompletedAt: params.CompletedAt,
		PackageID:   params.PackageID,
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

	return &saveLocationMovePreservationActionLocalActivityResult{}, pkgsvc.CreatePreservationTask(ctx, &pt)
}

type createPreservationActionLocalActivityParams struct {
	WorkflowID  string
	Type        enums.PreservationActionType
	Status      enums.PreservationActionStatus
	StartedAt   time.Time
	CompletedAt time.Time
	PackageID   int
}

func createPreservationActionLocalActivity(
	ctx context.Context,
	pkgsvc package_.Service,
	params *createPreservationActionLocalActivityParams,
) (int, error) {
	pa := datatypes.PreservationAction{
		WorkflowID: params.WorkflowID,
		Type:       params.Type,
		Status:     params.Status,
		PackageID:  params.PackageID,
	}
	if !params.StartedAt.IsZero() {
		pa.StartedAt = sql.NullTime{Time: params.StartedAt, Valid: true}
	}
	if !params.CompletedAt.IsZero() {
		pa.CompletedAt = sql.NullTime{Time: params.CompletedAt, Valid: true}
	}

	if err := pkgsvc.CreatePreservationAction(ctx, &pa); err != nil {
		return 0, err
	}

	return pa.ID, nil
}

type setPreservationActionStatusLocalActivityResult struct{}

func setPreservationActionStatusLocalActivity(
	ctx context.Context,
	pkgsvc package_.Service,
	ID int,
	status enums.PreservationActionStatus,
) (*setPreservationActionStatusLocalActivityResult, error) {
	return &setPreservationActionStatusLocalActivityResult{}, pkgsvc.SetPreservationActionStatus(ctx, ID, status)
}

type completePreservationActionLocalActivityParams struct {
	PreservationActionID int
	Status               enums.PreservationActionStatus
	CompletedAt          time.Time
}

type completePreservationActionLocalActivityResult struct{}

func completePreservationActionLocalActivity(
	ctx context.Context,
	pkgsvc package_.Service,
	params *completePreservationActionLocalActivityParams,
) (*completePreservationActionLocalActivityResult, error) {
	return &completePreservationActionLocalActivityResult{}, pkgsvc.CompletePreservationAction(
		ctx,
		params.PreservationActionID,
		params.Status,
		params.CompletedAt,
	)
}

type createPreservationTaskLocalActivityParams struct {
	PkgSvc           package_.Service
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

	if err := params.PkgSvc.CreatePreservationTask(ctx, &pt); err != nil {
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
	pkgsvc package_.Service,
	params *completePreservationTaskLocalActivityParams,
) (*completePreservationTaskLocalActivityResult, error) {
	return &completePreservationTaskLocalActivityResult{}, pkgsvc.CompletePreservationTask(
		ctx,
		params.ID,
		params.Status,
		params.CompletedAt,
		params.Note,
	)
}
