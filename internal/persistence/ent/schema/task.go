package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Task holds the schema definition for the Task entity.
type Task struct {
	ent.Schema
}

// Annotations of the Task.
func (Task) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "task"},
	}
}

// Fields of the Task.
func (Task) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("uuid", uuid.UUID{}).
			Unique().
			Immutable(),
		field.String("name").
			Annotations(entsql.Annotation{
				Size: 2048,
			}),
		field.Int8("status"),
		field.Time("started_at").
			Optional(),
		field.Time("completed_at").
			Optional(),
		field.Text("note"),
		field.Int("workflow_id").
			Positive(),
	}
}

// Edges of the Task.
func (Task) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow", Workflow.Type).
			Ref("tasks").
			Unique().
			Required().
			Field("workflow_id"),
	}
}
