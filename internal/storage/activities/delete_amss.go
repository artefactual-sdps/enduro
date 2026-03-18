package activities

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	ssclient_client "go.artefactual.dev/ssclient"
	ssclient_models "go.artefactual.dev/ssclient/kiota/models"
	temporal_tools "go.artefactual.dev/tools/temporal"

	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

const (
	defaultDeleteFromAMSSPollInterval       = time.Minute
	defaultDeletionUserID             int32 = 123
	defaultDeletionUserEmail                = "enduro@example.com"
)

// errDeletionRequestAlreadyExists reports that AMSS already has a pending
// deletion request for the AIP, so Enduro should not treat it as a new request.
var errDeletionRequestAlreadyExists = errors.New("deletion request already exists")

// DeleteFromAMSSLocationActivity deletes an AIP stored in Archivematica
// Storage Service.
//
// If the activity is configured for automatic approval, it immediately approves
// newly created deletion requests and returns success. Otherwise it polls the
// AIP status until AMSS finishes processing the request.
//
// AMSS may report that a deletion request already exists for the AIP instead of
// creating a new one. In that case the activity can continue with the polling
// path, but it cannot auto-approve the request because AMSS does not return an
// event ID for the existing request.
type DeleteFromAMSSLocationActivity struct {
	// The HTTP client to use for AMSS API calls.
	httpClient *http.Client
	// Whether to automatically approve deletion requests. If false, the
	// activity requests deletion when possible, or reuses an existing pending
	// request, and then polls the AIP status until processing completes.
	approve bool
	// The interval between polling attempts when not automatically approving.
	pollInterval time.Duration
}

type DeleteFromAMSSLocationActivityParams struct {
	Config  types.AMSSConfig
	AIPUUID uuid.UUID
}

type DeleteFromAMSSLocationActivityResult struct {
	Deleted bool
}

func NewDeleteFromAMSSLocationActivity(
	httpClient *http.Client,
	approve bool,
	pollInterval time.Duration,
) *DeleteFromAMSSLocationActivity {
	if pollInterval <= 0 {
		pollInterval = defaultDeleteFromAMSSPollInterval
	}
	return &DeleteFromAMSSLocationActivity{
		httpClient:   httpClient,
		approve:      approve,
		pollInterval: pollInterval,
	}
}

func (a *DeleteFromAMSSLocationActivity) Execute(
	ctx context.Context,
	params *DeleteFromAMSSLocationActivityParams,
) (*DeleteFromAMSSLocationActivityResult, error) {
	h := temporal_tools.StartAutoHeartbeat(ctx)
	defer h.Stop()

	// We build a fresh ssclient per execution. We could cache clients by AMSS
	// connection config, but the shared HTTP client already preserves transport
	// reuse and the extra caching complexity does not seem justified here yet.
	ssclient, err := ssclient_client.New(ssclient_client.Config{
		BaseURL:    params.Config.URL,
		Username:   params.Config.Username,
		Key:        params.Config.APIKey,
		HTTPClient: a.httpClient,
	})
	if err != nil {
		return nil, fmt.Errorf("build Storage Service client: %v", err)
	}

	pipelineUUID, err := a.getPipelineUUID(ctx, ssclient, params.AIPUUID)
	if err != nil {
		return nil, err
	}

	eventID, err := a.requestDeletion(ctx, ssclient, params.AIPUUID, pipelineUUID)
	alreadyExists := errors.Is(err, errDeletionRequestAlreadyExists)
	if err != nil && !alreadyExists {
		return nil, err
	}

	res := &DeleteFromAMSSLocationActivityResult{}

	if a.approve {
		if err := a.approveDeletionRequest(ctx, ssclient, params.AIPUUID, eventID, alreadyExists); err != nil {
			return nil, err
		}
		res.Deleted = true
		return res, nil
	}

	res.Deleted, err = a.waitForDeletion(ctx, ssclient, params.AIPUUID, alreadyExists)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (a *DeleteFromAMSSLocationActivity) approveDeletionRequest(
	ctx context.Context,
	ssclient *ssclient_client.Client,
	aipUUID uuid.UUID,
	eventID int32,
	alreadyExists bool,
) error {
	// AMSS does not return an event ID for an existing request, so Enduro
	// cannot auto-approve it.
	if alreadyExists {
		return fmt.Errorf(
			"approve deletion: deletion request already exists and cannot be approved without an event ID",
		)
	}
	if err := a.approveDeletion(ctx, ssclient, aipUUID, eventID); err != nil {
		return err
	}

	return nil
}

// waitForDeletion polls AMSS for the outcome of a non-auto-approved delete
// request. It returns true when the AIP reaches DELETED and false when the
// request resolves without deletion.
//
// State model:
//   - New request: start with UPLOADED meaning "keep polling", because AMSS may
//     accept the request before it flips the package status to DEL_REQ.
//   - Existing request: start with UPLOADED meaning "not deleted", because the
//     pending request was created before this activity started.
//   - DEL_REQ always means "deletion is pending", so keep polling.
//   - DELETED is the only successful terminal state.
//
// In short:
//   - UPLOADED before deletion is known to be pending => keep polling
//   - UPLOADED after deletion is known to be pending => return Deleted=false
func (a *DeleteFromAMSSLocationActivity) waitForDeletion(
	ctx context.Context,
	ssclient *ssclient_client.Client,
	aipUUID uuid.UUID,
	alreadyExists bool,
) (bool, error) {
	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()

	// For a reused request, an immediate UPLOADED is already a terminal
	// "not deleted" outcome for this activity.
	uploadedMeansNotDeleted := alreadyExists

	for {
		status, err := a.pollStatus(ctx, ssclient, aipUUID)
		if err != nil {
			return false, err
		}

		switch status {
		case "DEL_REQ":
			uploadedMeansNotDeleted = true
		case "DELETED":
			return true, nil
		case "UPLOADED":
			if uploadedMeansNotDeleted {
				return false, nil
			}
		default:
			return false, temporal_tools.NewNonRetryableError(
				fmt.Errorf("unexpected AMSS AIP status: %s", status),
			)
		}

		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-ticker.C:
		}
	}
}

func (a *DeleteFromAMSSLocationActivity) getPipelineUUID(
	ctx context.Context,
	ssclient *ssclient_client.Client,
	aipUUID uuid.UUID,
) (uuid.UUID, error) {
	pkg, err := ssclient.Packages().Get(ctx, aipUUID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get pipeline UUID: %w", err)
	}
	if pkg == nil || pkg.GetOriginPipeline() == nil {
		return uuid.Nil, fmt.Errorf("get pipeline UUID: missing origin pipeline")
	}

	_, pipelineUUID, err := ssclient_client.ParseResourceURI(*pkg.GetOriginPipeline())
	if err != nil {
		return uuid.Nil, fmt.Errorf("get pipeline UUID: %v", err)
	}

	parsed, err := uuid.Parse(pipelineUUID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get pipeline UUID: parse UUID: %v", err)
	}

	return parsed, nil
}

func (a *DeleteFromAMSSLocationActivity) requestDeletion(
	ctx context.Context,
	ssclient *ssclient_client.Client,
	aipUUID uuid.UUID,
	pipelineUUID uuid.UUID,
) (int32, error) {
	req := ssclient_models.NewDeleteAipRequest()
	req.SetEventReason(new("Deletion from Enduro"))
	req.SetPipeline(new(pipelineUUID))
	req.SetUserId(new(defaultDeletionUserID))
	req.SetUserEmail(new(defaultDeletionUserEmail))

	resp, err := ssclient.Packages().DeleteAIP(ctx, aipUUID, req)
	if err != nil {
		return 0, fmt.Errorf("request deletion: %w", err)
	}
	if resp == nil {
		return 0, fmt.Errorf("request deletion: empty response")
	}
	if resp.IsAccepted() {
		return resp.Accepted.ID, nil
	}
	if resp.HasExistingRequest() {
		return 0, errDeletionRequestAlreadyExists
	}
	return 0, fmt.Errorf("request deletion: unexpected delete AIP result")
}

func (a *DeleteFromAMSSLocationActivity) approveDeletion(
	ctx context.Context,
	ssclient *ssclient_client.Client,
	aipUUID uuid.UUID,
	eventID int32,
) error {
	req := ssclient_models.NewReviewAipDeletionRequest()
	req.SetEventId(new(eventID))
	req.SetDecision(new(ssclient_models.APPROVE_REVIEWAIPDELETIONDECISION))
	req.SetReason(new("Approval from Enduro"))

	resp, err := ssclient.Packages().ReviewAIPDeletion(ctx, aipUUID, req)
	if err != nil {
		if reviewErr, ok := errors.AsType[*ssclient_client.ReviewAIPDeletionError](err); ok {
			return fmt.Errorf("approve deletion: %w", reviewErr)
		}
		return fmt.Errorf("approve deletion: %w", err)
	}
	if resp == nil {
		return fmt.Errorf("approve deletion: empty response")
	}

	return nil
}

func (a *DeleteFromAMSSLocationActivity) pollStatus(
	ctx context.Context,
	ssclient *ssclient_client.Client,
	aipUUID uuid.UUID,
) (string, error) {
	pkg, err := ssclient.Packages().Get(ctx, aipUUID)
	if err != nil {
		return "", fmt.Errorf("poll status: %w", err)
	}
	if pkg == nil || pkg.GetStatus() == nil {
		return "", fmt.Errorf("poll status: missing package status")
	}

	return *pkg.GetStatus(), nil
}
