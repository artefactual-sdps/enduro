package am

import (
	context "context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"go.artefactual.dev/amclient"
	temporal_tools "go.artefactual.dev/tools/temporal"
	temporalsdk_activity "go.temporal.io/sdk/activity"
)

const PollTransferActivityName = "poll-transfer-activity"

var errWorkOngoing = errors.New("work ongoing")

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
			r, err := a.poll(reqCtx, params.TransferID, cancel)
			if err == errWorkOngoing {
				// Send a heartbeat then continue polling after the poll interval.
				temporalsdk_activity.RecordHeartbeat(ctx, fmt.Sprintf("Last HTTP response: %v", errWorkOngoing))
				continue
			}
			if err != nil {
				return nil, err
			}

			return r, nil
		}
	}
}

// poll polls the Archivematica transfer status endpoint, returning a non-nil
// result if transfer processing is complete. An errWorkOngoing error indicates
// work is ongoing and polling should continue. All other errors should
// terminate polling.
func (a *PollTransferActivity) poll(ctx context.Context, transferID string, cancel context.CancelFunc) (*PollTransferActivityResult, error) {
	// Cancel the context timer when we return so it doesn't wait for the
	// timeout deadline to expire.
	defer cancel()

	resp, httpResp, err := a.amts.Status(ctx, transferID)
	if ferr := transferFailedError(httpResp, err); ferr != nil {
		a.logger.V(2).Info("Poll transfer error",
			"StatusCode", httpResp.StatusCode,
			"Status", httpResp.Status,
		)
		return nil, ferr
	}

	complete, err := isComplete(resp)
	if err != nil {
		return nil, err
	}
	if complete {
		return &PollTransferActivityResult{SIPID: resp.SIPID, Path: resp.Path}, nil
	}

	return nil, errWorkOngoing
}

// transferFailedError checks an amclient error to determine if the transfer has
// failed, or if it is still processing (which returns a 400 status code). If an
// error is returned the activity should return the error, which may or may not
// be a non-retryable error.
func transferFailedError(r *amclient.Response, err error) error {
	if err == nil {
		return nil
	}

	// AM can return a "400 Bad request" HTTP status code while processing, in
	// which case we should continue polling.
	if r != nil && r.Response.StatusCode == http.StatusBadRequest {
		return nil
	}

	return convertAMClientError(r, err)
}

// isComplete checks the AM transfer status response to determine if
// processing has completed successfully. A non-nil error indicates then AM has
// ended processing with a failure or requires user input, and Enduro processing
// should stop. If error is nil then a true result indicates the transfer has
// completed successfully, and a false result means the transfer is still
// processing.
func isComplete(resp *amclient.TransferStatusResponse) (bool, error) {
	if resp == nil {
		return false, nil
	}

	switch resp.Status {
	case "COMPLETE":
		if resp.SIPID == "BACKLOG" {
			return false, temporal_tools.NewNonRetryableError(errors.New("Archivematica SIP sent to backlog"))
		}
		return true, nil
	case "PROCESSING", "":
		return false, nil
	case "REJECTED", "FAILED", "USER_INPUT":
		return false, temporal_tools.NewNonRetryableError(fmt.Errorf("Invalid Archivematica transfer status: %s", resp.Status))
	default:
		return false, temporal_tools.NewNonRetryableError(fmt.Errorf("Unknown Archivematica transfer status: %s", resp.Status))
	}
}
