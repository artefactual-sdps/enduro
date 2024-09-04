package am

import (
	context "context"
	"path/filepath"

	"go.artefactual.dev/amclient"
	temporal_tools "go.artefactual.dev/tools/temporal"
)

const StartTransferActivityName = "start-transfer-activity"

type StartTransferActivity struct {
	cfg  *Config
	amps amclient.PackageService
}

type StartTransferActivityParams struct {
	// Name of the transfer.
	Name string

	// RelativePath is the PIP path relative to the Archivematica transfer
	// source directory.
	RelativePath string
}

type StartTransferActivityResult struct {
	TransferID string
}

func NewStartTransferActivity(cfg *Config, amps amclient.PackageService) *StartTransferActivity {
	return &StartTransferActivity{
		cfg:  cfg,
		amps: amps,
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
	logger := temporal_tools.GetLogger(ctx)
	logger.V(1).Info(
		"Executing StartTransferActivity",
		"Name", opts.Name,
		"RelativePath", opts.RelativePath,
	)

	processingConfig := a.cfg.ProcessingConfig
	if processingConfig == "" {
		processingConfig = "automated" // Default value.
	}

	payload, resp, err := a.amps.Create(ctx, &amclient.PackageCreateRequest{
		Name:             opts.Name,
		Type:             "zipped bag",
		Path:             filepath.Join(a.cfg.TransferSourcePath, opts.RelativePath),
		ProcessingConfig: processingConfig,
		AutoApprove:      true,
	})
	if err != nil {
		return nil, convertAMClientError(resp, err)
	}

	return &StartTransferActivityResult{TransferID: payload.ID}, nil
}
