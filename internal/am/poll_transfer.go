package am

import (
	context "context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"go.artefactual.dev/amclient"
	temporal_tools "go.artefactual.dev/tools/temporal"
	temporalsdk_activity "go.temporal.io/sdk/activity"
)

const PollTransferActivityName = "poll-transfer-activity"

type PollTransferActivityParams struct {
	TransferID string
}

type PollTransferActivity struct {
	logger logr.Logger
	cfg    *Config
	amts   amclient.TransferService
}

type PollTransferActivityResult struct {
	SIPID string
	Path  string
}

func NewPollTransferActivity(logger logr.Logger, cfg *Config, amts amclient.TransferService) *PollTransferActivity {
	return &PollTransferActivity{logger: logger, cfg: cfg, amts: amts}
}

// Execute polls Archivematica for the status of a transfer and returns when
// the transfer is complete or returns an error status. Execute sends an
// activity heartbeat after each poll.
//
// A transfer status of "REJECTED", "FAILED", "USER_INPUT", or "BACKLOG" returns
// a temporal.NonRetryableApplicationError to indicate that processing can not
// continue.
func (a *PollTransferActivity) Execute(ctx context.Context, params *PollTransferActivityParams) (*PollTransferActivityResult, error) {
	a.logger.V(1).Info("Executing PollTransferActivity", "TransferID", params.TransferID)

	ticker := time.NewTicker(a.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			reqCtx, cancel := context.WithTimeout(ctx, a.cfg.PollInterval/2)
			resp, httpResp, err := a.poll(reqCtx, params.TransferID, cancel)
			if err == errWorkOngoing {
				// Send a heartbeat then continue polling after the poll interval.
				temporalsdk_activity.RecordHeartbeat(ctx, fmt.Sprintf("Last HTTP response: %v", httpResp.Status))
				continue
			}
			if err != nil {
				return nil, err
			}

			return &PollTransferActivityResult{SIPID: resp.SIPID, Path: resp.Path}, nil
		}
	}
}

// poll polls the Archivematica transfer status endpoint, returning a non-nil
// result if transfer processing is complete. An errWorkOngoing error indicates
// work is ongoing and polling should continue. All other errors should
// terminate polling.
func (a *PollTransferActivity) poll(ctx context.Context, transferID string, cancel context.CancelFunc) (*amclient.TransferStatusResponse, *amclient.Response, error) {
	// Cancel the context timer when we return so it doesn't wait for the
	// timeout deadline to expire.
	defer cancel()

	resp, httpResp, err := a.amts.Status(ctx, transferID)
	if err != nil {
		a.logger.V(2).Info("Poll transfer error",
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
		if resp.SIPID == "BACKLOG" {
			return resp, httpResp, temporal_tools.NewNonRetryableError(errors.New("Archivematica SIP sent to backlog"))
		}

		return resp, httpResp, nil
	}

	return resp, httpResp, errWorkOngoing
}
