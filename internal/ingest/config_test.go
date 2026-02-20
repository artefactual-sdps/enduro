package ingest_test

import (
	"testing"
	"time"

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
		assert.ErrorContains(t, err, "missing OIDC providerURL or tokenURL with storage OIDC auth. enabled")
		assert.ErrorContains(t, err, "missing OIDC client credentials with storage OIDC auth. enabled")
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
				Enabled:      true,
				ProviderURL:  "https://idp.example.com/realms/enduro",
				ClientID:     "enduro-worker",
				ClientSecret: "secret",
			},
		},
		{
			name: "Passes validation with token URL and client credentials",
			cfg: ingest.StorageOIDCConfig{
				Enabled:      true,
				TokenURL:     "https://idp.example.com/token",
				ClientID:     "enduro-worker",
				ClientSecret: "secret",
			},
		},
		{
			name: "Fails validation when both providerURL and tokenURL are missing",
			cfg: ingest.StorageOIDCConfig{
				Enabled:      true,
				ClientID:     "enduro-worker",
				ClientSecret: "secret",
			},
			wantErr: "missing OIDC providerURL or tokenURL with storage OIDC auth. enabled",
		},
		{
			name: "Fails validation when client credentials are missing",
			cfg: ingest.StorageOIDCConfig{
				Enabled:     true,
				ProviderURL: "https://idp.example.com/realms/enduro",
			},
			wantErr: "missing OIDC client credentials with storage OIDC auth. enabled",
		},
		{
			name: "Fails validation with invalid retry attempts",
			cfg: ingest.StorageOIDCConfig{
				Enabled:          true,
				ProviderURL:      "https://idp.example.com/realms/enduro",
				ClientID:         "enduro-worker",
				ClientSecret:     "secret",
				RetryMaxAttempts: -1,
			},
			wantErr: "invalid storage OIDC retry max attempts, value must be >= 0",
		},
		{
			name: "Fails validation with invalid retry backoff coefficient",
			cfg: ingest.StorageOIDCConfig{
				Enabled:                 true,
				ProviderURL:             "https://idp.example.com/realms/enduro",
				ClientID:                "enduro-worker",
				ClientSecret:            "secret",
				RetryBackoffCoefficient: 0.5,
				RetryMaxAttempts:        3,
				RetryInitialInterval:    1 * time.Millisecond,
				RetryMaxInterval:        2 * time.Millisecond,
			},
			wantErr: "invalid storage OIDC retry backoff coefficient, value must be >= 1",
		},
		{
			name: "Joins multiple OIDC validation errors",
			cfg: ingest.StorageOIDCConfig{
				Enabled:                 true,
				RetryMaxAttempts:        -1,
				RetryBackoffCoefficient: 0.5,
			},
			wantErr: "missing OIDC providerURL or tokenURL with storage OIDC auth. enabled\nmissing OIDC client credentials with storage OIDC auth. enabled\ninvalid storage OIDC retry max attempts, value must be >= 0\ninvalid storage OIDC retry backoff coefficient, value must be >= 1",
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
