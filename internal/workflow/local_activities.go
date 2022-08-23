package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	temporalsdk_activity "go.temporal.io/sdk/activity"

	"github.com/artefactual-sdps/enduro/internal/package_"
)

type createPackageLocalActivityParams struct {
	Key    string
	Status package_.Status
}

func createPackageLocalActivity(ctx context.Context, logger logr.Logger, pkgsvc package_.Service, params *createPackageLocalActivityParams) (uint, error) {
	info := temporalsdk_activity.GetInfo(ctx)

	col := &package_.Package{
		Name:       params.Key,
		WorkflowID: info.WorkflowExecution.ID,
		RunID:      info.WorkflowExecution.RunID,
		Status:     params.Status,
	}

	if err := pkgsvc.Create(ctx, col); err != nil {
		logger.Error(err, "Error creating package")
		return 0, err
	}

	return col.ID, nil
}

type updatePackageLocalActivityParams struct {
	PackageID uint
	Key       string
	SIPID     string
	StoredAt  time.Time
	Status    package_.Status
}

func updatePackageLocalActivity(ctx context.Context, logger logr.Logger, pkgsvc package_.Service, params *updatePackageLocalActivityParams) error {
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
		logger.Error(err, "Error updating package")
		return err
	}

	return nil
}

func setStatusInProgressLocalActivity(ctx context.Context, pkgsvc package_.Service, pkgID uint, startedAt time.Time) error {
	return pkgsvc.SetStatusInProgress(ctx, pkgID, startedAt)
}

func setStatusLocalActivity(ctx context.Context, pkgsvc package_.Service, pkgID uint, status package_.Status) error {
	return pkgsvc.SetStatus(ctx, pkgID, status)
}

func setLocationLocalActivity(ctx context.Context, pkgsvc package_.Service, pkgID uint, locationID uuid.UUID) error {
	return pkgsvc.SetLocationID(ctx, pkgID, locationID)
}

type saveLocationMovePreservationActionLocalActivityParams struct {
	PackageID   uint
	Location    string
	WorkflowID  string
	Type        package_.PreservationActionType
	Status      package_.PreservationActionStatus
	StartedAt   time.Time
	CompletedAt time.Time
}

func saveLocationMovePreservationActionLocalActivity(ctx context.Context, pkgsvc package_.Service, params *saveLocationMovePreservationActionLocalActivityParams) error {
	paID, err := createPreservationActionLocalActivity(ctx, pkgsvc, &createPreservationActionLocalActivityParams{
		WorkflowID:  params.WorkflowID,
		Type:        params.Type,
		Status:      params.Status,
		StartedAt:   params.StartedAt,
		CompletedAt: params.CompletedAt,
		PackageID:   params.PackageID,
	})
	if err != nil {
		return err
	}

	actionStatusToTaskStatus := map[package_.PreservationActionStatus]package_.PreservationTaskStatus{
		package_.ActionStatusUnspecified: package_.TaskStatusUnspecified,
		package_.ActionStatusDone:        package_.TaskStatusDone,
		package_.ActionStatusInProgress:  package_.TaskStatusInProgress,
		package_.ActionStatusError:       package_.TaskStatusError,
	}

	pt := package_.PreservationTask{
		TaskID:               uuid.NewString(),
		Name:                 "Move AIP",
		Status:               actionStatusToTaskStatus[params.Status],
		Note:                 fmt.Sprintf("Moved to %s", params.Location),
		PreservationActionID: paID,
	}
	pt.StartedAt.Time = params.StartedAt
	pt.CompletedAt.Time = params.CompletedAt

	return pkgsvc.CreatePreservationTask(ctx, &pt)
}

type createPreservationActionLocalActivityParams struct {
	WorkflowID  string
	Type        package_.PreservationActionType
	Status      package_.PreservationActionStatus
	StartedAt   time.Time
	CompletedAt time.Time
	PackageID   uint
}

func createPreservationActionLocalActivity(ctx context.Context, pkgsvc package_.Service, params *createPreservationActionLocalActivityParams) (uint, error) {
	pa := package_.PreservationAction{
		WorkflowID: params.WorkflowID,
		Type:       params.Type,
		Status:     params.Status,
		PackageID:  params.PackageID,
	}
	pa.StartedAt.Time = params.StartedAt
	pa.CompletedAt.Time = params.CompletedAt

	if err := pkgsvc.CreatePreservationAction(ctx, &pa); err != nil {
		return 0, err
	}

	return pa.ID, nil
}

func setPreservationActonStatusLocalActivity(ctx context.Context, pkgsvc package_.Service, ID uint, status package_.PreservationActionStatus) error {
	return pkgsvc.SetPreservationActionStatus(ctx, ID, status)
}

type completePreservationActionLocalActivityParams struct {
	PreservationActionID uint
	Status               package_.PreservationActionStatus
	CompletedAt          time.Time
}

func completePreservationActionLocalActivity(ctx context.Context, pkgsvc package_.Service, params *completePreservationActionLocalActivityParams) error {
	return pkgsvc.CompletePreservationAction(ctx, params.PreservationActionID, params.Status, params.CompletedAt)
}

type createPreservationTaskLocalActivityParams struct {
	TaskID               string
	Name                 string
	Status               package_.PreservationTaskStatus
	StartedAt            time.Time
	CompletedAt          time.Time
	Note                 string
	PreservationActionID uint
}

func createPreservationTaskLocalActivity(ctx context.Context, pkgsvc package_.Service, params *createPreservationTaskLocalActivityParams) (uint, error) {
	pt := package_.PreservationTask{
		TaskID:               params.TaskID,
		Name:                 params.Name,
		Status:               params.Status,
		Note:                 params.Note,
		PreservationActionID: params.PreservationActionID,
	}
	pt.StartedAt.Time = params.StartedAt
	pt.CompletedAt.Time = params.CompletedAt

	if err := pkgsvc.CreatePreservationTask(ctx, &pt); err != nil {
		return 0, err
	}

	return pt.ID, nil
}

type completePreservationTaskLocalActivityParams struct {
	ID          uint
	Status      package_.PreservationTaskStatus
	CompletedAt time.Time
	Note        *string
}

func completePreservationTaskLocalActivity(ctx context.Context, pkgsvc package_.Service, params *completePreservationTaskLocalActivityParams) error {
	return pkgsvc.CompletePreservationTask(ctx, params.ID, params.Status, params.CompletedAt, params.Note)
}
