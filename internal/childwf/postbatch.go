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
	UUID  uuid.UUID
	Name  string
	AIPID *uuid.UUID // Nullable.
}
