package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	temporalsdk_api_enums "go.temporal.io/api/enums/v1"
	temporalsdk_client "go.temporal.io/sdk/client"
)

const (
	CopyToPermanentLocationActivityName = "copy-to-permanent-location-activity"
	DeleteFromLocationActivityName      = "delete-from-location-activity"
	StorageUploadWorkflowName           = "storage-upload-workflow"
	StorageMoveWorkflowName             = "storage-move-workflow"
	UploadDoneSignalName                = "upload-done-signal"
)

type StorageUploadWorkflowRequest struct {
	AIPID     uuid.UUID
	TaskQueue string
}

type StorageMoveWorkflowRequest struct {
	AIPID      uuid.UUID
	LocationID uuid.UUID
	TaskQueue  string
}

type CopyToPermanentLocationActivityParams struct {
	AIPID      uuid.UUID
	LocationID uuid.UUID
}

type UploadDoneSignal struct{}

func InitStorageUploadWorkflow(
	ctx context.Context,
	tc temporalsdk_client.Client,
	req *StorageUploadWorkflowRequest,
) (temporalsdk_client.WorkflowRun, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	opts := temporalsdk_client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("%s-%s", StorageUploadWorkflowName, req.AIPID),
		TaskQueue:             req.TaskQueue,
		WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}
	return tc.ExecuteWorkflow(ctx, opts, StorageUploadWorkflowName, req)
}

func InitStorageMoveWorkflow(
	ctx context.Context,
	tc temporalsdk_client.Client,
	req *StorageMoveWorkflowRequest,
) (temporalsdk_client.WorkflowRun, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	opts := temporalsdk_client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("%s-%s", StorageMoveWorkflowName, req.AIPID),
		TaskQueue:             req.TaskQueue,
		WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}
	return tc.ExecuteWorkflow(ctx, opts, StorageMoveWorkflowName, req)
}
