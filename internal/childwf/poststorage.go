package childwf

import "encoding/json"

type PostStorageParams struct {
	AIPUUID string

	// CustomMetadata is opaque metadata returned by earlier child workflows.
	CustomMetadata map[string]json.RawMessage
}
