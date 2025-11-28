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
)

// AIP holds the schema definition for the AIP entity.
type AIP struct {
	ent.Schema
}

// Annotations of the AIP.
func (AIP) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "aip"},
	}
}

// Fields of the AIP.
func (AIP) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Annotations(entsql.Annotation{
				Size: 2048,
			}),
		field.UUID("aip_id", uuid.UUID{}).
			Unique(),
		field.Int("location_id").
			Optional(),
		field.Enum("status").
			GoType(enums.AIPStatusUnspecified),
		field.UUID("object_key", uuid.UUID{}).
			Unique(),
		field.Time("created_at").
			Immutable().
			Default(time.Now),
		field.String("deletion_report_key").
			Annotations(entsql.Annotation{Size: 1024}).
			Optional(),
	}
}

// Edges of the AIP.
func (AIP) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("location", Location.Type).
			Field("location_id").
			Unique(),
		edge.To("workflows", Workflow.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("deletion_requests", DeletionRequest.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the AIP.
func (AIP) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("aip_id"),
		index.Fields("object_key"),
	}
}
