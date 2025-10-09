package ingest

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	temporalsdk_api_enums "go.temporal.io/api/enums/v1"
	temporalsdk_client "go.temporal.io/sdk/client"

	"github.com/artefactual-sdps/enduro/internal/enums"
)

const (
	// The name of the SIP processing workflow.
	ProcessingWorkflowName = "processing-workflow"

	// The name of the signal for reviewing a SIP/AIP.
	ReviewPerformedSignalName = "review-performed-signal"
)

type ReviewPerformedSignal struct {
	Accepted   bool
	LocationID *uuid.UUID
}

type ProcessingWorkflowRequest struct {
	// The unique identifier of the SIP.
	SIPUUID uuid.UUID

	// The name of the SIP.
	SIPName string

	// The type of workflow to execute.
	Type enums.WorkflowType

	// The name of the watcher that received this blob.
	WatcherName string

	// The ID of the SIP source.
	SIPSourceID uuid.UUID

	// RetentionPeriod is the duration for which SIPs should be retained after
	// a successful ingest. If negative, SIPs will be retained indefinitely.
	RetentionPeriod time.Duration

	// The directory where the transfer is moved to once processing has completed
	// successfully.
	CompletedDir string

	// The key of the blob.
	Key string

	// Indicates whether the blob is a directory (used by the filesystem watcher).
	IsDir bool

	// The file extension of the original SIP. If it's missing and the SIP is not a
	// directory, the workflow will try to obtain the value after download.
	Extension string
}

func InitProcessingWorkflow(
	ctx context.Context,
	tc temporalsdk_client.Client,
	taskQueue string,
	req *ProcessingWorkflowRequest,
) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	opts := temporalsdk_client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("processing-workflow-%s", req.SIPUUID.String()),
		TaskQueue:             taskQueue,
		WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}
	_, err := tc.ExecuteWorkflow(ctx, opts, ProcessingWorkflowName, req)

	return err
}
