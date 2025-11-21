package workflow

import (
	temporalsdk_log "go.temporal.io/sdk/log"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/ingest"
)

type batchWorkflowState struct {
	logger temporalsdk_log.Logger
	batch  datatypes.Batch
}

func newBatchWorkflowState(ctx temporalsdk_workflow.Context, req *ingest.BatchWorkflowRequest) *batchWorkflowState {
	return &batchWorkflowState{
		logger: temporalsdk_workflow.GetLogger(ctx),
		batch:  req.Batch,
	}
}
