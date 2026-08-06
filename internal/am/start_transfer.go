package am

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"

	"go.artefactual.dev/amclient"
	temporal_tools "go.artefactual.dev/tools/temporal"
	temporalsdk_activity "go.temporal.io/sdk/activity"
)

const StartTransferActivityName = "start-transfer-activity"

type StartTransferActivity struct {
	cfg                *Config
	serverInfoProvider serverInfoProvider
	amps               amclient.PackageService
}

// serverInfoProvider keeps server discovery separate from PackageService.
// amclient exposes it directly on Client rather than through a service.
type serverInfoProvider interface {
	ServerInfo(context.Context) (*amclient.ServerInfo, *amclient.Response, error)
}

type StartTransferActivityParams struct {
	// Name of the transfer.
	Name string

	ZipPIP bool

	// RelativePath is the PIP path relative to the Archivematica transfer
	// source directory.
	RelativePath string
}

type StartTransferActivityResult struct {
	TransferID string
}

func NewStartTransferActivity(
	cfg *Config,
	client *amclient.Client,
) *StartTransferActivity {
	return &StartTransferActivity{
		cfg:                cfg,
		serverInfoProvider: client,
		amps:               client.Package,
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

	transferType := "unzipped bag"
	if opts.ZipPIP {
		transferType = "zipped bag"
	}

	// Inspect Archivematica on every execution so a running worker notices an
	// upgrade. Idempotent transfer submission is supported from 1.19.0; if
	// discovery fails, preserve existing behavior by submitting without a key.
	idempotencyKey := ""
	serverInfo, _, err := a.serverInfoProvider.ServerInfo(ctx)
	if err != nil {
		logger.Info(
			"Archivematica version unavailable; submitting transfer without idempotency protection.",
			"err", err,
		)
	} else if serverInfo.Version.AtLeast(1, 19, 0) {
		idempotencyKey = transferIdempotencyKey(temporalsdk_activity.GetInfo(ctx))
	}

	payload, resp, err := a.amps.Create(ctx, &amclient.PackageCreateRequest{
		Name:             opts.Name,
		Type:             transferType,
		Path:             filepath.Join(a.cfg.TransferSourcePath, opts.RelativePath),
		ProcessingConfig: processingConfig,
		AutoApprove:      new(true),
		IdempotencyKey:   idempotencyKey,
	})
	if err != nil {
		// Interpret 409 and 422 as idempotency responses only when a key was
		// sent. A pending request should be retried; reusing the key with a
		// different payload is permanent and must stop Temporal retries.
		if idempotencyKey != "" && resp != nil && resp.Response != nil {
			switch resp.StatusCode {
			case http.StatusConflict:
				return nil, errors.New("transfer request is still in progress")
			case http.StatusUnprocessableEntity:
				return nil, temporal_tools.NewNonRetryableError(
					errors.New("idempotency key was reused with a different transfer request"),
				)
			}
		}
		return nil, convertAMClientError(resp, err)
	}

	return &StartTransferActivityResult{TransferID: payload.ID}, nil
}

// transferIdempotencyKey derives the key from Temporal's workflow run ID. It
// identifies one logical Archivematica transfer submission and follows the
// retry semantics in the IETF HTTPAPI working-group draft:
// https://datatracker.ietf.org/doc/draft-ietf-httpapi-idempotency-key-header/
//
// The workflow run ID is the right scope because:
//
//   - Automatic retries: Activity and worker-session retries remain in the same
//     run. If Archivematica commits a request but its response never reaches
//     Enduro, the same key replays the transfer instead of creating a duplicate.
//     Activity IDs are too narrow because session retries can change them.
//   - User-triggered retries: Enduro Retry/reprocessing starts a new run and
//     should create a new transfer. SIP and workflow IDs are too broad because
//     they remain stable.
//   - Changed requests: A worker-session retry can change the PIP path.
//     Archivematica then returns HTTP 422 instead of creating a duplicate.
//     Automatic recovery would require Enduro to persist the original request.
//   - Scope limit: This assumes one transfer per workflow run. Supporting more
//     would require a persisted per-submission ID.
func transferIdempotencyKey(info temporalsdk_activity.Info) string {
	runID := info.WorkflowExecution.RunID
	if runID == "" {
		return ""
	}

	return "enduro-transfer-" + runID
}
