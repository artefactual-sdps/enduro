package workflow

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	temporalapi_enums "go.temporal.io/api/enums/v1"
	temporalsdk_client "go.temporal.io/sdk/client"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-sdps/enduro/internal/batch"
	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

type BatchWorkflow struct {
	cfg       config.Configuration
	rng       io.Reader
	ingestsvc ingest.Service
	tc        temporalsdk_client.Client
}

func NewBatchWorkflow(
	cfg config.Configuration,
	rng io.Reader,
	ingestsvc ingest.Service,
	tc temporalsdk_client.Client,
) *BatchWorkflow {
	return &BatchWorkflow{
		cfg:       cfg,
		rng:       rng,
		ingestsvc: ingestsvc,
		tc:        tc,
	}
}

func (w *BatchWorkflow) Execute(ctx temporalsdk_workflow.Context, req *ingest.BatchWorkflowRequest) (e error) {
	state := newBatchWorkflowState(ctx, req)
	state.logger.Info(
		"Starting batch workflow",
		"uuid", req.Batch.UUID.String(),
		"identifier", req.Batch.Identifier,
		"source_id", req.SIPSourceID.String(),
		"keys", req.Keys,
	)

	defer func() {
		// Update final batch status and add completion date.
		state.batch.CompletedAt = temporalsdk_workflow.Now(ctx)
		state.batch.Status = enums.BatchStatusIngested
		if e != nil {
			state.batch.Status = enums.BatchStatusFailed
		}
		if err := w.updateBatch(ctx, state); err != nil {
			e = errors.Join(e, err)
		}

		state.logger.Info(
			"Batch workflow completed",
			"uuid", state.batch.UUID.String(),
			"identifier", state.batch.Identifier,
			"status", state.batch.Status,
		)
	}()

	// Update batch status to "processing" and add start date.
	state.batch.Status = enums.BatchStatusProcessing
	state.batch.StartedAt = temporalsdk_workflow.Now(ctx)
	if err := w.updateBatch(ctx, state); err != nil {
		return err
	}

	// Create SIPs and start processing workflows.
	for i, key := range req.Keys {
		if err := w.startSIPWorkflow(ctx, state, i, key, req.SIPSourceID); err != nil {
			return err
		}
	}

	// TODO: consider errors in child workflows and handle failures, retries, or compensations.

	defer func() {
		// Always wait for all SIP workflows to complete from this point.
		if err := w.waitForWorkflowsCompletion(ctx, state); err != nil {
			e = errors.Join(e, err)
		}
	}()

	// Poll SIP statuses until all are validated.
	activityOpts := withActivityOptsForHeartbeatedPolling(ctx)
	var pollValidatedResult activities.PollSIPStatusesActivityResult
	err := temporalsdk_workflow.ExecuteActivity(
		activityOpts,
		activities.PollSIPStatusesActivityName,
		&activities.PollSIPStatusesActivityParams{
			BatchUUID:        state.batch.UUID,
			ExpectedSIPCount: len(state.sipDetails),
			ExpectedStatus:   enums.SIPStatusValidated,
		},
	).Get(activityOpts, &pollValidatedResult)
	if err != nil {
		return err
	}

	// Send signals to SIP workflows to continue or stop processing.
	if err := w.signalWorkflows(ctx, state, pollValidatedResult.AllExpectedStatus); err != nil {
		return err
	}

	// Fail workflow if not all SIPs reached "validated" status.
	if !pollValidatedResult.AllExpectedStatus {
		return fmt.Errorf("not all SIPs reached %q status", enums.SIPStatusValidated)
	}

	// Poll SIP statuses until all are ingested.
	var pollIngestedResult activities.PollSIPStatusesActivityResult
	err = temporalsdk_workflow.ExecuteActivity(
		activityOpts,
		activities.PollSIPStatusesActivityName,
		&activities.PollSIPStatusesActivityParams{
			BatchUUID:        state.batch.UUID,
			ExpectedSIPCount: len(state.sipDetails),
			ExpectedStatus:   enums.SIPStatusIngested,
		},
	).Get(activityOpts, &pollIngestedResult)
	if err != nil {
		return err
	}

	// Fail workflow if not all SIPs reached "ingested" status.
	if !pollIngestedResult.AllExpectedStatus {
		return fmt.Errorf("not all SIPs reached %q status", enums.SIPStatusIngested)
	}

	// Run post-storage child workflow, if one is configured.
	if w.cfg.Batch.Poststorage != nil {
		if err := w.postStorageWorkflow(ctx, *w.cfg.Batch.Poststorage, state); err != nil {
			return err
		}
	}

	// TODO: handle retention period.

	return nil
}

func (w *BatchWorkflow) updateBatch(ctx temporalsdk_workflow.Context, state *batchWorkflowState) error {
	activityOpts := withLocalActivityOpts(ctx)
	err := temporalsdk_workflow.ExecuteLocalActivity(
		activityOpts,
		updateBatchLocalActivity,
		w.ingestsvc,
		&updateBatchLocalActivityParams{
			UUID:        state.batch.UUID,
			Status:      state.batch.Status,
			StartedAt:   state.batch.StartedAt,
			CompletedAt: state.batch.CompletedAt,
		},
	).Get(activityOpts, nil)
	if err != nil {
		return fmt.Errorf("update batch: %v", err)
	}

	return nil
}

func (w *BatchWorkflow) startSIPWorkflow(
	ctx temporalsdk_workflow.Context,
	state *batchWorkflowState,
	index int,
	key string,
	sourceID uuid.UUID,
) error {
	// Generate SIP UUID using SideEffect to ensure determinism.
	var sipUUID uuid.UUID
	genUUID := temporalsdk_workflow.SideEffect(ctx, func(ctx temporalsdk_workflow.Context) any {
		return uuid.Must(uuid.NewRandomFromReader(w.rng))
	})
	if err := genUUID.Get(&sipUUID); err != nil {
		return fmt.Errorf("generate SIP UUID: %v", err)
	}

	// Create SIP.
	sip := datatypes.SIP{
		UUID:     sipUUID,
		Name:     key,
		Status:   enums.SIPStatusQueued,
		Batch:    &state.batch,
		Uploader: state.batch.Uploader,
	}
	activityOpts := withLocalActivityOpts(ctx)
	err := temporalsdk_workflow.ExecuteLocalActivity(
		activityOpts,
		createSIPLocalActivity,
		w.ingestsvc,
		&createSIPLocalActivityParams{SIP: sip},
	).Get(activityOpts, nil)
	if err != nil {
		return fmt.Errorf("create SIP: %v", err)
	}

	// Start processing workflow for the SIP, keeping track of the workflow future and execution.
	var we temporalsdk_workflow.Execution
	processingCtx := temporalsdk_workflow.WithChildOptions(ctx, temporalsdk_workflow.ChildWorkflowOptions{
		Namespace:         w.cfg.Temporal.Namespace,
		TaskQueue:         w.cfg.Temporal.TaskQueue,
		WorkflowID:        fmt.Sprintf("%s-%s", ingest.ProcessingWorkflowName, sipUUID.String()),
		ParentClosePolicy: temporalapi_enums.PARENT_CLOSE_POLICY_TERMINATE,
	})
	wf := temporalsdk_workflow.ExecuteChildWorkflow(
		processingCtx,
		ingest.ProcessingWorkflowName,
		&ingest.ProcessingWorkflowRequest{
			SIPUUID:         sipUUID,
			SIPName:         key,
			Key:             key,
			SIPSourceID:     sourceID,
			Type:            enums.WorkflowTypeCreateAip,
			RetentionPeriod: -1 * time.Second,
			BatchUUID:       state.batch.UUID,
		},
	)
	err = wf.GetChildWorkflowExecution().Get(processingCtx, &we)
	if err != nil {
		return fmt.Errorf("processing workflow: %v", err)
	}

	// Store SIP details in the batch workflow state.
	state.addSIPDetails(index, sip, wf, we)

	return nil
}

func (w *BatchWorkflow) signalWorkflows(ctx temporalsdk_workflow.Context, state *batchWorkflowState, cont bool) error {
	var signalErr error
	for _, sd := range state.sipDetails {
		err := temporalsdk_workflow.SignalExternalWorkflow(
			ctx,
			sd.workflowExecution.ID,
			sd.workflowExecution.RunID,
			ingest.BatchSignalName,
			ingest.BatchSignal{Continue: cont},
		).Get(ctx, nil)
		if err != nil {
			signalErr = errors.Join(signalErr, err)
		}
	}

	if signalErr != nil {
		return fmt.Errorf("signal workflows: %v", signalErr)
	}

	return nil
}

func (w *BatchWorkflow) waitForWorkflowsCompletion(ctx temporalsdk_workflow.Context, state *batchWorkflowState) error {
	selector := temporalsdk_workflow.NewSelector(ctx)
	for _, sd := range state.sipDetails {
		selector.AddFuture(sd.workflowFuture, func(f temporalsdk_workflow.Future) {
			// Ignore error and result, we just want to know when it's done.
			_ = f.Get(ctx, nil)
		})
	}

	// Wait for all SIP workflows to complete or context cancellation.
	// Block once per SIP/future.
	for range state.sipDetails {
		selector.Select(ctx)
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("waiting for workflows: %v", err)
		}
	}

	return nil
}

func (w *BatchWorkflow) postStorageWorkflow(
	ctx temporalsdk_workflow.Context,
	cfg batch.PostStorageConfig,
	state *batchWorkflowState,
) error {
	state.logger.Info(
		"Starting post-storage workflow",
		"Batch ID", state.batch.UUID.String(),
		"Workflow name", cfg.WorkflowName,
	)

	childCtx := temporalsdk_workflow.WithChildOptions(
		ctx,
		temporalsdk_workflow.ChildWorkflowOptions{
			Namespace:         cfg.Namespace,
			TaskQueue:         cfg.TaskQueue,
			WorkflowID:        fmt.Sprintf("%s-%s", cfg.WorkflowName, state.batch.UUID.String()),
			ParentClosePolicy: temporalapi_enums.PARENT_CLOSE_POLICY_TERMINATE,
		},
	)

	var res batch.PostStorageResult
	err := temporalsdk_workflow.ExecuteChildWorkflow(
		childCtx,
		cfg.WorkflowName,
		&batch.PostStorageParams{SIPs: state.SIPs()},
	).Get(childCtx, &res)
	if err != nil {
		return fmt.Errorf("batch %q post-storage workflow: %v", state.batch.UUID.String(), err)
	}

	state.logger.Info(
		"Post-storage workflow completed",
		"Batch ID", state.batch.UUID.String(),
		"Status", res.Status,
		"Message", res.Message,
	)

	return nil
}
