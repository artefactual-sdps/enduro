package am

import (
	"fmt"

	temporal_tools "go.artefactual.dev/tools/temporal"
)

// isComplete checks the AM transfer status response to determine if processing
// has completed successfully. A non-nil error indicates then AM has stopped
// processing due to a failure or requires user input, and Enduro processing
// should stop. If error is nil then a true result indicates the transfer has
// completed successfully, and a false result means the transfer is still
// processing.
func isComplete(status string) (bool, error) {
	switch status {
	case "COMPLETE":
		return true, nil
	// AM sometimes returns an empty "status" value when processing. :-/
	case "PROCESSING", "":
		return false, nil
	case "REJECTED", "FAILED", "USER_INPUT":
		return false, temporal_tools.NewNonRetryableError(
			fmt.Errorf("Invalid Archivematica response status: %s", status),
		)
	default:
		return false, temporal_tools.NewNonRetryableError(
			fmt.Errorf("Unknown Archivematica response status: %s", status),
		)
	}
}
