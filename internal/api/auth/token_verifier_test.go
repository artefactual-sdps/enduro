package auth_test

import (
	"context"
	"testing"
	"time"

	"chainguard.dev/go-oidctest/pkg/oidctest"
	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
)

const audience = "test-audience"

func token(t *testing.T, signer jose.Signer, iss string, claims interface{}) (token string) {
	t.Helper()

	// Use signed builder to generate token with given claims.
	builder := jwt.Signed(signer).
		Claims(jwt.Claims{
			Issuer:   iss,
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Expiry:   jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
			Subject:  "test-subject",
			Audience: jwt.Audience{audience},
		})

	// Include extra claims.
	if claims != nil {
		builder = builder.Claims(claims)
	}

	// Serialize token.
	token, err := builder.CompactSerialize()
	assert.NilError(t, err)

	return token
}

func TestOIDCTokenVerifier(t *testing.T) {
	t.Parallel()

	t.Run("Verifies tokens with email verified", func(t *testing.T) {
		t.Parallel()

		signer, iss := oidctest.NewIssuer(t)
		token := token(t, signer, iss, auth.Claims{
			Email:         "info@artefactual.com",
			EmailVerified: true,
		})

		ctx := context.Background()
		v, err := auth.NewOIDCTokenVerifier(ctx, &auth.OIDCConfig{
			ProviderURL: iss,
			ClientID:    audience,
		})
		assert.NilError(t, err)

		claims, err := v.Verify(ctx, token)
		assert.NilError(t, err)
		assert.DeepEqual(t, claims, &auth.Claims{
			Email:         "info@artefactual.com",
			EmailVerified: true,
		})
	})

	t.Run("Verifies tokens without email verified (skipEmailVerifiedCheck)", func(t *testing.T) {
		t.Parallel()

		signer, iss := oidctest.NewIssuer(t)
		token := token(t, signer, iss, auth.Claims{Email: "info@artefactual.com"})

		ctx := context.Background()
		v, err := auth.NewOIDCTokenVerifier(ctx, &auth.OIDCConfig{
			ProviderURL:            iss,
			ClientID:               audience,
			SkipEmailVerifiedCheck: true,
		})
		assert.NilError(t, err)

		claims, err := v.Verify(ctx, token)
		assert.NilError(t, err)
		assert.DeepEqual(t, claims, &auth.Claims{Email: "info@artefactual.com"})
	})

	t.Run("Rejects tokens without email verified", func(t *testing.T) {
		t.Parallel()

		signer, iss := oidctest.NewIssuer(t)
		token := token(t, signer, iss, auth.Claims{Email: "info@artefactual.com"})

		ctx := context.Background()
		v, err := auth.NewOIDCTokenVerifier(ctx, &auth.OIDCConfig{
			ProviderURL: iss,
			ClientID:    audience,
		})
		assert.NilError(t, err)

		claims, err := v.Verify(ctx, token)
		assert.ErrorIs(t, err, auth.ErrUnauthorized)
		assert.Assert(t, cmp.Nil(claims))
	})

	t.Run("Rejects tokens under other errorful conditions", func(t *testing.T) {
		t.Parallel()

		signer, iss := oidctest.NewIssuer(t)
		token := token(t, signer, iss, auth.Claims{
			Email:         "info@artefactual.com",
			EmailVerified: false,
		})

		ctx := context.Background()
		v, err := auth.NewOIDCTokenVerifier(ctx, &auth.OIDCConfig{
			ProviderURL: iss,
			ClientID:    "--- wrong-audience ---",
		})
		assert.NilError(t, err)

		claims, err := v.Verify(ctx, token)
		assert.Error(t, err, "oidc: expected audience \"--- wrong-audience ---\" got [\"test-audience\"]")
		assert.Assert(t, cmp.Nil(claims))
	})

	t.Run("Constructor fails when context is canceled", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := auth.NewOIDCTokenVerifier(ctx, &auth.OIDCConfig{
			ProviderURL: "http://test",
		})
		assert.Error(t, err, "Get \"http://test/.well-known/openid-configuration\": context canceled")
	})
}

func TestParseAttributes(t *testing.T) {
	t.Parallel()

	signer, iss := oidctest.NewIssuer(t)

	type nestedAttr struct {
		NestedAttributes interface{} `json:"nested_attributes,omitempty"`
	}

	type customClaims struct {
		Email         string      `json:"email,omitempty"`
		EmailVerified bool        `json:"email_verified,omitempty"`
		Attributes    interface{} `json:"attributes,omitempty"`
	}

	type test struct {
		name       string
		config     *auth.OIDCConfig
		token      string
		wantClaims *auth.Claims
		wantErr    string
	}
	for _, tt := range []test{
		{
			name: "Parses attributes based on configuration",
			config: &auth.OIDCConfig{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:   true,
					ClaimPath: "attributes",
				},
			},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{"*"},
			}),
			wantClaims: &auth.Claims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{"*"},
			},
		},
		{
			name: "Parses attributes based on configuration (disabled)",
			config: &auth.OIDCConfig{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled: false,
				},
			},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{"*"},
			}),
			wantClaims: &auth.Claims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    nil,
			},
		},
		{
			name: "Parses attributes based on configuration (no attributes)",
			config: &auth.OIDCConfig{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:   true,
					ClaimPath: "attributes",
				},
			},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{},
			}),
			wantClaims: &auth.Claims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{},
			},
		},
		{
			name: "Parses attributes based on configuration (nested)",
			config: &auth.OIDCConfig{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:            true,
					ClaimPath:          "attributes.nested_attributes",
					ClaimPathSeparator: ".",
				},
			},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    nestedAttr{NestedAttributes: []string{"*"}},
			}),
			wantClaims: &auth.Claims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{"*"},
			},
		},
		{
			name: "Parses attributes based on configuration (filtering by prefix)",
			config: &auth.OIDCConfig{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:          true,
					ClaimPath:        "attributes",
					ClaimValuePrefix: "enduro:",
				},
			},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{"enduro:*", "ignore:*"},
			}),
			wantClaims: &auth.Claims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{"*"},
			},
		},
		{
			name: "Parses attributes based on configuration (mapping roles)",
			config: &auth.OIDCConfig{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:   true,
					ClaimPath: "attributes",
					UseRoles:  true,
					RolesMapping: map[string][]string{
						"admin":    {"*"},
						"operator": {"package:list", "package:listActions", "package:move", "package:read", "package:upload"},
						"readonly": {"package:list", "package:listActions", "package:read"},
					},
				},
			},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{"admin", "operator", "readonly"},
			}),
			wantClaims: &auth.Claims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{"*", "package:list", "package:listActions", "package:move", "package:read", "package:upload"},
			},
		},
		{
			name: "Parses attributes based on configuration (mapping roles, no attributes)",
			config: &auth.OIDCConfig{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:      true,
					ClaimPath:    "attributes",
					UseRoles:     true,
					RolesMapping: map[string][]string{"admin": {"*"}},
				},
			},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{"other", "random", "role"},
			}),
			wantClaims: &auth.Claims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{},
			},
		},
		{
			name: "Fails to parse attributes (missing claim)",
			config: &auth.OIDCConfig{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:   true,
					ClaimPath: "attributes",
				},
			},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
			}),
			wantErr: "attributes not found in token, claim path: attributes",
		},
		{
			name: "Fails to parse attributes (missing nested claim)",
			config: &auth.OIDCConfig{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:            true,
					ClaimPath:          "attributes.nested_attributes",
					ClaimPathSeparator: ".",
				},
			},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    nestedAttr{},
			}),
			wantErr: "attributes not found in token, claim path: attributes.nested_attributes",
		},
		{
			name: "Fails to parse attributes (non multivalue claim)",
			config: &auth.OIDCConfig{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:   true,
					ClaimPath: "attributes",
				},
			},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    "*",
			}),
			wantErr: "attributes are not part of a multivalue claim, claim path: attributes",
		},
		{
			name: "Fails to parse attributes (expected nested claim)",
			config: &auth.OIDCConfig{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:            true,
					ClaimPath:          "attributes.nested_attributes",
					ClaimPathSeparator: ".",
				},
			},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    "*",
			}),
			wantErr: "attributes not found in token, claim path: attributes.nested_attributes",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			v, err := auth.NewOIDCTokenVerifier(ctx, tt.config)
			assert.NilError(t, err)

			claims, err := v.Verify(ctx, tt.token)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, claims, tt.wantClaims)
		})
	}
}

func TestNoopTokenVerifier(t *testing.T) {
	t.Run("Verifies tokens", func(t *testing.T) {
		ctx := context.Background()
		v := &auth.NoopTokenVerifier{}

		claims, err := v.Verify(ctx, "")
		assert.NilError(t, err)
		assert.Assert(t, cmp.Nil(claims))
	})
}
