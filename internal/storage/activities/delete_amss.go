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
	ssclient *http.Client
}

type DeleteFromAMSSLocationActivityParams struct {
	Config  types.AMSSConfig
	AIPUUID uuid.UUID
}

type DeleteFromAMSSLocationActivityResult struct {
	Deleted bool
}

func NewDeleteFromAMSSLocationActivity() *DeleteFromAMSSLocationActivity {
	return &DeleteFromAMSSLocationActivity{}
}

func (a *DeleteFromAMSSLocationActivity) Execute(
	ctx context.Context,
	params *DeleteFromAMSSLocationActivityParams,
) (*DeleteFromAMSSLocationActivityResult, error) {
	h := temporal_tools.StartAutoHeartbeat(ctx)
	defer h.Stop()

	a.ssclient = ssblob.NewClient(params.Config.Username, params.Config.APIKey)

	pipelineUUID, err := a.getPipelineUUID(ctx, params)
	if err != nil {
		return nil, err
	}

	if err := a.requestDeletion(ctx, params, pipelineUUID); err != nil {
		return nil, err
	}

	ticker := time.NewTicker(time.Second * 60)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			status, err := a.pollStatus(ctx, params)
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

	resp, err := a.ssclient.Do(req)
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
	params *DeleteFromAMSSLocationActivityParams,
	pipelineUUID string,
) error {
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
		return err
	}

	resp, err := a.ssclient.Do(req)
	if err != nil {
		return fmt.Errorf("request deletion: %v", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("request deletion: response code: %d", resp.StatusCode)
	}

	return nil
}

func (a *DeleteFromAMSSLocationActivity) pollStatus(
	ctx context.Context,
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

	resp, err := a.ssclient.Do(req)
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
			fmt.Errorf("Unexpected AMSS AIP status: %s", status),
		)
	}
}
