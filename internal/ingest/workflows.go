package ingest

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	temporalsdk_api_enums "go.temporal.io/api/enums/v1"
	temporalsdk_client "go.temporal.io/sdk/client"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
)

const (
	// BatchDecisionSignalName is the name of the signal to continue or cancel a batch.
	BatchDecisionSignalName = "batch-decision-signal"

	// BatchSignalName is the name of the signal to continue processing a SIP.
	BatchSignalName = "batch-signal"

	// BatchWorkflowName is the name of the Batch processing workflow.
	BatchWorkflowName = "batch-workflow"

	// ProcessingWorkflowName is the name of the SIP processing workflow.
	ProcessingWorkflowName = "processing-workflow"

	// ReviewPerformedSignalName is the name of the signal for reviewing a SIP/AIP.
	ReviewPerformedSignalName = "review-performed-signal"
)

type (
	BatchDecisionSignal struct {
		// Continue indicates whether to continue with a partially successful batch.
		Continue bool
	}

	BatchSignal struct {
		// Continue indicates whether to continue processing the SIP.
		Continue bool
	}

	BatchWorkflowRequest struct {
		// Batch contains the Batch details.
		Batch datatypes.Batch

		// SIPSourceID is the ID of the SIP source.
		SIPSourceID uuid.UUID

		// Keys contains the keys of the SIP objects.
		Keys []string

		// RetentionPeriod is the duration for which SIPs should be retained after
		// a successful ingest. If negative, SIPs will be retained indefinitely.
		RetentionPeriod time.Duration
	}

	ProcessingWorkflowRequest struct {
		// SIPUUID is the unique identifier of the SIP.
		SIPUUID uuid.UUID

		// SIPName is the name of the SIP.
		SIPName string

		// Type is the type of workflow to execute.
		Type enums.WorkflowType

		// WatcherName is the name of the watcher that received this blob.
		WatcherName string

		// SIPSourceID is the ID of the SIP source.
		SIPSourceID uuid.UUID

		// RetentionPeriod is the duration for which SIPs should be retained after
		// a successful ingest. If negative, SIPs will be retained indefinitely.
		RetentionPeriod time.Duration

		// CompletedDir is the directory where the transfer is moved to once processing
		// has completed successfully.
		CompletedDir string

		// Key is the key of the blob.
		Key string

		// IsDir indicates whether the blob is a directory (used by the filesystem watcher).
		IsDir bool

		// Extension is the file extension of the original SIP. If it's missing and the SIP
		// is not a directory, the workflow will try to obtain the value after download.
		Extension string

		// BatchUUID is the UUID of the batch this SIP belongs to, if any.
		BatchUUID uuid.UUID
	}

	ReviewPerformedSignal struct {
		Accepted   bool
		LocationID *uuid.UUID
	}
)

func InitBatchWorkflow(
	ctx context.Context,
	tc temporalsdk_client.Client,
	taskQueue string,
	req *BatchWorkflowRequest,
) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	opts := temporalsdk_client.StartWorkflowOptions{
		ID:                    BatchWorkflowID(req.Batch.UUID),
		TaskQueue:             taskQueue,
		WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}
	_, err := tc.ExecuteWorkflow(ctx, opts, BatchWorkflowName, req)
	return err
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
		ID:                    fmt.Sprintf("%s-%s", ProcessingWorkflowName, req.SIPUUID.String()),
		TaskQueue:             taskQueue,
		WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}
	_, err := tc.ExecuteWorkflow(ctx, opts, ProcessingWorkflowName, req)

	return err
}

func BatchWorkflowID(batchID uuid.UUID) string {
	return fmt.Sprintf("%s-%s", BatchWorkflowName, batchID)
}
