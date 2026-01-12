package batch

import (
	"github.com/artefactual-sdps/enduro/internal/datatypes"
)

type PostStorageResultStatus int

// PostStorageResultStatus is a status code indicating the overall result of the
// batch, with zero (0) being reserved for error, and one (1) indicating
// success. Additional status codes may be added in the future.
const (
	PostStorageError PostStorageResultStatus = iota
	PostStorageSuccess
)

type PostStorageParams struct {
	// SIPs is the list of SIPs in the batch. Batch data is embedded in each
	// SIP.
	SIPs []datatypes.SIP
}

type PostStorageResult struct {
	// Status indicates the overall result of the post-storage workflow. An
	// error status (0) indicates that one or more events failed, but the
	// workflow itself completed. If a critical error occurs during workflow
	// execution, the workflow should return a nil result and a standard error
	// instead.
	Status PostStorageResultStatus

	// Message provides additional context about the result of the post-storage
	// workflow execution.
	Message string
}
