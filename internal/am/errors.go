package am

import (
	"errors"
	"fmt"
	"net/http"

	"go.artefactual.dev/amclient"
	temporal_tools "go.artefactual.dev/tools/temporal"
)

// ConvertAMClientError converts an Archivematica API response to a
// non-retryable temporal ApplicationError if the response is guaranteed not to
// change in subsequent requests.
func convertAMClientError(resp *amclient.Response, err error) error {
	if resp == nil || resp.Response == nil {
		return err
	}

	switch {
	case resp.Response.StatusCode == http.StatusUnauthorized:
		return temporal_tools.NewNonRetryableError(errors.New("invalid Archivematica credentials"))
	case resp.Response.StatusCode == http.StatusForbidden:
		return temporal_tools.NewNonRetryableError(errors.New("insufficient Archivematica permissions"))
	case resp.Response.StatusCode == http.StatusNotFound:
		return temporal_tools.NewNonRetryableError(errors.New("Archivematica resource not found"))
	// Archivematica returns a "400 Bad request" code when transfer or ingest
	// status are "in progress", so the request should be retried.
	case resp.Response.StatusCode == http.StatusBadRequest:
		return errors.New("Archivematica response: 400 Bad request")
	// All status codes between 401 and 499 are non-retryable.
	case resp.Response.StatusCode >= 401 && resp.Response.StatusCode < 500:
		return temporal_tools.NewNonRetryableError(
			fmt.Errorf("Archivematica error: %s", resp.Response.Status),
		)
	}

	// Retry any requests that don't return one of the above status codes.
	return fmt.Errorf("Archivematica error: %s", resp.Response.Status)
}
