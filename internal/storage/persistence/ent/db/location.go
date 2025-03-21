// Code generated by ent, DO NOT EDIT.

package db

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/location"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
	"github.com/google/uuid"
)

// Location is the model entity for the Location schema.
type Location struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Name holds the value of the "name" field.
	Name string `json:"name,omitempty"`
	// Description holds the value of the "description" field.
	Description string `json:"description,omitempty"`
	// Source holds the value of the "source" field.
	Source enums.LocationSource `json:"source,omitempty"`
	// Purpose holds the value of the "purpose" field.
	Purpose enums.LocationPurpose `json:"purpose,omitempty"`
	// UUID holds the value of the "uuid" field.
	UUID uuid.UUID `json:"uuid,omitempty"`
	// Config holds the value of the "config" field.
	Config types.LocationConfig `json:"config,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the LocationQuery when eager-loading is set.
	Edges        LocationEdges `json:"edges"`
	selectValues sql.SelectValues
}

// LocationEdges holds the relations/edges for other nodes in the graph.
type LocationEdges struct {
	// Aips holds the value of the aips edge.
	Aips []*AIP `json:"aips,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// AipsOrErr returns the Aips value or an error if the edge
// was not loaded in eager-loading.
func (e LocationEdges) AipsOrErr() ([]*AIP, error) {
	if e.loadedTypes[0] {
		return e.Aips, nil
	}
	return nil, &NotLoadedError{edge: "aips"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Location) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case location.FieldConfig:
			values[i] = new([]byte)
		case location.FieldID:
			values[i] = new(sql.NullInt64)
		case location.FieldName, location.FieldDescription, location.FieldSource, location.FieldPurpose:
			values[i] = new(sql.NullString)
		case location.FieldCreatedAt:
			values[i] = new(sql.NullTime)
		case location.FieldUUID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Location fields.
func (l *Location) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case location.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			l.ID = int(value.Int64)
		case location.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				l.Name = value.String
			}
		case location.FieldDescription:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[i])
			} else if value.Valid {
				l.Description = value.String
			}
		case location.FieldSource:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field source", values[i])
			} else if value.Valid {
				l.Source = enums.LocationSource(value.String)
			}
		case location.FieldPurpose:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field purpose", values[i])
			} else if value.Valid {
				l.Purpose = enums.LocationPurpose(value.String)
			}
		case location.FieldUUID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field uuid", values[i])
			} else if value != nil {
				l.UUID = *value
			}
		case location.FieldConfig:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field config", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &l.Config); err != nil {
					return fmt.Errorf("unmarshal field config: %w", err)
				}
			}
		case location.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				l.CreatedAt = value.Time
			}
		default:
			l.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Location.
// This includes values selected through modifiers, order, etc.
func (l *Location) Value(name string) (ent.Value, error) {
	return l.selectValues.Get(name)
}

// QueryAips queries the "aips" edge of the Location entity.
func (l *Location) QueryAips() *AIPQuery {
	return NewLocationClient(l.config).QueryAips(l)
}

// Update returns a builder for updating this Location.
// Note that you need to call Location.Unwrap() before calling this method if this Location
// was returned from a transaction, and the transaction was committed or rolled back.
func (l *Location) Update() *LocationUpdateOne {
	return NewLocationClient(l.config).UpdateOne(l)
}

// Unwrap unwraps the Location entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (l *Location) Unwrap() *Location {
	_tx, ok := l.config.driver.(*txDriver)
	if !ok {
		panic("db: Location is not a transactional entity")
	}
	l.config.driver = _tx.drv
	return l
}

// String implements the fmt.Stringer.
func (l *Location) String() string {
	var builder strings.Builder
	builder.WriteString("Location(")
	builder.WriteString(fmt.Sprintf("id=%v, ", l.ID))
	builder.WriteString("name=")
	builder.WriteString(l.Name)
	builder.WriteString(", ")
	builder.WriteString("description=")
	builder.WriteString(l.Description)
	builder.WriteString(", ")
	builder.WriteString("source=")
	builder.WriteString(fmt.Sprintf("%v", l.Source))
	builder.WriteString(", ")
	builder.WriteString("purpose=")
	builder.WriteString(fmt.Sprintf("%v", l.Purpose))
	builder.WriteString(", ")
	builder.WriteString("uuid=")
	builder.WriteString(fmt.Sprintf("%v", l.UUID))
	builder.WriteString(", ")
	builder.WriteString("config=")
	builder.WriteString(fmt.Sprintf("%v", l.Config))
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(l.CreatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// Locations is a parsable slice of Location.
type Locations []*Location
