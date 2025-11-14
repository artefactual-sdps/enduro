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

	"github.com/artefactual-sdps/enduro/internal/enums"
)

// Batch holds the schema definition for the Batch entity.
type Batch struct {
	ent.Schema
}

// Annotations of the Batch.
func (Batch) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "batch"},
	}
}

// Fields of the Batch.
func (Batch) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("uuid", uuid.UUID{}).
			Unique().
			Immutable(),
		field.String("identifier").
			Annotations(entsql.Annotation{
				Size: 2048,
			}),
		field.Enum("status").
			GoType(enums.BatchStatusIngested),
		field.Int("sips_count").
			Positive(),
		field.Time("created_at").
			Immutable().
			Default(time.Now),
		field.Time("started_at").
			Optional(),
		field.Time("completed_at").
			Optional(),
		field.Int("uploader_id").
			Optional().
			Positive(),
	}
}

// Edges of the Batch.
func (Batch) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("sips", SIP.Type).
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.From("uploader", User.Type).
			Field("uploader_id").
			Ref("uploaded_batches").
			Unique(),
	}
}

// Indexes of the Batch.
func (Batch) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("identifier").
			StorageKey("batch_identifier_idx").
			Annotations(entsql.Prefix(50)),
		index.Fields("status").
			StorageKey("batch_status_idx"),
		index.Fields("created_at").
			StorageKey("batch_created_at_idx"),
		index.Fields("started_at").
			StorageKey("batch_started_at_idx"),
		index.Fields("uploader_id").
			StorageKey("batch_uploader_id_idx"),
	}
}
