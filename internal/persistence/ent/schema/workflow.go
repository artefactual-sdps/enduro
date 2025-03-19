package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
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
		field.String("temporal_id").
			Annotations(entsql.Annotation{
				Size: 255,
			}),
		field.Int8("type"),
		field.Int8("status"),
		field.Time("started_at").
			Optional(),
		field.Time("completed_at").
			Optional(),
		field.Int("sip_id").
			Positive(),
	}
}

// Edges of the Workflow.
func (Workflow) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("sip", SIP.Type).
			Ref("workflows").
			Unique().
			Required().
			Field("sip_id"),
		edge.To("tasks", Task.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
