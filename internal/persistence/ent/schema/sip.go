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

// SIP holds the schema definition for the SIP entity.
type SIP struct {
	ent.Schema
}

// Annotations of the SIP.
func (SIP) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sip"},
	}
}

// Fields of the SIP.
func (SIP) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("uuid", uuid.UUID{}).
			Unique().
			Immutable(),
		field.String("name").
			Annotations(entsql.Annotation{
				Size: 2048,
			}),
		field.UUID("aip_id", uuid.UUID{}).
			Optional(),
		field.Enum("status").
			GoType(enums.SIPStatusIngested),
		field.Time("created_at").
			Immutable().
			Default(time.Now),
		field.Time("started_at").
			Optional(),
		field.Time("completed_at").
			Optional(),
		field.Enum("failed_as").
			GoType(enums.SIPFailedAsSIP).
			Optional(),
		field.String("failed_key").
			Annotations(entsql.Annotation{
				Size: 1024,
			}).
			Optional(),
	}
}

// Edges of the SIP.
func (SIP) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("workflows", Workflow.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the SIP.
func (SIP) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").
			StorageKey("sip_name_idx").
			Annotations(entsql.Prefix(50)),
		index.Fields("aip_id").
			StorageKey("sip_aip_id_idx"),
		index.Fields("status").
			StorageKey("sip_status_idx"),
		index.Fields("created_at").
			StorageKey("sip_created_at_idx"),
		index.Fields("started_at").
			StorageKey("sip_started_at_idx"),
	}
}
