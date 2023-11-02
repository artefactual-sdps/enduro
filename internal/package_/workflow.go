package package_

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	temporalsdk_api_enums "go.temporal.io/api/enums/v1"
	temporalsdk_client "go.temporal.io/sdk/client"
)

const (
	// Name of the package move workflow.
	MoveWorkflowName = "move-workflow"

	// Name of the package processing workflow.
	ProcessingWorkflowName = "processing-workflow"

	// Name of the signal for reviewing a package.
	ReviewPerformedSignalName = "review-performed-signal"
)

type ReviewPerformedSignal struct {
	Accepted   bool
	LocationID *uuid.UUID
}

type ProcessingWorkflowRequest struct {
	WorkflowID string `json:"-"`

	// The zero value represents a new package. It can be used to indicate
	// an existing package in retries.
	PackageID uint

	// Name of the watcher that received this blob.
	WatcherName string

	// Period of time to schedule the deletion of the original blob from the
	// watched data source. nil means no deletion.
	RetentionPeriod *time.Duration

	// Directory where the transfer is moved to once processing has completed
	// successfully.
	CompletedDir string

	// Whether the top-level directory is meant to be stripped.
	StripTopLevelDir bool

	// Key of the blob.
	Key string

	// Whether the blob is a directory (fs watcher)
	IsDir bool

	// Whether the AIP is stored automatically in the default permanent location.
	AutoApproveAIP bool

	// Location identifier for storing auto approved AIPs.
	DefaultPermanentLocationID *uuid.UUID

	// Task queues used for starting new workflows.
	TaskQueue    string
	A3mTaskQueue string
}

func InitProcessingWorkflow(ctx context.Context, tc temporalsdk_client.Client, req *ProcessingWorkflowRequest) error {
	if req.WorkflowID == "" {
		req.WorkflowID = fmt.Sprintf("processing-workflow-%s", uuid.New().String())
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	opts := temporalsdk_client.StartWorkflowOptions{
		ID:                    req.WorkflowID,
		TaskQueue:             req.TaskQueue,
		WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}
	_, err := tc.ExecuteWorkflow(ctx, opts, ProcessingWorkflowName, req)

	return err
}

type MoveWorkflowRequest struct {
	ID         uint
	AIPID      string
	LocationID uuid.UUID
	TaskQueue  string
}

func InitMoveWorkflow(ctx context.Context, tc temporalsdk_client.Client, req *MoveWorkflowRequest) (temporalsdk_client.WorkflowRun, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	opts := temporalsdk_client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("%s-%s", MoveWorkflowName, req.AIPID),
		TaskQueue:             req.TaskQueue,
		WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}
	exec, err := tc.ExecuteWorkflow(ctx, opts, MoveWorkflowName, req)

	return exec, err
}
