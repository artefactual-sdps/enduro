package am

import (
	context "context"

	"github.com/go-logr/logr"
	"go.artefactual.dev/amclient"
)

const StartTransferActivityName = "start-transfer-activity"

type StartTransferActivity struct {
	logger logr.Logger
	cfg    *Config
	amps   amclient.PackageService
}

type StartTransferActivityParams struct {
	Name string
	Path string
}

type StartTransferActivityResult struct {
	Path string
	UUID string
}

func NewStartTransferActivity(logger logr.Logger, cfg *Config, amps amclient.PackageService) *StartTransferActivity {
	return &StartTransferActivity{
		logger: logger,
		cfg:    cfg,
		amps:   amps,
	}
}

// Execute sends a request to the Archivematica API to start a new
// "auto-approved" transfer. If the request is successful a transfer UUID is
// returned.  An error response will return a retryable or non-retryable
// temporal.ApplicationError, depending on the nature of the error.
func (a *StartTransferActivity) Execute(ctx context.Context, opts *StartTransferActivityParams) (*StartTransferActivityResult, error) {
	payload, resp, err := a.amps.Create(ctx, &amclient.PackageCreateRequest{
		Name:        opts.Name,
		Type:        "standard",
		Path:        opts.Path,
		AutoApprove: true,
	})
	if err != nil {
		return nil, convertAMClientError(resp, err)
	}

	return &StartTransferActivityResult{UUID: payload.ID}, nil
}
