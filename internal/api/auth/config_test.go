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
				OIDC: &auth.OIDCConfig{
					ProviderURL: "http://keycloak:7470/realms/artefactual",
					ClientID:    "enduro",
				},
			},
		},
		{
			name: "Passes validation (enabled, ABAC enabled)",
			config: &auth.Config{
				Enabled: true,
				OIDC: &auth.OIDCConfig{
					ProviderURL: "http://keycloak:7470/realms/artefactual",
					ClientID:    "enduro",
					ABAC: auth.OIDCABACConfig{
						Enabled:   true,
						ClaimPath: "enduro",
					},
				},
			},
		},
		{
			name: "Fails validation (missing OIDC entire config)",
			config: &auth.Config{
				Enabled: true,
			},
			wantErr: "missing OIDC configuration with API auth. enabled",
		},
		{
			name: "Fails validation (missing OIDC config values)",
			config: &auth.Config{
				Enabled: true,
				OIDC:    &auth.OIDCConfig{},
			},
			wantErr: "missing OIDC configuration with API auth. enabled",
		},
		{
			name: "Fails validation (missing OIDC ABAC config values)",
			config: &auth.Config{
				Enabled: true,
				OIDC: &auth.OIDCConfig{
					ProviderURL: "http://keycloak:7470/realms/artefactual",
					ClientID:    "enduro",
					ABAC: auth.OIDCABACConfig{
						Enabled: true,
					},
				},
			},
			wantErr: "missing OIDC ABAC claim path with ABAC enabled",
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
