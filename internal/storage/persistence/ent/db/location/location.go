// Code generated by ent, DO NOT EDIT.

package location

import (
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

const (
	// Label holds the string label denoting the location type in the database.
	Label = "location"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldDescription holds the string denoting the description field in the database.
	FieldDescription = "description"
	// FieldSource holds the string denoting the source field in the database.
	FieldSource = "source"
	// FieldPurpose holds the string denoting the purpose field in the database.
	FieldPurpose = "purpose"
	// FieldUUID holds the string denoting the uuid field in the database.
	FieldUUID = "uuid"
	// FieldConfig holds the string denoting the config field in the database.
	FieldConfig = "config"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// EdgePackages holds the string denoting the packages edge name in mutations.
	EdgePackages = "packages"
	// Table holds the table name of the location in the database.
	Table = "location"
	// PackagesTable is the table that holds the packages relation/edge.
	PackagesTable = "package"
	// PackagesInverseTable is the table name for the Pkg entity.
	// It exists in this package in order to avoid circular dependency with the "pkg" package.
	PackagesInverseTable = "package"
	// PackagesColumn is the table column denoting the packages relation/edge.
	PackagesColumn = "location_id"
)

// Columns holds all SQL columns for location fields.
var Columns = []string{
	FieldID,
	FieldName,
	FieldDescription,
	FieldSource,
	FieldPurpose,
	FieldUUID,
	FieldConfig,
	FieldCreatedAt,
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

var (
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
)

// SourceValidator is a validator for the "source" field enum values. It is called by the builders before save.
func SourceValidator(s types.LocationSource) error {
	switch s.String() {
	case "unspecified", "minio", "sftp", "amss":
		return nil
	default:
		return fmt.Errorf("location: invalid enum value for source field: %q", s)
	}
}

// PurposeValidator is a validator for the "purpose" field enum values. It is called by the builders before save.
func PurposeValidator(pu types.LocationPurpose) error {
	switch pu.String() {
	case "unspecified", "aip_store":
		return nil
	default:
		return fmt.Errorf("location: invalid enum value for purpose field: %q", pu)
	}
}

// OrderOption defines the ordering options for the Location queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByDescription orders the results by the description field.
func ByDescription(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDescription, opts...).ToFunc()
}

// BySource orders the results by the source field.
func BySource(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSource, opts...).ToFunc()
}

// ByPurpose orders the results by the purpose field.
func ByPurpose(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPurpose, opts...).ToFunc()
}

// ByUUID orders the results by the uuid field.
func ByUUID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUUID, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByPackagesCount orders the results by packages count.
func ByPackagesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newPackagesStep(), opts...)
	}
}

// ByPackages orders the results by packages terms.
func ByPackages(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newPackagesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newPackagesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(PackagesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, PackagesTable, PackagesColumn),
	)
}
