package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
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
			GoType(enums.LocationSourceUnspecified),
		field.Enum("purpose").
			GoType(enums.LocationPurposeUnspecified),
		field.UUID("uuid", uuid.UUID{}).
			Unique(),
		field.JSON("config", types.LocationConfig{}),
		field.Time("created_at").
			Immutable().
			Default(time.Now),
	}
}

// Edges of the Location.
func (Location) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("aips", AIP.Type).
			Ref("location"),
	}
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
