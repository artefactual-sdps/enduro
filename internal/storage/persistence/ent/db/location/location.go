// Code generated by ent, DO NOT EDIT.

package location

import (
	"fmt"

	"github.com/artefactual-sdps/enduro/internal/storage/purpose"
	"github.com/artefactual-sdps/enduro/internal/storage/source"
)

const (
	// Label holds the string label denoting the location type in the database.
	Label = "location"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldSource holds the string denoting the source field in the database.
	FieldSource = "source"
	// FieldPurpose holds the string denoting the purpose field in the database.
	FieldPurpose = "purpose"
	// FieldUUID holds the string denoting the uuid field in the database.
	FieldUUID = "uuid"
	// Table holds the table name of the location in the database.
	Table = "location"
)

// Columns holds all SQL columns for location fields.
var Columns = []string{
	FieldID,
	FieldName,
	FieldSource,
	FieldPurpose,
	FieldUUID,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

// SourceValidator is a validator for the "source" field enum values. It is called by the builders before save.
func SourceValidator(s source.LocationSource) error {
	switch s.String() {
	case "unspecified", "minio":
		return nil
	default:
		return fmt.Errorf("location: invalid enum value for source field: %q", s)
	}
}

// PurposeValidator is a validator for the "purpose" field enum values. It is called by the builders before save.
func PurposeValidator(pu purpose.LocationPurpose) error {
	switch pu.String() {
	case "unspecified", "aip_store":
		return nil
	default:
		return fmt.Errorf("location: invalid enum value for purpose field: %q", pu)
	}
}
