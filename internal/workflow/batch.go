package workflow

import (
	"errors"
	"fmt"
	"io"

	temporalsdk_client "go.temporal.io/sdk/client"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
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
		"batch_uuid", req.Batch.UUID.String(),
		"batch_identifier", req.Batch.Identifier,
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
	}()

	// Update batch status to "processing" and add start date.
	state.batch.Status = enums.BatchStatusProcessing
	state.batch.StartedAt = temporalsdk_workflow.Now(ctx)
	if err := w.updateBatch(ctx, state); err != nil {
		return err
	}

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
