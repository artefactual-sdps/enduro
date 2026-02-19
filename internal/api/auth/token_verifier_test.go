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

const (
	audience = "test-audience"
	subject  = "test-subject"
)

func token(t *testing.T, signer jose.Signer, iss string, claims any) (token string) {
	t.Helper()

	// Use signed builder to generate token with given claims.
	builder := jwt.Signed(signer).
		Claims(jwt.Claims{
			Issuer:   iss,
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Expiry:   jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
			Subject:  subject,
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

	type verification struct {
		token           string
		wantClaims      *auth.Claims
		wantErrContains []string
		wantErrIs       error
	}
	type test struct {
		name  string
		setup func(t *testing.T) (auth.OIDCConfigs, []verification)
	}
	for _, tt := range []test{
		{
			name: "Verifies tokens with email verified",
			setup: func(t *testing.T) (auth.OIDCConfigs, []verification) {
				signer, iss := oidctest.NewIssuer(t)
				token := token(t, signer, iss, auth.Claims{
					Email:         "info@artefactual.com",
					EmailVerified: true,
				})

				return auth.OIDCConfigs{
						{ProviderURL: iss, ClientID: audience},
					}, []verification{
						{
							token: token,
							wantClaims: &auth.Claims{
								Email:         "info@artefactual.com",
								EmailVerified: true,
								Iss:           iss,
								Sub:           subject,
							},
						},
					}
			},
		},
		{
			name: "Verifies tokens without email verified (skipEmailVerifiedCheck)",
			setup: func(t *testing.T) (auth.OIDCConfigs, []verification) {
				signer, iss := oidctest.NewIssuer(t)
				token := token(t, signer, iss, auth.Claims{
					Email: "info@artefactual.com",
				})

				return auth.OIDCConfigs{
						{
							ProviderURL:            iss,
							ClientID:               audience,
							SkipEmailVerifiedCheck: true,
						},
					}, []verification{
						{
							token: token,
							wantClaims: &auth.Claims{
								Email: "info@artefactual.com",
								Iss:   iss,
								Sub:   subject,
							},
						},
					}
			},
		},
		{
			name: "Rejects tokens without email verified",
			setup: func(t *testing.T) (auth.OIDCConfigs, []verification) {
				signer, iss := oidctest.NewIssuer(t)
				token := token(t, signer, iss, auth.Claims{
					Email: "info@artefactual.com",
				})

				return auth.OIDCConfigs{
						{ProviderURL: iss, ClientID: audience},
					}, []verification{
						{
							token:     token,
							wantErrIs: auth.ErrUnauthorized,
						},
					}
			},
		},
		{
			name: "Rejects tokens under other errorful conditions",
			setup: func(t *testing.T) (auth.OIDCConfigs, []verification) {
				signer, iss := oidctest.NewIssuer(t)
				token := token(t, signer, iss, auth.Claims{
					Email:         "info@artefactual.com",
					EmailVerified: false,
				})

				return auth.OIDCConfigs{
						{ProviderURL: iss, ClientID: "--- wrong-audience ---"},
					}, []verification{
						{
							token: token,
							wantErrContains: []string{
								`oidc: expected audience "--- wrong-audience ---" got ["test-audience"]`,
							},
						},
					}
			},
		},
		{
			name: "Verifies token when multiple providers are configured",
			setup: func(t *testing.T) (auth.OIDCConfigs, []verification) {
				signerA, issA := oidctest.NewIssuer(t)
				tokenA := token(t, signerA, issA, auth.Claims{
					Email:         "example@artefactual.com",
					EmailVerified: true,
				})
				signerB, issB := oidctest.NewIssuer(t)
				tokenB := token(t, signerB, issB, auth.Claims{
					Email:         "example@artefactual.com",
					EmailVerified: true,
				})

				return auth.OIDCConfigs{
						{ProviderURL: issA, ClientID: audience},
						{ProviderURL: issB, ClientID: audience},
					}, []verification{
						{
							token: tokenA,
							wantClaims: &auth.Claims{
								Email:         "example@artefactual.com",
								EmailVerified: true,
								Iss:           issA,
								Sub:           subject,
							},
						},
						{
							token: tokenB,
							wantClaims: &auth.Claims{
								Email:         "example@artefactual.com",
								EmailVerified: true,
								Iss:           issB,
								Sub:           subject,
							},
						},
					}
			},
		},
		{
			name: "Verifies token using per-verifier skipEmailVerifiedCheck",
			setup: func(t *testing.T) (auth.OIDCConfigs, []verification) {
				signer, iss := oidctest.NewIssuer(t)
				token := token(t, signer, iss, auth.Claims{
					Email: "example@artefactual.com",
				})

				return auth.OIDCConfigs{
						{
							ProviderURL:            iss,
							ClientID:               audience,
							SkipEmailVerifiedCheck: false,
						},
						{
							ProviderURL:            iss,
							ClientID:               audience,
							SkipEmailVerifiedCheck: true,
						},
					}, []verification{
						{
							token: token,
							wantClaims: &auth.Claims{
								Email: "example@artefactual.com",
								Iss:   iss,
								Sub:   subject,
							},
						},
					}
			},
		},
		{
			name: "Verifies token when an earlier verifier has ABAC parsing error",
			setup: func(t *testing.T) (auth.OIDCConfigs, []verification) {
				signer, iss := oidctest.NewIssuer(t)
				token := token(t, signer, iss, map[string]any{
					"email":          "example@artefactual.com",
					"email_verified": true,
					"attributes":     []string{"*"},
				})

				return auth.OIDCConfigs{
						{
							ProviderURL: iss,
							ClientID:    audience,
							ABAC: auth.OIDCABACConfig{
								Enabled:   true,
								ClaimPath: "missing",
							},
						},
						{
							ProviderURL: iss,
							ClientID:    audience,
							ABAC: auth.OIDCABACConfig{
								Enabled:   true,
								ClaimPath: "attributes",
							},
						},
					}, []verification{
						{
							token: token,
							wantClaims: &auth.Claims{
								Email:         "example@artefactual.com",
								EmailVerified: true,
								Iss:           iss,
								Sub:           subject,
								Attributes:    []string{"*"},
							},
						},
					}
			},
		},
		{
			name: "Returns joined errors when all verifiers fail with non authorization errors",
			setup: func(t *testing.T) (auth.OIDCConfigs, []verification) {
				signer, iss := oidctest.NewIssuer(t)
				token := token(t, signer, iss, auth.Claims{
					Email:         "example@artefactual.com",
					EmailVerified: true,
				})

				return auth.OIDCConfigs{
						{ProviderURL: iss, ClientID: "wrong-audience-a"},
						{ProviderURL: iss, ClientID: "wrong-audience-b"},
					}, []verification{
						{
							token: token,
							wantErrContains: []string{
								`oidc: expected audience "wrong-audience-a" got ["test-audience"]`,
								`oidc: expected audience "wrong-audience-b" got ["test-audience"]`,
							},
						},
					}
			},
		},
		{
			name: "Returns unauthorized when all verifiers fail with unauthorized",
			setup: func(t *testing.T) (auth.OIDCConfigs, []verification) {
				signer, iss := oidctest.NewIssuer(t)
				token := token(t, signer, iss, auth.Claims{
					Email: "example@artefactual.com",
				})

				return auth.OIDCConfigs{
						{ProviderURL: iss, ClientID: audience},
						{ProviderURL: iss, ClientID: audience},
					}, []verification{
						{
							token:     token,
							wantErrIs: auth.ErrUnauthorized,
						},
					}
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfgs, verifications := tt.setup(t)
			ctx := context.Background()
			v, err := auth.NewOIDCTokenVerifiers(ctx, cfgs)
			assert.NilError(t, err)

			for _, verify := range verifications {
				claims, err := v.Verify(ctx, verify.token)
				if len(verify.wantErrContains) > 0 {
					assert.Assert(t, cmp.Nil(claims))
					for _, wantErr := range verify.wantErrContains {
						assert.ErrorContains(t, err, wantErr)
					}
					continue
				}
				if verify.wantErrIs != nil {
					assert.Assert(t, cmp.Nil(claims))
					assert.ErrorIs(t, err, verify.wantErrIs)
					continue
				}

				assert.NilError(t, err)
				assert.DeepEqual(t, claims, verify.wantClaims)
			}
		})
	}

	t.Run("Fails when no OIDC configs are provided", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		_, err := auth.NewOIDCTokenVerifiers(ctx, nil)
		assert.Error(t, err, "missing OIDC token verifier configuration")
	})

	t.Run("Constructor fails when context is canceled", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := auth.NewOIDCTokenVerifiers(ctx, auth.OIDCConfigs{{
			ProviderURL: "http://test",
		}})
		assert.Error(t, err, "Get \"http://test/.well-known/openid-configuration\": context canceled")
	})
}

func TestParseAttributes(t *testing.T) {
	t.Parallel()

	signer, iss := oidctest.NewIssuer(t)

	type nestedAttr struct {
		NestedAttributes any `json:"nested_attributes,omitempty"`
	}

	type customClaims struct {
		Email         string `json:"email,omitempty"`
		EmailVerified bool   `json:"email_verified,omitempty"`
		Attributes    any    `json:"attributes,omitempty"`
	}

	type test struct {
		name       string
		config     auth.OIDCConfigs
		token      string
		wantClaims *auth.Claims
		wantErr    string
	}
	for _, tt := range []test{
		{
			name: "Parses attributes based on configuration",
			config: auth.OIDCConfigs{{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:   true,
					ClaimPath: "attributes",
				},
			}},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{"*"},
			}),
			wantClaims: &auth.Claims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Iss:           iss,
				Sub:           subject,
				Attributes:    []string{"*"},
			},
		},
		{
			name: "Parses attributes based on configuration (disabled)",
			config: auth.OIDCConfigs{{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled: false,
				},
			}},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{"*"},
			}),
			wantClaims: &auth.Claims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Iss:           iss,
				Sub:           subject,
				Attributes:    nil,
			},
		},
		{
			name: "Parses attributes based on configuration (no attributes)",
			config: auth.OIDCConfigs{{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:   true,
					ClaimPath: "attributes",
				},
			}},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{},
			}),
			wantClaims: &auth.Claims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Iss:           iss,
				Sub:           subject,
				Attributes:    []string{},
			},
		},
		{
			name: "Parses attributes based on configuration (nested)",
			config: auth.OIDCConfigs{{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:            true,
					ClaimPath:          "attributes.nested_attributes",
					ClaimPathSeparator: ".",
				},
			}},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    nestedAttr{NestedAttributes: []string{"*"}},
			}),
			wantClaims: &auth.Claims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Iss:           iss,
				Sub:           subject,
				Attributes:    []string{"*"},
			},
		},
		{
			name: "Parses attributes based on configuration (filtering by prefix)",
			config: auth.OIDCConfigs{{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:          true,
					ClaimPath:        "attributes",
					ClaimValuePrefix: "enduro:",
				},
			}},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{"enduro:*", "ignore:*"},
			}),
			wantClaims: &auth.Claims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Iss:           iss,
				Sub:           subject,
				Attributes:    []string{"*"},
			},
		},
		{
			name: "Parses attributes based on configuration (mapping roles)",
			config: auth.OIDCConfigs{{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:   true,
					ClaimPath: "attributes",
					UseRoles:  true,
					RolesMapping: map[string][]string{
						"admin": {"*"},
						"operator": {
							auth.IngestSIPSListAttr,
							auth.IngestSIPSReadAttr,
							auth.IngestSIPSUploadAttr,
							auth.IngestSIPSWorkflowsListAttr,
						},
						"readonly": {
							auth.IngestSIPSListAttr,
							auth.IngestSIPSReadAttr,
							auth.IngestSIPSWorkflowsListAttr,
						},
					},
				},
			}},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{"admin", "operator", "readonly"},
			}),
			wantClaims: &auth.Claims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Iss:           iss,
				Sub:           subject,
				Attributes: []string{
					"*",
					auth.IngestSIPSListAttr,
					auth.IngestSIPSReadAttr,
					auth.IngestSIPSUploadAttr,
					auth.IngestSIPSWorkflowsListAttr,
				},
			},
		},
		{
			name: "Parses attributes based on configuration (mapping roles, no attributes)",
			config: auth.OIDCConfigs{{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:      true,
					ClaimPath:    "attributes",
					UseRoles:     true,
					RolesMapping: map[string][]string{"admin": {"*"}},
				},
			}},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{"other", "random", "role"},
			}),
			wantClaims: &auth.Claims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Iss:           iss,
				Sub:           subject,
				Attributes:    []string{},
			},
		},
		{
			name: "Fails to parse attributes (missing claim)",
			config: auth.OIDCConfigs{{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:   true,
					ClaimPath: "attributes",
				},
			}},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
			}),
			wantErr: "attributes not found in token, claim path: attributes",
		},
		{
			name: "Fails to parse attributes (missing nested claim)",
			config: auth.OIDCConfigs{{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:            true,
					ClaimPath:          "attributes.nested_attributes",
					ClaimPathSeparator: ".",
				},
			}},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    nestedAttr{},
			}),
			wantErr: "attributes not found in token, claim path: attributes.nested_attributes",
		},
		{
			name: "Fails to parse attributes (non multivalue claim)",
			config: auth.OIDCConfigs{{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:   true,
					ClaimPath: "attributes",
				},
			}},
			token: token(t, signer, iss, customClaims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    "*",
			}),
			wantErr: "attributes are not part of a multivalue claim, claim path: attributes",
		},
		{
			name: "Fails to parse attributes (expected nested claim)",
			config: auth.OIDCConfigs{{
				ProviderURL: iss,
				ClientID:    audience,
				ABAC: auth.OIDCABACConfig{
					Enabled:            true,
					ClaimPath:          "attributes.nested_attributes",
					ClaimPathSeparator: ".",
				},
			}},
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
			v, err := auth.NewOIDCTokenVerifiers(ctx, tt.config)
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
