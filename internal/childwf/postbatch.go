package childwf

import (
	"github.com/google/uuid"
)

type PostbatchParams struct {
	// Batch represents data general to the whole batch.
	Batch *PostbatchBatch

	// SIPs is the list of SIPs in the batch.
	SIPs []*PostbatchSIP
}

type PostbatchBatch struct {
	UUID      uuid.UUID
	SIPSCount int
}

type PostbatchSIP struct {
	UUID      uuid.UUID
	Name      string
	AIPID     *uuid.UUID // Nullable.
	FileCount int32
}

type PostbatchResult struct {
	// Outcome indicates the overall result of the child workflow. A value of 0
	// indicates success, while non-zero values indicate various error states.
	Outcome Outcome

	// Message provides additional context about the result of the child
	// workflow execution.
	Message string
}
