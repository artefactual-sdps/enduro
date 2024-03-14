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
		field.String("workflow_id").
			Annotations(entsql.Annotation{
				Size: 255,
			}),
		field.UUID("run_id", uuid.UUID{}).
			Unique(),
		field.UUID("aip_id", uuid.UUID{}).
			Optional(),
		field.UUID("location_id", uuid.UUID{}).
			Optional(),
		field.Int8("status"),
		field.Time("created_at").
			Immutable().
			Default(time.Now),
		field.Time("started_at").
			Optional(),
		field.Time("completed_at").
			Optional(),
	}
}

// Edges of the Pkg.
func (Pkg) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("preservation_actions", PreservationAction.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the Pkg.
func (Pkg) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").
			StorageKey("package_name_idx").
			Annotations(entsql.Prefix(50)),
		index.Fields("aip_id").
			StorageKey("package_aip_id_idx"),
		index.Fields("location_id").
			StorageKey("package_location_id_idx"),
		index.Fields("status").
			StorageKey("package_status_idx"),
		index.Fields("created_at").
			StorageKey("package_created_at_idx"),
		index.Fields("started_at").
			StorageKey("package_started_at_idx"),
	}
}
