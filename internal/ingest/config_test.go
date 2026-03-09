package ingest_test

import (
	"testing"
	"time"

	"go.artefactual.dev/tools/clientauth"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/ingest"
)

func TestStorageConfigValidate(t *testing.T) {
	t.Parallel()

	t.Run("Requires address and default permanent location ID", func(t *testing.T) {
		t.Parallel()

		err := (ingest.StorageConfig{}).Validate()
		assert.ErrorContains(t, err, "missing storage API address")
		assert.ErrorContains(t, err, "missing storage default permanent location ID")
	})

	t.Run("Joins storage and OIDC validation errors", func(t *testing.T) {
		t.Parallel()

		cfg := ingest.StorageConfig{
			OIDC: ingest.StorageOIDCConfig{
				Enabled: true,
			},
		}

		err := cfg.Validate()
		assert.ErrorContains(t, err, "missing storage API address")
		assert.ErrorContains(t, err, "missing storage default permanent location ID")
		assert.ErrorContains(t, err, "storage OIDC:")
		assert.ErrorContains(t, err, "missing OIDC providerURL or tokenURL")
		assert.ErrorContains(t, err, "missing OIDC client credentials")
	})
}

func TestStorageOIDCConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     ingest.StorageOIDCConfig
		wantErr string
	}{
		{
			name: "Passes validation when disabled",
			cfg: ingest.StorageOIDCConfig{
				Enabled: false,
			},
		},
		{
			name: "Passes validation with provider URL and client credentials",
			cfg: ingest.StorageOIDCConfig{
				Enabled: true,
				OIDCAccessTokenProviderConfig: clientauth.OIDCAccessTokenProviderConfig{ // #nosec G101 -- test-only placeholder credential.
					ProviderURL:  "https://idp.example.com/realms/enduro",
					ClientID:     "enduro-worker",
					ClientSecret: "placeholder-value",
				},
			},
		},
		{
			name: "Passes validation with token URL and client credentials",
			cfg: ingest.StorageOIDCConfig{
				Enabled: true,
				OIDCAccessTokenProviderConfig: clientauth.OIDCAccessTokenProviderConfig{ // #nosec G101 -- test-only placeholder credential.
					TokenURL:     "https://idp.example.com/token",
					ClientID:     "enduro-worker",
					ClientSecret: "placeholder-value",
				},
			},
		},
		{
			name: "Fails validation when both providerURL and tokenURL are missing",
			cfg: ingest.StorageOIDCConfig{
				Enabled: true,
				OIDCAccessTokenProviderConfig: clientauth.OIDCAccessTokenProviderConfig{ // #nosec G101 -- test-only placeholder credential.
					ClientID:     "enduro-worker",
					ClientSecret: "placeholder-value",
				},
			},
			wantErr: "storage OIDC:\nmissing OIDC providerURL or tokenURL",
		},
		{
			name: "Fails validation when client credentials are missing",
			cfg: ingest.StorageOIDCConfig{
				Enabled: true,
				OIDCAccessTokenProviderConfig: clientauth.OIDCAccessTokenProviderConfig{
					ProviderURL: "https://idp.example.com/realms/enduro",
				},
			},
			wantErr: "storage OIDC:\nmissing OIDC client credentials",
		},
		{
			name: "Fails validation with invalid retry attempts",
			cfg: ingest.StorageOIDCConfig{
				Enabled: true,
				OIDCAccessTokenProviderConfig: clientauth.OIDCAccessTokenProviderConfig{ // #nosec G101 -- test-only placeholder credential.
					ProviderURL:      "https://idp.example.com/realms/enduro",
					ClientID:         "enduro-worker",
					ClientSecret:     "placeholder-value",
					RetryMaxAttempts: -1,
				},
			},
			wantErr: "storage OIDC:\ninvalid OIDC retry max attempts, value must be >= 1",
		},
		{
			name: "Fails validation with invalid retry backoff coefficient",
			cfg: ingest.StorageOIDCConfig{
				Enabled: true,
				OIDCAccessTokenProviderConfig: clientauth.OIDCAccessTokenProviderConfig{ // #nosec G101 -- test-only placeholder credential.
					ProviderURL:             "https://idp.example.com/realms/enduro",
					ClientID:                "enduro-worker",
					ClientSecret:            "placeholder-value",
					RetryBackoffCoefficient: 0.5,
					RetryMaxAttempts:        3,
					RetryInitialInterval:    1 * time.Millisecond,
					RetryMaxInterval:        2 * time.Millisecond,
				},
			},
			wantErr: "storage OIDC:\ninvalid OIDC retry backoff coefficient, value must be >= 1",
		},
		{
			name: "Joins multiple OIDC validation errors",
			cfg: ingest.StorageOIDCConfig{
				Enabled: true,
				OIDCAccessTokenProviderConfig: clientauth.OIDCAccessTokenProviderConfig{
					RetryMaxAttempts:        -1,
					RetryBackoffCoefficient: 0.5,
				},
			},
			wantErr: "storage OIDC:\nmissing OIDC providerURL or tokenURL\nmissing OIDC client credentials\ninvalid OIDC retry max attempts, value must be >= 1\ninvalid OIDC retry backoff coefficient, value must be >= 1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.cfg.Validate()
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
		})
	}
}
