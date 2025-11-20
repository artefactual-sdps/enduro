package workflow

import (
	"io"

	temporalsdk_client "go.temporal.io/sdk/client"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-sdps/enduro/internal/config"
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
	logger := temporalsdk_workflow.GetLogger(ctx)
	logger.Info(
		"Starting batch workflow",
		"batch_uuid", req.Batch.UUID.String(),
		"batch_identifier", req.Batch.Identifier,
		"source_id", req.SIPSourceID.String(),
		"keys", req.Keys,
	)
	return nil
}
