package storage

import (
	"context"
	"fmt"
	"io"
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
	logger     logr.Logger
	storagesvc Service
}

func NewStorageWorkflow(logger logr.Logger) *StorageWorkflow {
	return &StorageWorkflow{
		logger: logger,
	}
}

func NewStorageMoveWorkflow(logger logr.Logger, storagesvc Service) *StorageMoveWorkflow {
	return &StorageMoveWorkflow{
		logger:     logger,
		storagesvc: storagesvc,
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

func copyToPermanentLocation(ctx context.Context, storagesvc Service, AIPID, location string) error {
	p, err := storagesvc.ReadPackage(ctx, AIPID)
	if err != nil {
		return err
	}

	reader, err := storagesvc.Bucket().NewReader(ctx, p.ObjectKey, nil)
	if err != nil {
		return err
	}
	defer reader.Close()

	l, err := storagesvc.Location(location)
	if err != nil {
		return err
	}

	bucket, err := l.OpenBucket()
	if err != nil {
		return err
	}
	defer bucket.Close()

	// XXX: what key should we use for the permanent location?
	writer, err := bucket.NewWriter(ctx, p.AIPID, nil)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(writer, reader)
	closeErr := writer.Close()

	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}

	return nil
}

func (w *StorageMoveWorkflow) Execute(ctx temporalsdk_workflow.Context, req StorageMoveWorkflowRequest) error {
	// XXX: how do we get a regular context from the temporal one?
	childContext := context.Background()
	err := copyToPermanentLocation(childContext, w.storagesvc, req.AIPID, req.Location)
	if err != nil {
		return err
	}

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
