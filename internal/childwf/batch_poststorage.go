package childwf

import (
	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
)

const BatchPostStorageName = "batch-poststorage"

type BPSParams struct {
	// Batch represents data general to the whole batch.
	Batch *BPSBatch

	// SIPs is the list of SIPs in the batch.
	SIPs []*BPSSIP
}

type BPSBatch struct {
	UUID      uuid.UUID
	SIPSCount int
}

type BPSSIP struct {
	UUID  uuid.UUID
	Name  string
	AIPID *uuid.UUID // Nullable.
}

func SIPstoBPSSIPs(sips []datatypes.SIP) []*BPSSIP {
	r := make([]*BPSSIP, len(sips))
	for i, sip := range sips {
		s := &BPSSIP{
			UUID: sip.UUID,
			Name: sip.Name,
		}
		if sip.AIPID.Valid {
			s.AIPID = &sip.AIPID.UUID
		}
		r[i] = s
	}

	return r
}

func BatchtoBPSBatch(batch datatypes.Batch) *BPSBatch {
	return &BPSBatch{
		UUID:      batch.UUID,
		SIPSCount: batch.SIPSCount,
	}
}
