package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	temporalsdk_api_enums "go.temporal.io/api/enums/v1"
	temporalsdk_client "go.temporal.io/sdk/client"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-labs/enduro/internal/temporal"
)

const (
	StorageWorkflowName  = "storage-workflow"
	UploadDoneSignalName = "upload-done-signal"
)

type StorageWorkflowRequest struct {
	AIPID string
}

type UploadDoneSignal struct{}

type StorageWorkflow struct {
	logger logr.Logger
}

func NewStorageWorkflow(logger logr.Logger) *StorageWorkflow {
	return &StorageWorkflow{
		logger: logger,
	}
}

func (w *StorageWorkflow) Execute(ctx temporalsdk_workflow.Context, req StorageWorkflowRequest) error {
	var signal UploadDoneSignal
	timerFuture := temporalsdk_workflow.NewTimer(ctx, urlExpirationTime)
	signalChan := temporalsdk_workflow.GetSignalChannel(ctx, UploadDoneSignalName)
	selector := temporalsdk_workflow.NewSelector(ctx)
	selector.AddReceive(signalChan, func(channel temporalsdk_workflow.ReceiveChannel, more bool) {
		_ = channel.Receive(ctx, &signal)
	})
	selector.AddFuture(timerFuture, func(f temporalsdk_workflow.Future) {
		_ = f.Get(ctx, nil)
	})
	selector.Select(ctx)
	return nil
}

func InitStorageWorkflow(ctx context.Context, tc temporalsdk_client.Client, req *StorageWorkflowRequest) (temporalsdk_client.WorkflowRun, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	opts := temporalsdk_client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("%s-%s", StorageWorkflowName, req.AIPID),
		TaskQueue:             temporal.GlobalTaskQueue,
		WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}
	exec, err := tc.ExecuteWorkflow(ctx, opts, StorageWorkflowName, req)
	if err != nil {
		return nil, err
	}
	return exec, nil
}
