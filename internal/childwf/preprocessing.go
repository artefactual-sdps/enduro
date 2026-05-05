package childwf

import (
	"encoding/json"

	"github.com/google/uuid"
)

type PreprocessingParams struct {
	// Relative path to the shared path.
	RelativePath string

	// SIPID is the identifier of the SIP being processed.
	SIPID uuid.UUID

	// BatchID is the identifier of the batch being processed. If the SIP is not
	// part of a batch, this will equal uuid.Nil.
	BatchID uuid.UUID

	// SIPName is the original filename of the SIP being processed.
	SIPName string
}

type PreprocessingResult struct {
	// Outcome is an integer indicating if the workflow completed successfully,
	// or with errors.
	Outcome Outcome

	// CustomMetadata is opaque metadata to carry to later child workflows.
	CustomMetadata map[string]json.RawMessage

	// Relative path to the shared path.
	RelativePath string

	// PreservationTasks is a log of the tasks performed by preprocessing.
	PreservationTasks []Task
}
