package auth_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		config  *auth.Config
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Passes validation (disabled)",
			config: &auth.Config{
				Enabled: false,
			},
		},
		{
			name: "Passes validation (enabled, ABAC disabled)",
			config: &auth.Config{
				Enabled: true,
				OIDC: auth.OIDCConfigs{{
					ProviderURL: "http://keycloak:7470/realms/artefactual",
					ClientID:    "enduro",
				}},
			},
		},
		{
			name: "Passes validation (enabled, ABAC enabled)",
			config: &auth.Config{
				Enabled: true,
				OIDC: auth.OIDCConfigs{{
					ProviderURL: "http://keycloak:7470/realms/artefactual",
					ClientID:    "enduro",
					ABAC: auth.OIDCABACConfig{
						Enabled:   true,
						ClaimPath: "enduro",
					},
				}},
			},
		},
		{
			name: "Passes validation (enabled, multiple OIDC configs)",
			config: &auth.Config{
				Enabled: true,
				OIDC: auth.OIDCConfigs{
					{
						ProviderURL: "http://keycloak:7470/realms/artefactual",
						ClientID:    "enduro",
						ABAC: auth.OIDCABACConfig{
							Enabled:   true,
							ClaimPath: "enduro",
						},
					},
					{
						ProviderURL: "http://keycloak:7470/realms/artefactual-internal",
						ClientID:    "enduro-s2s",
					},
				},
			},
		},
		{
			name: "Fails validation (missing OIDC entire config)",
			config: &auth.Config{
				Enabled: true,
			},
			wantErr: "OIDC configuration required when API auth is enabled",
		},
		{
			name: "Fails validation (missing OIDC provider URL)",
			config: &auth.Config{
				Enabled: true,
				OIDC:    auth.OIDCConfigs{{}},
			},
			wantErr: "OIDC provider URL required",
		},
		{
			name: "Fails validation (missing OIDC client ID)",
			config: &auth.Config{
				Enabled: true,
				OIDC: auth.OIDCConfigs{{
					ProviderURL: "http://keycloak:7470/realms/artefactual",
				}},
			},
			wantErr: "OIDC client ID required",
		},
		{
			name: "Fails validation (missing OIDC ABAC claim path)",
			config: &auth.Config{
				Enabled: true,
				OIDC: auth.OIDCConfigs{{
					ProviderURL: "http://keycloak:7470/realms/artefactual",
					ClientID:    "enduro",
					ABAC: auth.OIDCABACConfig{
						Enabled: true,
					},
				}},
			},
			wantErr: "OIDC ABAC claim path required when ABAC is enabled",
		},
		{
			name: "Fails validation (missing OIDC ABAC roles mapping)",
			config: &auth.Config{
				Enabled: true,
				OIDC: auth.OIDCConfigs{{
					ProviderURL: "http://keycloak:7470/realms/artefactual",
					ClientID:    "enduro",
					ABAC: auth.OIDCABACConfig{
						Enabled:   true,
						ClaimPath: "enduro",
						UseRoles:  true,
					},
				}},
			},
			wantErr: "OIDC ABAC roles mapping required when use roles is enabled",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
		})
	}
}
