package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// PreservationTask holds the schema definition for the PreservationTask entity.
type PreservationTask struct {
	ent.Schema
}

// Annotations of the PreservationTask.
func (PreservationTask) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "preservation_task",
			Collation: "utf8mb4_0900_ai_ci",
		},
	}
}

// Fields of the PreservationTask.
func (PreservationTask) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			SchemaType(map[string]string{
				dialect.MySQL: "INT UNSIGNED",
			}).
			Immutable(),
		field.UUID("task_id", uuid.New()).
			SchemaType(map[string]string{
				dialect.MySQL: "VARCHAR(36)",
			}),
		field.String("name").
			Annotations(entsql.Annotation{
				Size: 2048,
			}),
		field.Int8("status"),
		field.Time("started_at").
			SchemaType(map[string]string{
				dialect.MySQL: "TIMESTAMP(6)",
			}).
			Optional(),
		field.Time("completed_at").
			SchemaType(map[string]string{
				dialect.MySQL: "TIMESTAMP(6)",
			}).
			Optional(),
		field.Text("note"),
		field.Int("preservation_action_id").
			Positive(),
	}
}

// Edges of the PreservationTask.
func (PreservationTask) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("action", PreservationAction.Type).
			Ref("tasks").
			Unique().
			Required().
			Field("preservation_action_id"),
	}
}
