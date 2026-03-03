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
		Email:             "user@example.com",
		EmailVerified:     true,
		PreferredUsername: "Preferred username",
		Name:              "Test User",
		Iss:               "issuer",
		Sub:               "subject",
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
				Attributes: []string{auth.IngestSIPSListAttr},
			},
			attributes: []string{auth.IngestSIPSListAttr},
			want:       true,
		},
		{
			name: "Checks multiple attributes exist",
			claims: &auth.Claims{
				Attributes: []string{auth.IngestSIPSListAttr, auth.IngestSIPSReadAttr},
			},
			attributes: []string{auth.IngestSIPSListAttr, auth.IngestSIPSReadAttr},
			want:       true,
		},
		{
			name: "Checks attribute is missing",
			claims: &auth.Claims{
				Attributes: []string{},
			},
			attributes: []string{auth.IngestSIPSDownloadAttr},
			want:       false,
		},
		{
			name:       "Checks attributes on nil claim (auth disabled)",
			attributes: []string{auth.IngestSIPSListAttr},
			want:       true,
		},
		{
			name:       "Checks attributes on nil attributes (ABAC disabled)",
			claims:     &auth.Claims{},
			attributes: []string{auth.IngestSIPSListAttr},
			want:       true,
		},
		{
			name: "Checks attributes with wildcards",
			claims: &auth.Claims{
				Attributes: []string{"ingest:sips:*", "storage:*"},
			},
			attributes: []string{"ingest:sips:list:something", auth.StorageAIPSDownloadAttr},
			want:       true,
		},
		{
			name: "Checks attributes with all wildcard",
			claims: &auth.Claims{
				Attributes: []string{"*"},
			},
			attributes: []string{auth.IngestSIPSListAttr, auth.StorageAIPSDownloadAttr},
			want:       true,
		},
		{
			name: "Checks missing attributes with wildcard",
			claims: &auth.Claims{
				Attributes: []string{"ingest:sips:*"},
			},
			attributes: []string{auth.IngestSIPSListAttr, auth.StorageAIPSDownloadAttr},
			want:       false,
		},
		{
			name: "Checks a more specific attribute doesn't match a general one",
			claims: &auth.Claims{
				Attributes: []string{auth.IngestSIPSListAttr},
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

func TestDisplayName(t *testing.T) {
	t.Parallel()

	type test struct {
		name   string
		claims *auth.Claims
		want   string
	}
	for _, tt := range []test{
		{
			name: "Returns email when all fields are set",
			claims: &auth.Claims{
				Email:             "user@example.com",
				PreferredUsername: "jdoe",
				Name:              "John Doe",
			},
			want: "user@example.com",
		},
		{
			name: "Falls back to preferred username when email is empty",
			claims: &auth.Claims{
				PreferredUsername: "jdoe",
				Name:              "John Doe",
			},
			want: "jdoe",
		},
		{
			name: "Falls back to name when email and preferred username are empty",
			claims: &auth.Claims{
				Name: "John Doe",
			},
			want: "John Doe",
		},
		{
			name:   "Returns empty string when no fields are set",
			claims: &auth.Claims{},
			want:   "",
		},
		{
			name: "Returns empty string for nil claims",
			want: "",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.claims.DisplayName(), tt.want)
		})
	}
}
