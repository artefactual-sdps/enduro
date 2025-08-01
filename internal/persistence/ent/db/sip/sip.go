// Code generated by ent, DO NOT EDIT.

package sip

import (
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/artefactual-sdps/enduro/internal/enums"
)

const (
	// Label holds the string label denoting the sip type in the database.
	Label = "sip"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldUUID holds the string denoting the uuid field in the database.
	FieldUUID = "uuid"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldAipID holds the string denoting the aip_id field in the database.
	FieldAipID = "aip_id"
	// FieldStatus holds the string denoting the status field in the database.
	FieldStatus = "status"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldStartedAt holds the string denoting the started_at field in the database.
	FieldStartedAt = "started_at"
	// FieldCompletedAt holds the string denoting the completed_at field in the database.
	FieldCompletedAt = "completed_at"
	// FieldFailedAs holds the string denoting the failed_as field in the database.
	FieldFailedAs = "failed_as"
	// FieldFailedKey holds the string denoting the failed_key field in the database.
	FieldFailedKey = "failed_key"
	// FieldUploaderID holds the string denoting the uploader_id field in the database.
	FieldUploaderID = "uploader_id"
	// EdgeWorkflows holds the string denoting the workflows edge name in mutations.
	EdgeWorkflows = "workflows"
	// EdgeUser holds the string denoting the user edge name in mutations.
	EdgeUser = "user"
	// Table holds the table name of the sip in the database.
	Table = "sip"
	// WorkflowsTable is the table that holds the workflows relation/edge.
	WorkflowsTable = "workflow"
	// WorkflowsInverseTable is the table name for the Workflow entity.
	// It exists in this package in order to avoid circular dependency with the "workflow" package.
	WorkflowsInverseTable = "workflow"
	// WorkflowsColumn is the table column denoting the workflows relation/edge.
	WorkflowsColumn = "sip_id"
	// UserTable is the table that holds the user relation/edge.
	UserTable = "sip"
	// UserInverseTable is the table name for the User entity.
	// It exists in this package in order to avoid circular dependency with the "user" package.
	UserInverseTable = "user"
	// UserColumn is the table column denoting the user relation/edge.
	UserColumn = "uploader_id"
)

// Columns holds all SQL columns for sip fields.
var Columns = []string{
	FieldID,
	FieldUUID,
	FieldName,
	FieldAipID,
	FieldStatus,
	FieldCreatedAt,
	FieldStartedAt,
	FieldCompletedAt,
	FieldFailedAs,
	FieldFailedKey,
	FieldUploaderID,
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
	// UploaderIDValidator is a validator for the "uploader_id" field. It is called by the builders before save.
	UploaderIDValidator func(int) error
)

// StatusValidator is a validator for the "status" field enum values. It is called by the builders before save.
func StatusValidator(s enums.SIPStatus) error {
	switch s.String() {
	case "error", "failed", "queued", "processing", "pending", "ingested":
		return nil
	default:
		return fmt.Errorf("sip: invalid enum value for status field: %q", s)
	}
}

// FailedAsValidator is a validator for the "failed_as" field enum values. It is called by the builders before save.
func FailedAsValidator(fa enums.SIPFailedAs) error {
	switch fa.String() {
	case "SIP", "PIP":
		return nil
	default:
		return fmt.Errorf("sip: invalid enum value for failed_as field: %q", fa)
	}
}

// OrderOption defines the ordering options for the SIP queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByUUID orders the results by the uuid field.
func ByUUID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUUID, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByAipID orders the results by the aip_id field.
func ByAipID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAipID, opts...).ToFunc()
}

// ByStatus orders the results by the status field.
func ByStatus(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStatus, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByStartedAt orders the results by the started_at field.
func ByStartedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStartedAt, opts...).ToFunc()
}

// ByCompletedAt orders the results by the completed_at field.
func ByCompletedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCompletedAt, opts...).ToFunc()
}

// ByFailedAs orders the results by the failed_as field.
func ByFailedAs(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldFailedAs, opts...).ToFunc()
}

// ByFailedKey orders the results by the failed_key field.
func ByFailedKey(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldFailedKey, opts...).ToFunc()
}

// ByUploaderID orders the results by the uploader_id field.
func ByUploaderID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUploaderID, opts...).ToFunc()
}

// ByWorkflowsCount orders the results by workflows count.
func ByWorkflowsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newWorkflowsStep(), opts...)
	}
}

// ByWorkflows orders the results by workflows terms.
func ByWorkflows(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newWorkflowsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByUserField orders the results by user field.
func ByUserField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newUserStep(), sql.OrderByField(field, opts...))
	}
}
func newWorkflowsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(WorkflowsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, WorkflowsTable, WorkflowsColumn),
	)
}
func newUserStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(UserInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, UserTable, UserColumn),
	)
}
