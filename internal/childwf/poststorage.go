package childwf

import "encoding/json"

type PostStorageParams struct {
	AIPUUID string

	// CustomMetadata is opaque metadata returned by earlier child workflows.
	CustomMetadata map[string]json.RawMessage
}

type PostStorageResult struct {
	// Outcome indicates the overall result of the child workflow. A value of 0
	// indicates success, while non-zero values indicate various error states.
	Outcome Outcome

	// CustomMetadata is opaque metadata to carry to later child workflows.
	CustomMetadata map[string]json.RawMessage

	// PreservationTasks is a log of the tasks performed by poststorage.
	PreservationTasks []Task
}
