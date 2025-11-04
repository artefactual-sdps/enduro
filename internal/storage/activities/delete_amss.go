package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	temporal_tools "go.artefactual.dev/tools/temporal"

	"github.com/artefactual-sdps/enduro/internal/storage/ssblob"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

type DeleteFromAMSSLocationActivity struct {
	approve  bool
	tickerCh <-chan time.Time
}

type DeleteFromAMSSLocationActivityParams struct {
	Config  types.AMSSConfig
	AIPUUID uuid.UUID
}

type DeleteFromAMSSLocationActivityResult struct {
	Deleted bool
}

func NewDeleteFromAMSSLocationActivity(approve bool) *DeleteFromAMSSLocationActivity {
	return &DeleteFromAMSSLocationActivity{approve: approve}
}

// NewDeleteFromAMSSLocationActivityWithTicker creates an activity
// with a custom ticker channel. Intended for tests.
func NewDeleteFromAMSSLocationActivityWithTicker(
	approve bool,
	ch <-chan time.Time,
) *DeleteFromAMSSLocationActivity {
	return &DeleteFromAMSSLocationActivity{
		approve:  approve,
		tickerCh: ch,
	}
}

func (a *DeleteFromAMSSLocationActivity) Execute(
	ctx context.Context,
	params *DeleteFromAMSSLocationActivityParams,
) (*DeleteFromAMSSLocationActivityResult, error) {
	h := temporal_tools.StartAutoHeartbeat(ctx)
	defer h.Stop()

	ssclient := ssblob.NewClient(params.Config.Username, params.Config.APIKey)

	pipelineUUID, err := a.getPipelineUUID(ctx, ssclient, params)
	if err != nil {
		return nil, err
	}

	eventID, err := a.requestDeletion(ctx, ssclient, params, pipelineUUID)
	if err != nil {
		return nil, err
	}

	if a.approve {
		if err := a.approveDeletion(ctx, ssclient, params, eventID); err != nil {
			return nil, err
		}
		return &DeleteFromAMSSLocationActivityResult{Deleted: true}, nil
	}

	// Set up ticker channel if not provided.
	if a.tickerCh == nil {
		ticker := time.NewTicker(time.Second * 60)
		defer ticker.Stop()
		a.tickerCh = ticker.C
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-a.tickerCh:
			status, err := a.pollStatus(ctx, ssclient, params)
			if err != nil {
				return nil, err
			}

			done, err := isProcessed(status)
			if err != nil {
				return nil, err
			}

			if !done {
				continue
			}

			deleted := false
			if status == "DELETED" {
				deleted = true
			}

			return &DeleteFromAMSSLocationActivityResult{Deleted: deleted}, nil
		}
	}
}

func (a *DeleteFromAMSSLocationActivity) getPipelineUUID(
	ctx context.Context,
	ssclient *http.Client,
	params *DeleteFromAMSSLocationActivityParams,
) (string, error) {
	url := fmt.Sprintf(
		"%s/api/v2/file/%s/",
		strings.TrimSuffix(params.Config.URL, "/"),
		params.AIPUUID.String(),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := ssclient.Do(req)
	if err != nil {
		return "", fmt.Errorf("get pipeline UUID: %v", err)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("get pipeline UUID: response code: %d", resp.StatusCode)
	}

	var responseData struct {
		OriginPipeline string `json:"origin_pipeline"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return "", fmt.Errorf("get pipeline UUID: failed to decode response body: %v", err)
	}

	pipelineUUID := strings.TrimSuffix(responseData.OriginPipeline, "/")
	pipelineUUID = strings.TrimPrefix(pipelineUUID, "/api/v2/pipeline/")

	return pipelineUUID, nil
}

func (a *DeleteFromAMSSLocationActivity) requestDeletion(
	ctx context.Context,
	ssclient *http.Client,
	params *DeleteFromAMSSLocationActivityParams,
	pipelineUUID string,
) (eventID int64, err error) {
	url := fmt.Sprintf(
		"%s/api/v2/file/%s/delete_aip/",
		strings.TrimSuffix(params.Config.URL, "/"),
		params.AIPUUID.String(),
	)
	payload := fmt.Sprintf(
		`{"event_reason": "%s", "pipeline": "%s", "user_id": %d, "user_email": "%s"}`,
		"Deletion from Enduro",
		pipelineUUID,
		123,
		"enduro@example.com",
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return 0, err
	}

	resp, err := ssclient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request deletion: %v", err)
	}
	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("request deletion: response code: %d", resp.StatusCode)
	}

	var responseData struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return 0, fmt.Errorf("request deletion: failed to decode response body: %w", err)
	}

	return responseData.ID, nil
}

func (a *DeleteFromAMSSLocationActivity) approveDeletion(
	ctx context.Context,
	ssclient *http.Client,
	params *DeleteFromAMSSLocationActivityParams,
	eventID int64,
) error {
	url := fmt.Sprintf(
		"%s/api/v2/file/%s/review_aip_deletion/",
		strings.TrimSuffix(params.Config.URL, "/"),
		params.AIPUUID.String(),
	)
	payload := fmt.Sprintf(
		`{"event_id": %d, "decision": "%s", "reason": "%s"}`,
		eventID,
		"approve",
		"Approval from Enduro",
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return err
	}

	resp, err := ssclient.Do(req)
	if err != nil {
		return fmt.Errorf("approve deletion: %v", err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("approve deletion: response code: %d", resp.StatusCode)
	}

	var responseData struct {
		ErrorMessage string `json:"error_message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return fmt.Errorf("approve deletion: failed to decode response body: %w", err)
	}

	if responseData.ErrorMessage != "" {
		return fmt.Errorf("approve deletion: %s", responseData.ErrorMessage)
	}

	return nil
}

func (a *DeleteFromAMSSLocationActivity) pollStatus(
	ctx context.Context,
	ssclient *http.Client,
	params *DeleteFromAMSSLocationActivityParams,
) (string, error) {
	url := fmt.Sprintf(
		"%s/api/v2/file/%s/",
		strings.TrimSuffix(params.Config.URL, "/"),
		params.AIPUUID.String(),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := ssclient.Do(req)
	if err != nil {
		return "", fmt.Errorf("poll status: %v", err)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("poll status: response code: %d", resp.StatusCode)
	}

	var responseData struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return "", fmt.Errorf("poll status: failed to decode response body: %w", err)
	}

	return responseData.Status, nil
}

func isProcessed(status string) (bool, error) {
	switch status {
	case "DELETED", "UPLOADED":
		return true, nil
	case "DEL_REQ":
		return false, nil
	default:
		return false, temporal_tools.NewNonRetryableError(
			fmt.Errorf("unexpected AMSS AIP status: %s", status),
		)
	}
}
