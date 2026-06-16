package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	temporalsdk_api_enums "go.temporal.io/api/enums/v1"
	temporalsdk_client "go.temporal.io/sdk/client"

	"github.com/artefactual-sdps/enduro/internal/storage/enums"
)

const (
	CopyToPermanentLocationActivityName = "copy-to-permanent-location-activity"
	DeleteFromAMSSLocationActivityName  = "delete-from-amss-location-activity"
	StorageDeleteWorkflowName           = "storage-delete-workflow"
	StorageMoveWorkflowName             = "storage-move-workflow"
	DeletionDecisionSignalName          = "deletion-decision-signal"
)

type StorageDeleteWorkflowRequest struct {
	AIPID       uuid.UUID
	Reason      string
	UserEmail   string
	UserSub     string
	UserIss     string
	TaskQueue   string
	AutoApprove bool
	SkipReport  bool
}

type DeletionDecisionSignal struct {
	Status    enums.DeletionRequestStatus
	UserEmail string
	UserSub   string
	UserIss   string
}

type StorageMoveWorkflowRequest struct {
	AIPID      uuid.UUID
	LocationID uuid.UUID
	TaskQueue  string
}

func StorageDeleteWorkflowID(aipID uuid.UUID) string {
	return fmt.Sprintf("%s-%s", StorageDeleteWorkflowName, aipID)
}

func InitStorageDeleteWorkflow(
	ctx context.Context,
	tc temporalsdk_client.Client,
	req *StorageDeleteWorkflowRequest,
) (temporalsdk_client.WorkflowRun, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	opts := temporalsdk_client.StartWorkflowOptions{
		ID:                    StorageDeleteWorkflowID(req.AIPID),
		TaskQueue:             req.TaskQueue,
		WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
	}
	return tc.ExecuteWorkflow(ctx, opts, StorageDeleteWorkflowName, req)
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
