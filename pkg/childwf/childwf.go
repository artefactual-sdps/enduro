// Package childwf defines the wire contracts shared between Enduro and child
// workflow implementations.
package childwf

import "encoding/json"

// Outcome is a status code indicating the overall result of the child
// workflow, with 0 being reserved for success, 1 indicating a critical system
// error, and 2 indicating a non-critical content error. Further outcome codes
// may be added in the future.
type Outcome int

const (
	OutcomeSuccess Outcome = iota
	OutcomeSystemError
	OutcomeContentError
)

// CustomMetadata is opaque JSON metadata carried between child workflows.
type CustomMetadata map[string]json.RawMessage

// User carries non-sensitive user information to child workflows.
type User struct {
	Email string
}
