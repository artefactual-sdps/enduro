package am

import (
	context "context"
	"errors"
	"fmt"
	"time"

	"github.com/artefactual-sdps/enduro/internal/temporal"
	"github.com/go-logr/logr"
	"go.artefactual.dev/amclient"
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
	dur    time.Duration
}

type PollTransferActivityResult struct {
	SIPID string
	Path  string
}

func NewPollTransferActivity(logger logr.Logger, cfg *Config, amts amclient.TransferService, dur time.Duration) *PollTransferActivity {
	return &PollTransferActivity{logger: logger, cfg: cfg, amts: amts, dur: dur}
}

// Execute sends a transfer status request to the Archivematica API. If a SIP
// has been successfully created from the transfer then the SIP ID and path are
// returned.
//
// A transfer status of "PROCESSING" returns a retryable
// temporal.ApplicationError to indicate that this activity should be retried
// until processing is complete.
//
// A transfer status of "REJECTED", "FAILED", "USER_INPUT", or a transfer that
// has been sent to the Backlog will return a
// temporal.NonRetryableApplicationError to indicate a non-recoverable failure
// state that prevents the workflow from continuing.
func (a *PollTransferActivity) Execute(ctx context.Context, params *PollTransferActivityParams) (*PollTransferActivityResult, error) {
	// Start Heartbeating
	for {
		time.Sleep(a.dur)
		temporalsdk_activity.RecordHeartbeat(ctx, nil)
		resp, httpResp, err := a.amts.Status(ctx, params.TransferID)
		if err != nil {
			return nil, convertAMClientError(httpResp, err)
		}
		if ok, err := transferStatusCheck(resp); !ok {
			if resp.Status == "PROCESSING" {
				continue
			}
			return nil, err
		}
		return &PollTransferActivityResult{SIPID: resp.SIPID, Path: resp.Path}, nil
	}
}

func transferStatusCheck(resp *amclient.TransferStatusResponse) (bool, error) {
	switch resp.Status {
	case "PROCESSING":
		return false, nil
	case "REJECTED", "FAILED", "USER_INPUT":
		return false, temporal.NonRetryableError(fmt.Errorf("Invalid Archivematica transfer status: %s", resp.Status))
	case "COMPLETE":
		if resp.SIPID == "BACKLOG" {
			return false, temporal.NonRetryableError(errors.New("Archivematica transfer sent to backlog"))
		}
		return true, nil
	default:
		return false, temporal.NonRetryableError(fmt.Errorf("Unknown Archivematica transfer status: %s", resp.Status))

	}
}
