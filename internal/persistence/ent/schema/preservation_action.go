package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// PreservationAction holds the schema definition for the PreservationAction entity.
type PreservationAction struct {
	ent.Schema
}

// Annotations of the PreservationAction.
func (PreservationAction) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "preservation_action"},
	}
}

// Fields of the PreservationAction.
func (PreservationAction) Fields() []ent.Field {
	return []ent.Field{
		field.String("workflow_id").
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

// Edges of the PreservationAction.
func (PreservationAction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("sip", SIP.Type).
			Ref("preservation_actions").
			Unique().
			Required().
			Field("sip_id"),
		edge.To("tasks", PreservationTask.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
