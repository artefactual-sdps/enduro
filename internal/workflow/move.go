package workflow

import (
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

type MoveWorkflow struct {
	ingestsvc ingest.Service
}

func NewMoveWorkflow(ingestsvc ingest.Service) *MoveWorkflow {
	return &MoveWorkflow{
		ingestsvc: ingestsvc,
	}
}

func (w *MoveWorkflow) Execute(ctx temporalsdk_workflow.Context, req *ingest.MoveWorkflowRequest) error {
	// Save starting time for preservation action.
	startedAt := temporalsdk_workflow.Now(ctx).UTC()

	// Assume the preservation action will be successful.
	status := enums.PreservationActionStatusDone

	// Set SIP to in progress status.
	{
		ctx := withLocalActivityOpts(ctx)
		err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setStatusLocalActivity, w.ingestsvc, req.ID, enums.SIPStatusInProgress).
			Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	// Move AIP to permanent storage
	{
		activityOpts := withActivityOptsForRequest(ctx)
		err := temporalsdk_workflow.ExecuteActivity(activityOpts, activities.MoveToPermanentStorageActivityName, &activities.MoveToPermanentStorageActivityParams{
			AIPID:      req.AIPID,
			LocationID: req.LocationID,
		}).
			Get(activityOpts, nil)
		if err != nil {
			status = enums.PreservationActionStatusError
		}
	}

	// Poll AIP move to permanent storage
	{
		if status != enums.PreservationActionStatusError {
			activityOpts := withActivityOptsForLongLivedRequest(ctx)
			err := temporalsdk_workflow.ExecuteActivity(activityOpts, activities.PollMoveToPermanentStorageActivityName, &activities.PollMoveToPermanentStorageActivityParams{
				AIPID: req.AIPID,
			}).
				Get(activityOpts, nil)
			if err != nil {
				status = enums.PreservationActionStatusError
			}
		}
	}

	// TODO: add compensation to these next activities?

	// Set SIP to done status.
	{
		ctx := withLocalActivityOpts(ctx)
		err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setStatusLocalActivity, w.ingestsvc, req.ID, enums.SIPStatusDone).
			Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	// Set SIP location.
	{
		if status != enums.PreservationActionStatusError {
			ctx := withLocalActivityOpts(ctx)
			err := temporalsdk_workflow.ExecuteLocalActivity(ctx, setLocationIDLocalActivity, w.ingestsvc, req.ID, req.LocationID).
				Get(ctx, nil)
			if err != nil {
				return err
			}
		}
	}

	// Create preservation action.
	{
		ctx := withLocalActivityOpts(ctx)
		completedAt := temporalsdk_workflow.Now(ctx).UTC()
		err := temporalsdk_workflow.ExecuteLocalActivity(ctx, saveLocationMovePreservationActionLocalActivity, w.ingestsvc, &saveLocationMovePreservationActionLocalActivityParams{
			SIPID:       req.ID,
			LocationID:  req.LocationID,
			WorkflowID:  temporalsdk_workflow.GetInfo(ctx).WorkflowExecution.ID,
			Type:        enums.PreservationActionTypeMovePackage,
			Status:      status,
			StartedAt:   startedAt,
			CompletedAt: completedAt,
		}).
			Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
