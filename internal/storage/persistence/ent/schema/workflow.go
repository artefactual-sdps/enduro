package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/storage/enums"
)

// Workflow holds the schema definition for the Workflow entity.
type Workflow struct {
	ent.Schema
}

// Annotations of the Workflow.
func (Workflow) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "workflow"},
	}
}

// Fields of the Workflow.
func (Workflow) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("uuid", uuid.UUID{}).
			Unique(),
		field.String("temporal_id").
			Annotations(entsql.Annotation{
				Size: 255,
			}),
		field.Enum("type").
			GoType(enums.WorkflowTypeUnspecified),
		field.Enum("status").
			GoType(enums.WorkflowStatusUnspecified),
		field.Time("started_at").
			Optional(),
		field.Time("completed_at").
			Optional(),
		field.Int("aip_id").
			Positive(),
	}
}

// Edges of the Workflow.
func (Workflow) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("aip", AIP.Type).
			Ref("workflows").
			Unique().
			Required().
			Field("aip_id"),
		edge.To("tasks", Task.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("deletion_request", DeletionRequest.Type).
			Annotations(entsql.OnDelete(entsql.SetNull)).
			Unique(),
	}
}

// Indexes of the Workflow.
func (Workflow) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uuid"),
	}
}
