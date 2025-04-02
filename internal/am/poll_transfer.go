package am

import (
	context "context"
	"errors"
	"fmt"
	"time"

	"github.com/jonboulle/clockwork"
	"go.artefactual.dev/amclient"
	temporal_tools "go.artefactual.dev/tools/temporal"
	temporalsdk_activity "go.temporal.io/sdk/activity"

	"github.com/artefactual-sdps/enduro/internal/ingest"
)

const PollTransferActivityName = "poll-transfer-activity"

type PollTransferActivityParams struct {
	WorkflowID int
	TransferID string
}

type PollTransferActivity struct {
	cfg       *Config
	clock     clockwork.Clock
	tfrSvc    amclient.TransferService
	jobSvc    amclient.JobsService
	ingestsvc ingest.Service
}

type PollTransferActivityResult struct {
	SIPID     string
	Path      string
	TaskCount int
}

func NewPollTransferActivity(
	cfg *Config,
	clock clockwork.Clock,
	tfrSvc amclient.TransferService,
	jobSvc amclient.JobsService,
	ingestsvc ingest.Service,
) *PollTransferActivity {
	return &PollTransferActivity{
		cfg:       cfg,
		clock:     clock,
		jobSvc:    jobSvc,
		ingestsvc: ingestsvc,
		tfrSvc:    tfrSvc,
	}
}

// Execute polls Archivematica for the status of a transfer and returns when
// the transfer is complete or returns an error status. Execute sends an
// activity heartbeat after each poll.
//
// On each poll, Execute requests an updated list of AM jobs performed and saves
// the job data to the ingest service as tasks.
//
// A transfer status of "REJECTED", "FAILED", "USER_INPUT", or "BACKLOG" returns
// a temporal.NonRetryableApplicationError to indicate that processing can not
// continue.
func (a *PollTransferActivity) Execute(
	ctx context.Context,
	params *PollTransferActivityParams,
) (*PollTransferActivityResult, error) {
	var taskCount int

	logger := temporal_tools.GetLogger(ctx)
	logger.V(1).Info("Executing PollTransferActivity",
		"WorkflowID", params.WorkflowID,
		"TransferID", params.TransferID,
	)

	jobTracker := NewJobTracker(a.clock, a.jobSvc, a.ingestsvc, params.WorkflowID)
	ticker := time.NewTicker(a.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			resp, count, err := a.poll(ctx, jobTracker, params.TransferID)
			if err == ErrWorkOngoing {
				taskCount += count

				// Send a heartbeat then continue polling after the poll interval.
				temporalsdk_activity.RecordHeartbeat(ctx, fmt.Sprintf("tasks completed: %d", taskCount))
				continue
			}
			if err != nil {
				return nil, err
			}

			taskCount += count

			return &PollTransferActivityResult{SIPID: resp.SIPID, Path: resp.Path, TaskCount: taskCount}, nil
		}
	}
}

// poll queries Archivematica for transfer status and job progress, returning a
// non-nil result if transfer processing is ongoing or complete. An
// ErrWorkOngoing error indicates work is ongoing and polling should continue.
// All other errors should terminate polling.
//
// If the transfer is still in progress or completed successfully, poll saves
// the AM jobs progress as tasks via JobTracker.
func (a *PollTransferActivity) poll(
	ctx context.Context,
	jobTracker *JobTracker,
	transferID string,
) (*amclient.TransferStatusResponse, int, error) {
	var stillWorking bool

	// Add a context timeout to prevent missing the heartbeat deadline.
	ctx, cancel := context.WithTimeout(ctx, a.cfg.PollInterval)
	defer cancel()

	resp, err := a.transferStatus(ctx, transferID)
	if err != nil {
		switch err {
		case ErrBadRequest:
			// Continue polling on a "400 Bad request" response, but don't try
			// and save jobs progress; the jobs endpoint will most likely return
			// a 400 error or an empty jobs list.
			return resp, 0, ErrWorkOngoing
		case ErrWorkOngoing:
			// Save job progress before returning.
			stillWorking = true
		default:
			return nil, 0, err
		}
	}

	// Save job progress as tasks.
	count, err := jobTracker.SaveTasks(ctx, transferID)
	if err == ErrBadRequest {
		// Continue polling on a "400 Bad request" response.
		return resp, 0, ErrWorkOngoing
	}
	if err != nil {
		return nil, 0, fmt.Errorf("save tasks: %v", err)
	}

	// Continue polling.
	if stillWorking {
		return resp, count, ErrWorkOngoing
	}

	return resp, count, nil
}

func (a *PollTransferActivity) transferStatus(
	ctx context.Context,
	transferID string,
) (*amclient.TransferStatusResponse, error) {
	resp, httpResp, err := a.tfrSvc.Status(ctx, transferID)
	if err != nil {
		return resp, convertAMClientError(httpResp, err)
	}

	complete, err := isComplete(resp.Status)
	if err != nil {
		return resp, err
	}
	if complete {
		if resp.SIPID == "BACKLOG" {
			//nolint:staticcheck
			return resp, temporal_tools.NewNonRetryableError(errors.New("Archivematica SIP sent to backlog"))
		}

		return resp, nil
	}

	return resp, ErrWorkOngoing
}
