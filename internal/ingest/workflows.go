package ingest

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	temporalsdk_api_enums "go.temporal.io/api/enums/v1"
	temporalsdk_client "go.temporal.io/sdk/client"
)

const (
	// Name of the SIP processing workflow.
	ProcessingWorkflowName = "processing-workflow"

	// Name of the signal for reviewing a SIP/AIP.
	ReviewPerformedSignalName = "review-performed-signal"
)

type ReviewPerformedSignal struct {
	Accepted   bool
	LocationID *uuid.UUID
}

type ProcessingWorkflowRequest struct {
	// The zero value represents a new SIP. It can be used to indicate
	// an existing SIP in retries.
	SIPID int

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
	GlobalTaskQueue       string
	PreservationTaskQueue string

	// PollInterval is the time to wait between poll requests to the AM API.
	PollInterval time.Duration

	// TransferDeadline is the maximum time to wait for a transfer to complete.
	// Set to zero for no deadline.
	TransferDeadline time.Duration
}

func InitProcessingWorkflow(ctx context.Context, tc temporalsdk_client.Client, req *ProcessingWorkflowRequest) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	opts := temporalsdk_client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("processing-workflow-%s", uuid.New().String()),
		TaskQueue:             req.GlobalTaskQueue,
		WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}
	_, err := tc.ExecuteWorkflow(ctx, opts, ProcessingWorkflowName, req)

	return err
}
