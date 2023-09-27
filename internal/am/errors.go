package am

import (
	"errors"
	"net/http"

	"go.artefactual.dev/amclient"

	"github.com/artefactual-sdps/enduro/internal/temporal"
)

// ConvertAMClientError converts an Archivematica API response to a
// non-retryable temporal ApplicationError if the response is guaranteed not to
// change in subsequent requests.
func convertAMClientError(resp *amclient.Response, err error) error {
	if resp != nil {
		switch resp.Response.StatusCode {
		case http.StatusUnauthorized:
			return temporal.NonRetryableError(errors.New("invalid Archivematica credentials"))
		case http.StatusForbidden:
			return temporal.NonRetryableError(errors.New("insufficient Archivematica permissions"))
		case http.StatusNotFound:
			return temporal.NonRetryableError(errors.New("Archivematica transfer not found"))
		}
	}

	// Retry any client requests that don't return one of the above responses.
	return err
}
