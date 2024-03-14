// Code generated by ent, DO NOT EDIT.

package db

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db/preservationaction"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db/preservationtask"
	"github.com/google/uuid"
)

// PreservationTask is the model entity for the PreservationTask schema.
type PreservationTask struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// TaskID holds the value of the "task_id" field.
	TaskID uuid.UUID `json:"task_id,omitempty"`
	// Name holds the value of the "name" field.
	Name string `json:"name,omitempty"`
	// Status holds the value of the "status" field.
	Status int8 `json:"status,omitempty"`
	// StartedAt holds the value of the "started_at" field.
	StartedAt time.Time `json:"started_at,omitempty"`
	// CompletedAt holds the value of the "completed_at" field.
	CompletedAt time.Time `json:"completed_at,omitempty"`
	// Note holds the value of the "note" field.
	Note string `json:"note,omitempty"`
	// PreservationActionID holds the value of the "preservation_action_id" field.
	PreservationActionID int `json:"preservation_action_id,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the PreservationTaskQuery when eager-loading is set.
	Edges        PreservationTaskEdges `json:"edges"`
	selectValues sql.SelectValues
}

// PreservationTaskEdges holds the relations/edges for other nodes in the graph.
type PreservationTaskEdges struct {
	// Action holds the value of the action edge.
	Action *PreservationAction `json:"action,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// ActionOrErr returns the Action value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e PreservationTaskEdges) ActionOrErr() (*PreservationAction, error) {
	if e.Action != nil {
		return e.Action, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: preservationaction.Label}
	}
	return nil, &NotLoadedError{edge: "action"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*PreservationTask) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case preservationtask.FieldID, preservationtask.FieldStatus, preservationtask.FieldPreservationActionID:
			values[i] = new(sql.NullInt64)
		case preservationtask.FieldName, preservationtask.FieldNote:
			values[i] = new(sql.NullString)
		case preservationtask.FieldStartedAt, preservationtask.FieldCompletedAt:
			values[i] = new(sql.NullTime)
		case preservationtask.FieldTaskID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the PreservationTask fields.
func (pt *PreservationTask) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case preservationtask.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			pt.ID = int(value.Int64)
		case preservationtask.FieldTaskID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field task_id", values[i])
			} else if value != nil {
				pt.TaskID = *value
			}
		case preservationtask.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				pt.Name = value.String
			}
		case preservationtask.FieldStatus:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				pt.Status = int8(value.Int64)
			}
		case preservationtask.FieldStartedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field started_at", values[i])
			} else if value.Valid {
				pt.StartedAt = value.Time
			}
		case preservationtask.FieldCompletedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field completed_at", values[i])
			} else if value.Valid {
				pt.CompletedAt = value.Time
			}
		case preservationtask.FieldNote:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field note", values[i])
			} else if value.Valid {
				pt.Note = value.String
			}
		case preservationtask.FieldPreservationActionID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field preservation_action_id", values[i])
			} else if value.Valid {
				pt.PreservationActionID = int(value.Int64)
			}
		default:
			pt.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the PreservationTask.
// This includes values selected through modifiers, order, etc.
func (pt *PreservationTask) Value(name string) (ent.Value, error) {
	return pt.selectValues.Get(name)
}

// QueryAction queries the "action" edge of the PreservationTask entity.
func (pt *PreservationTask) QueryAction() *PreservationActionQuery {
	return NewPreservationTaskClient(pt.config).QueryAction(pt)
}

// Update returns a builder for updating this PreservationTask.
// Note that you need to call PreservationTask.Unwrap() before calling this method if this PreservationTask
// was returned from a transaction, and the transaction was committed or rolled back.
func (pt *PreservationTask) Update() *PreservationTaskUpdateOne {
	return NewPreservationTaskClient(pt.config).UpdateOne(pt)
}

// Unwrap unwraps the PreservationTask entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (pt *PreservationTask) Unwrap() *PreservationTask {
	_tx, ok := pt.config.driver.(*txDriver)
	if !ok {
		panic("db: PreservationTask is not a transactional entity")
	}
	pt.config.driver = _tx.drv
	return pt
}

// String implements the fmt.Stringer.
func (pt *PreservationTask) String() string {
	var builder strings.Builder
	builder.WriteString("PreservationTask(")
	builder.WriteString(fmt.Sprintf("id=%v, ", pt.ID))
	builder.WriteString("task_id=")
	builder.WriteString(fmt.Sprintf("%v", pt.TaskID))
	builder.WriteString(", ")
	builder.WriteString("name=")
	builder.WriteString(pt.Name)
	builder.WriteString(", ")
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", pt.Status))
	builder.WriteString(", ")
	builder.WriteString("started_at=")
	builder.WriteString(pt.StartedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("completed_at=")
	builder.WriteString(pt.CompletedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("note=")
	builder.WriteString(pt.Note)
	builder.WriteString(", ")
	builder.WriteString("preservation_action_id=")
	builder.WriteString(fmt.Sprintf("%v", pt.PreservationActionID))
	builder.WriteByte(')')
	return builder.String()
}

// PreservationTasks is a parsable slice of PreservationTask.
type PreservationTasks []*PreservationTask
