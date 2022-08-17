package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/storage/purpose"
	"github.com/artefactual-sdps/enduro/internal/storage/source"
)

// Location holds the schema definition for the Location entity.
type Location struct {
	ent.Schema
}

// Annotations of the Location.
func (Location) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "location"},
	}
}

// Fields of the Location.
func (Location) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Annotations(entsql.Annotation{
				Size: 2048,
			}),
		field.String("description").
			Annotations(entsql.Annotation{
				Size: 2048,
			}),
		field.Enum("source").
			GoType(source.LocationSourceUnspecified),
		field.Enum("purpose").
			GoType(purpose.LocationPurposeUnspecified),
		field.UUID("uuid", uuid.UUID{}),
	}
}

// Edges of the Location.
func (Location) Edges() []ent.Edge {
	return nil
}

// Indexes of the Location.
func (Location) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").
			Annotations(
				entsql.Prefix(50),
			),
		index.Fields("uuid"),
	}
}
