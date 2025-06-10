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

// User holds the schema definition for the user entity.
type User struct {
	ent.Schema
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user"},
	}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("uuid", uuid.UUID{}).
			Unique().
			Immutable(),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Annotations(entsql.Default("CURRENT_TIMESTAMP")),
		field.String("email").
			Annotations(entsql.Annotation{
				Size: 1024,
			}).
			Optional(),
		field.String("name").
			Annotations(entsql.Annotation{
				Size: 1024,
			}).
			Optional(),
		field.String("oidc_iss").
			Annotations(entsql.Annotation{
				Size: 1024,
			}).
			Optional(),
		field.String("oidc_sub").
			Annotations(entsql.Annotation{
				Size: 1024,
			}).
			Optional(),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("uploaded_sips", SIP.Type).
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("oidc_iss").
			StorageKey("user_oidc_iss_idx").
			Annotations(entsql.Prefix(50)),
		index.Fields("oidc_sub").
			StorageKey("user_oidc_sub_idx").
			Annotations(entsql.Prefix(50)),
	}
}
