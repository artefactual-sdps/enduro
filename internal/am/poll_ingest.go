package am

import (
	context "context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"go.artefactual.dev/amclient"
	temporalsdk_activity "go.temporal.io/sdk/activity"
)

const PollIngestActivityName = "poll-ingest-activity"

type PollIngestActivityParams struct {
	SIPID string
}

type PollIngestActivity struct {
	logger logr.Logger
	cfg    *Config
	svc    amclient.IngestService
}

type PollIngestActivityResult struct {
	Status string
}

func NewPollIngestActivity(logger logr.Logger, cfg *Config, svc amclient.IngestService) *PollIngestActivity {
	return &PollIngestActivity{logger: logger, cfg: cfg, svc: svc}
}

// Execute polls Archivematica for the status of an ingest and returns when
// ingest is complete or returns an error status. Execute sends an activity
// heartbeat after each poll.
//
// A response status of "REJECTED", "FAILED", "USER_INPUT", or "BACKLOG" returns
// a temporal.NonRetryableApplicationError to indicate that processing can not
// continue.
func (a *PollIngestActivity) Execute(ctx context.Context, params *PollIngestActivityParams) (*PollIngestActivityResult, error) {
	a.logger.V(1).Info("Executing PollIngestActivity", "SIPID", params.SIPID)

	ticker := time.NewTicker(a.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			reqCtx, cancel := context.WithTimeout(ctx, a.cfg.PollInterval/2)
			resp, httpResp, err := a.poll(reqCtx, params.SIPID, cancel)
			if err == errWorkOngoing {
				// Send a heartbeat then continue polling after the poll interval.
				temporalsdk_activity.RecordHeartbeat(ctx, fmt.Sprintf("Last HTTP response: %v", httpResp.Status))
				continue
			}
			if err != nil {
				return nil, err
			}

			return &PollIngestActivityResult{Status: resp.Status}, nil
		}
	}
}

// poll polls the Archivematica ingest status endpoint, returning a non-nil
// result if ingest processing is complete. An errWorkOngoing error indicates
// work is ongoing and polling should continue. All other errors should
// terminate polling.
func (a *PollIngestActivity) poll(ctx context.Context, transferID string, cancel context.CancelFunc) (*amclient.IngestStatusResponse, *amclient.Response, error) {
	// Cancel the context timer when we return so it doesn't wait for the
	// timeout deadline to expire.
	defer cancel()

	resp, httpResp, err := a.svc.Status(ctx, transferID)
	if err != nil {
		a.logger.V(2).Info("Poll ingest error",
			"StatusCode", httpResp.StatusCode,
			"Status", httpResp.Status,
		)

		amErr := convertAMClientError(httpResp, err)

		// Continue polling on a "400 Bad request" response.
		if amErr == errBadRequest {
			return resp, httpResp, errWorkOngoing
		}

		return resp, httpResp, amErr
	}

	complete, err := isComplete(resp.Status)
	if err != nil {
		return resp, httpResp, err
	}
	if complete {
		return resp, httpResp, nil
	}

	return resp, httpResp, errWorkOngoing
}
