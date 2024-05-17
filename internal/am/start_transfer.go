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
	TransferID string
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
func (a *StartTransferActivity) Execute(
	ctx context.Context,
	opts *StartTransferActivityParams,
) (*StartTransferActivityResult, error) {
	a.logger.V(1).Info("Executing StartTransferActivity", "Name", opts.Name, "Path", opts.Path)

	processingConfig := a.cfg.ProcessingConfig
	if processingConfig == "" {
		processingConfig = "automated" // Default value.
	}

	payload, resp, err := a.amps.Create(ctx, &amclient.PackageCreateRequest{
		Name:             opts.Name,
		Type:             "zipped bag",
		Path:             opts.Path,
		ProcessingConfig: processingConfig,
		AutoApprove:      true,
	})
	if err != nil {
		return nil, convertAMClientError(resp, err)
	}

	return &StartTransferActivityResult{TransferID: payload.ID}, nil
}
