package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/storage/status"
)

// Pkg holds the schema definition for the Pkg entity.
type Pkg struct {
	ent.Schema
}

// Annotations of the Pkg.
func (Pkg) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "package"},
	}
}

// Fields of the Pkg.
func (Pkg) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Annotations(entsql.Annotation{
				Size: 2048,
			}),
		field.UUID("aip_id", uuid.UUID{}),
		field.String("location").
			Annotations(entsql.Annotation{
				Size: 2048,
			}).Optional(),
		field.Enum("status").
			GoType(status.StatusUnspecified),
		field.UUID("object_key", uuid.UUID{}),
	}
}

// Edges of the Pkg.
func (Pkg) Edges() []ent.Edge {
	return nil
}

// Indexes of the Pkg.
func (Pkg) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("aip_id"),
		index.Fields("location").
			Annotations(
				entsql.Prefix(50),
			),
		index.Fields("object_key"),
	}
}