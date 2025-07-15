package auth_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
)

func TestUserClaimsFromContext(t *testing.T) {
	t.Parallel()

	t.Run("Returns claims when found", func(t *testing.T) {
		t.Parallel()

		claims := auth.Claims{
			Email:         "info@artefactual.com",
			EmailVerified: true,
			Name:          "Test User",
			Iss:           "http://keycloak:7470/realms/artefactual",
			Sub:           "61a16d59-5029-4d85-8aef-290d1951b8d3",
			Attributes:    []string{"*"},
		}

		ctx := context.Background()
		ctx = auth.WithUserClaims(ctx, &claims)
		assert.Equal(t, auth.UserClaimsFromContext(ctx), &claims)
	})

	t.Run("Returns nil when not found", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		assert.Assert(t, cmp.Nil(auth.UserClaimsFromContext(ctx)))
	})
}

func TestMarshalBinary(t *testing.T) {
	t.Parallel()

	claims := &auth.Claims{
		Email:         "user@example.com",
		EmailVerified: true,
		Name:          "Test User",
		Iss:           "issuer",
		Sub:           "subject",
	}

	data, err := claims.MarshalBinary()
	assert.NilError(t, err)

	var decoded auth.Claims
	err = decoded.UnmarshalBinary(data)
	assert.NilError(t, err)
	assert.DeepEqual(t, decoded, *claims)
}

func TestCheckAttributes(t *testing.T) {
	t.Parallel()

	type test struct {
		name       string
		claims     *auth.Claims
		attributes []string
		want       bool
	}
	for _, tt := range []test{
		{
			name: "Checks without required attributes",
			claims: &auth.Claims{
				Attributes: []string{},
			},
			attributes: []string{},
			want:       true,
		},
		{
			name: "Checks a single attribute exists",
			claims: &auth.Claims{
				Attributes: []string{"ingest:sips:list"},
			},
			attributes: []string{"ingest:sips:list"},
			want:       true,
		},
		{
			name: "Checks multiple attributes exist",
			claims: &auth.Claims{
				Attributes: []string{"ingest:sips:list", "ingest:sips:read"},
			},
			attributes: []string{"ingest:sips:list", "ingest:sips:read"},
			want:       true,
		},
		{
			name: "Checks attribute is missing",
			claims: &auth.Claims{
				Attributes: []string{},
			},
			attributes: []string{"ingest:sips:download"},
			want:       false,
		},
		{
			name:       "Checks attributes on nil claim (auth disabled)",
			attributes: []string{"ingest:sips:list"},
			want:       true,
		},
		{
			name:       "Checks attributes on nil attributes (ABAC disabled)",
			claims:     &auth.Claims{},
			attributes: []string{"ingest:sips:list"},
			want:       true,
		},
		{
			name: "Checks attributes with wildcards",
			claims: &auth.Claims{
				Attributes: []string{"ingest:sips:*", "storage:*"},
			},
			attributes: []string{"ingest:sips:list:something", "storage:aips:download"},
			want:       true,
		},
		{
			name: "Checks attributes with all wildcard",
			claims: &auth.Claims{
				Attributes: []string{"*"},
			},
			attributes: []string{"ingest:sips:list", "storage:aips:download"},
			want:       true,
		},
		{
			name: "Checks missing attributes with wildcard",
			claims: &auth.Claims{
				Attributes: []string{"ingest:sips:*"},
			},
			attributes: []string{"ingest:sips:list", "storage:aips:download"},
			want:       false,
		},
		{
			name: "Checks a more specific attribute doesn't match a general one",
			claims: &auth.Claims{
				Attributes: []string{"ingest:sips:list"},
			},
			attributes: []string{"ingest:sips:list:something"},
			want:       false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.claims.CheckAttributes(tt.attributes), tt.want)
		})
	}
}
