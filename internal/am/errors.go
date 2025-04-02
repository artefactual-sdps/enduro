package am

import (
	"errors"
	"fmt"
	"net/http"

	"go.artefactual.dev/amclient"
	temporal_tools "go.artefactual.dev/tools/temporal"
)

var (
	// ErrWorkOngoing indicates work is ongoing and polling should continue.
	ErrWorkOngoing = errors.New("work ongoing")

	// ErrBadRequest respresents an AM "400 Bad request" response, which can
	// occur while a transfer or ingest is still processing and may require
	// special handling.
	//nolint:staticcheck
	ErrBadRequest = errors.New("Archivematica response: 400 Bad request")
)

// ConvertAMClientError converts an Archivematica API response to a
// non-retryable temporal ApplicationError if the response is guaranteed not to
// change in subsequent requests.
func convertAMClientError(resp *amclient.Response, err error) error {
	if resp == nil || resp.Response == nil {
		return err
	}

	switch {
	case resp.StatusCode == http.StatusBadRequest:
		// Allow retries for "400 Bad request" errors.
		return ErrBadRequest
	case resp.StatusCode == http.StatusUnauthorized:
		return temporal_tools.NewNonRetryableError(errors.New("invalid Archivematica credentials"))
	case resp.StatusCode == http.StatusForbidden:
		return temporal_tools.NewNonRetryableError(errors.New("insufficient Archivematica permissions"))
	case resp.StatusCode == http.StatusNotFound:
		//nolint:staticcheck
		return temporal_tools.NewNonRetryableError(errors.New("Archivematica resource not found"))
	// All status codes between 401 and 499 are non-retryable.
	case resp.StatusCode >= 401 && resp.StatusCode < 500:
		return temporal_tools.NewNonRetryableError(
			//nolint:staticcheck
			fmt.Errorf("Archivematica error: %s", resp.Status),
		)
	}

	// Retry any requests that don't return one of the above status codes.
	//nolint:staticcheck
	return fmt.Errorf("Archivematica error: %s", resp.Status)
}
