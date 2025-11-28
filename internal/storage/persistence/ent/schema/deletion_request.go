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

// DeletionRequest holds the schema definition for the DeletionRequest entity.
type DeletionRequest struct {
	ent.Schema
}

// Annotations of the DeletionRequest.
func (DeletionRequest) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "deletion_request"},
	}
}

// Fields of the DeletionRequest.
func (DeletionRequest) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("uuid", uuid.UUID{}).
			Unique(),
		field.String("requester").
			Annotations(entsql.Annotation{
				Size: 1024,
			}).
			Immutable(),
		field.String("requester_iss").
			Annotations(entsql.Annotation{
				Size: 1024,
			}).
			Immutable(),
		field.String("requester_sub").
			Annotations(entsql.Annotation{
				Size: 1024,
			}).
			Immutable(),
		field.String("reviewer").
			Annotations(entsql.Annotation{
				Size: 1024,
			}).
			Optional(),
		field.String("reviewer_iss").
			Annotations(entsql.Annotation{
				Size: 1024,
			}).
			Optional(),
		field.String("reviewer_sub").
			Annotations(entsql.Annotation{
				Size: 1024,
			}).
			Optional(),
		field.String("reason").
			Annotations(entsql.Annotation{
				Size: 2048,
			}).
			Immutable(),
		field.Enum("status").
			GoType(enums.DeletionRequestStatusPending).
			Default(enums.DeletionRequestStatusPending.String()),
		field.Time("requested_at").
			Immutable().
			Default(time.Now),
		field.Time("reviewed_at").
			Optional(),
		field.Int("aip_id").
			Positive(),
		field.Int("workflow_id").
			Positive().
			Optional(),
	}
}

// Edges of the DeletionRequest.
func (DeletionRequest) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("aip", AIP.Type).
			Ref("deletion_requests").
			Unique().
			Required().
			Field("aip_id"),
		edge.From("workflow", Workflow.Type).
			Ref("deletion_request").
			Unique().
			Field("workflow_id"),
	}
}

// Indexes of the DeletionRequest.
func (DeletionRequest) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uuid"),
	}
}
