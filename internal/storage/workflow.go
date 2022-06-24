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
	StorageWorkflowName     = "storage-workflow"
	StorageMoveWorkflowName = "storage-move-workflow"
	UploadDoneSignalName    = "upload-done-signal"
)

type StorageWorkflowRequest struct {
	AIPID string
}

type StorageMoveWorkflowRequest struct {
	AIPID    string
	Location string
}

type UploadDoneSignal struct{}

type StorageWorkflow struct {
	logger logr.Logger
}

type StorageMoveWorkflow struct {
	logger logr.Logger
}

func NewStorageWorkflow(logger logr.Logger) *StorageWorkflow {
	return &StorageWorkflow{
		logger: logger,
	}
}

func NewStorageMoveWorkflow(logger logr.Logger) *StorageMoveWorkflow {
	return &StorageMoveWorkflow{
		logger: logger,
	}
}

func (w *StorageWorkflow) Execute(ctx temporalsdk_workflow.Context, req StorageWorkflowRequest) error {
	var signal UploadDoneSignal
	timerFuture := temporalsdk_workflow.NewTimer(ctx, submitURLExpirationTime)
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

func (w *StorageMoveWorkflow) Execute(ctx temporalsdk_workflow.Context, req StorageMoveWorkflowRequest) error {
	var signal string
	timerFuture := temporalsdk_workflow.NewTimer(ctx, 1*time.Minute)
	signalChan := temporalsdk_workflow.GetSignalChannel(ctx, "signal-name")
	selector := temporalsdk_workflow.NewSelector(ctx)
	selector.AddReceive(signalChan, func(channel temporalsdk_workflow.ReceiveChannel, more bool) {
		_ = channel.Receive(ctx, &signal)
	})
	selector.AddFuture(timerFuture, func(f temporalsdk_workflow.Future) {
		_ = f.Get(ctx, nil)
	})
	selector.Select(ctx)
	// XXX: modify minio-setup-buckets-job.yaml to include location buckets?
	// XXX: add activity to copy package.ObjectKey from s.bucket to req.Location
	// XXX: add activity to delete package.Object from s.bucket
	// XXX: add local activity to set storage package location to req.Location
	//      err = s.updatePackageLocation(ctx, req.Location, req.AIPID)
	// XXX: add local activity to set storage package status to Stored
	//      err = s.updatePackageStatus(ctx, StatusStored, req.AIPID)

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

func InitStorageMoveWorkflow(ctx context.Context, tc temporalsdk_client.Client, req *StorageMoveWorkflowRequest) (temporalsdk_client.WorkflowRun, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	opts := temporalsdk_client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("%s-%s", StorageMoveWorkflowName, req.AIPID),
		TaskQueue:             temporal.GlobalTaskQueue,
		WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}
	exec, err := tc.ExecuteWorkflow(ctx, opts, StorageMoveWorkflowName, req)
	if err != nil {
		return nil, err
	}
	return exec, nil
}
