package package_

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	temporalsdk_api_enums "go.temporal.io/api/enums/v1"
	temporalsdk_client "go.temporal.io/sdk/client"

	"github.com/artefactual-sdps/enduro/internal/temporal"
	"github.com/artefactual-sdps/enduro/internal/validation"
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

	// Batch directory that contains the blob.
	BatchDir string

	// Configuration for the validating the transfer.
	ValidationConfig validation.Config
}

func InitProcessingWorkflow(ctx context.Context, tc temporalsdk_client.Client, req *ProcessingWorkflowRequest) error {
	if req.WorkflowID == "" {
		req.WorkflowID = fmt.Sprintf("processing-workflow-%s", uuid.New().String())
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	opts := temporalsdk_client.StartWorkflowOptions{
		ID:                    req.WorkflowID,
		TaskQueue:             temporal.GlobalTaskQueue,
		WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}
	_, err := tc.ExecuteWorkflow(ctx, opts, ProcessingWorkflowName, req)

	return err
}

type MoveWorkflowRequest struct {
	ID         uint
	AIPID      string
	LocationID uuid.UUID
}

func InitMoveWorkflow(ctx context.Context, tc temporalsdk_client.Client, req *MoveWorkflowRequest) (temporalsdk_client.WorkflowRun, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	opts := temporalsdk_client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("%s-%s", MoveWorkflowName, req.AIPID),
		TaskQueue:             temporal.GlobalTaskQueue,
		WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}
	exec, err := tc.ExecuteWorkflow(ctx, opts, MoveWorkflowName, req)

	return exec, err
}
