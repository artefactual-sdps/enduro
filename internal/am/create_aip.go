package am

import (
	context "context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"go.artefactual.dev/amclient"

	"github.com/artefactual-sdps/enduro/internal/temporal"
)

const CreateAIPActivityName = "create-am-aip-activity"

type CreateAIPActivity struct {
	logger logr.Logger
	cfg    *Config
	amis   amclient.IngestService
}

type CreateAIPActivityParams struct {
	UUID string
}

func NewCreateAIPActivity(logger logr.Logger, cfg *Config, amis amclient.IngestService) *CreateAIPActivity {
	return &CreateAIPActivity{
		logger: logger,
		cfg:    cfg,
		amis:   amis,
	}
}

func (a *CreateAIPActivity) Execute(ctx context.Context, opts *CreateAIPActivityParams) error {
	childCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// Start ingest
	payload, resp, err := a.amis.Status(childCtx, opts.UUID)
	if err != nil {
		return temporal.HTTPStatusCodeError(resp.StatusCode, err)
	}
	// Check the ingest status, if it is ok, return no err.
	if ok, err := ingestStatus(payload); !ok {
		return err
	}

	return nil
}

// IngestStatus returns a false bool when the SIP is not fully ingested and an error if it failed.
func ingestStatus(status *amclient.IngestStatusResponse) (bool, error) {
	var ok bool
	if status.Status == "" {
		return ok, fmt.Errorf("error checking ingest status (%w): status is empty", temporal.ContinuePollingError())
	}

	switch status.Status {
	case "COMPLETE":
		ok = true
		return ok, nil
	case "gpPROCESSING":
		return ok, temporal.ContinuePollingError()
	case "USER_INPUT", "FAILED", "REJECTED":
		// TODO: (not in POC) User interactions with workflow.
		return ok, temporal.NonRetryableError(fmt.Errorf("ingest is in a state that we can't handle: %s", status.Status))
	default:
		return ok, fmt.Errorf("error not implemented")
	}
}
