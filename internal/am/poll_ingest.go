package am

import (
	context "context"
	"fmt"
	"time"

	"github.com/jonboulle/clockwork"
	"go.artefactual.dev/amclient"
	temporal_tools "go.artefactual.dev/tools/temporal"
	temporalsdk_activity "go.temporal.io/sdk/activity"

	"github.com/artefactual-sdps/enduro/internal/ingest"
)

const PollIngestActivityName = "poll-ingest-activity"

type PollIngestActivityParams struct {
	WorkflowID int
	SIPID      string
}

type PollIngestActivity struct {
	cfg       *Config
	clock     clockwork.Clock
	ingSvc    amclient.IngestService
	jobSvc    amclient.JobsService
	ingestsvc ingest.Service
}

type PollIngestActivityResult struct {
	Status    string
	TaskCount int
}

func NewPollIngestActivity(
	cfg *Config,
	clock clockwork.Clock,
	ingSvc amclient.IngestService,
	jobSvc amclient.JobsService,
	ingestsvc ingest.Service,
) *PollIngestActivity {
	return &PollIngestActivity{
		cfg:       cfg,
		clock:     clock,
		ingSvc:    ingSvc,
		jobSvc:    jobSvc,
		ingestsvc: ingestsvc,
	}
}

// Execute polls Archivematica for the status of an ingest and returns when
// ingest is complete or returns an error status. Execute sends an activity
// heartbeat after each poll.
//
// A response status of "REJECTED", "FAILED", "USER_INPUT", or "BACKLOG" returns
// a temporal.NonRetryableApplicationError to indicate that processing can not
// continue.
func (a *PollIngestActivity) Execute(
	ctx context.Context,
	params *PollIngestActivityParams,
) (*PollIngestActivityResult, error) {
	logger := temporal_tools.GetLogger(ctx)
	logger.V(1).Info("Executing PollIngestActivity",
		"WorkflowID", params.WorkflowID,
		"SIPID", params.SIPID,
	)

	var taskCount int
	jobTracker := NewJobTracker(a.clock, a.jobSvc, a.ingestsvc, params.WorkflowID)
	ticker := time.NewTicker(a.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			resp, count, err := a.poll(ctx, jobTracker, params.SIPID)
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

			return &PollIngestActivityResult{
				Status:    resp.Status,
				TaskCount: taskCount,
			}, nil
		}
	}
}

// poll polls the Archivematica ingest status endpoint, returning a non-nil
// result if ingest processing is complete. An errWorkOngoing error indicates
// work is ongoing and polling should continue. All other errors should
// terminate polling.
func (a *PollIngestActivity) poll(
	ctx context.Context,
	jobTracker *JobTracker,
	transferID string,
) (*amclient.IngestStatusResponse, int, error) {
	var stillWorking bool

	// Add a context timeout to prevent missing the heartbeat deadline.
	ctx, cancel := context.WithTimeout(ctx, a.cfg.PollInterval)
	defer cancel()

	resp, err := a.ingestStatus(ctx, transferID)
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

func (a *PollIngestActivity) ingestStatus(
	ctx context.Context,
	transferID string,
) (*amclient.IngestStatusResponse, error) {
	resp, httpResp, err := a.ingSvc.Status(ctx, transferID)
	if err != nil {
		return resp, convertAMClientError(httpResp, err)
	}

	complete, err := isComplete(resp.Status)
	if err != nil {
		return resp, err
	}
	if complete {
		return resp, nil
	}

	return resp, ErrWorkOngoing
}
