package ingest

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultStorageOIDCTokenExpiryLeeway       = 30 * time.Second
	DefaultStorageOIDCRetryMaxAttempts        = 3
	DefaultStorageOIDCRetryInitialInterval    = 500 * time.Millisecond
	DefaultStorageOIDCRetryMaxInterval        = 2 * time.Second
	DefaultStorageOIDCRetryBackoffCoefficient = 2.0
)

type Config struct {
	Storage StorageConfig
}

type StorageConfig struct {
	Address                    string
	DefaultPermanentLocationID uuid.UUID
	OIDC                       StorageOIDCConfig
}

type StorageOIDCConfig struct {
	Enabled                 bool
	ProviderURL             string
	TokenURL                string
	ClientID                string
	ClientSecret            string
	Scopes                  []string
	Audience                string
	TokenExpiryLeeway       time.Duration
	RetryMaxAttempts        int
	RetryInitialInterval    time.Duration
	RetryMaxInterval        time.Duration
	RetryBackoffCoefficient float64
}

func (c Config) Validate() error {
	return c.Storage.Validate()
}

func (c StorageConfig) Validate() error {
	var errs []error

	if strings.TrimSpace(c.Address) == "" {
		errs = append(errs, errors.New("missing storage API address"))
	}
	if c.DefaultPermanentLocationID == uuid.Nil {
		errs = append(errs, errors.New("missing storage default permanent location ID"))
	}

	if err := c.OIDC.Validate(); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (c StorageOIDCConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	var errs []error
	if c.ProviderURL == "" && c.TokenURL == "" {
		errs = append(errs, errors.New("missing OIDC providerURL or tokenURL with storage OIDC auth. enabled"))
	}
	if c.ClientID == "" || c.ClientSecret == "" {
		errs = append(errs, errors.New("missing OIDC client credentials with storage OIDC auth. enabled"))
	}
	if c.RetryMaxAttempts < 0 {
		errs = append(errs, errors.New("invalid storage OIDC retry max attempts, value must be >= 0"))
	}
	if c.RetryInitialInterval < 0 || c.RetryMaxInterval < 0 || c.TokenExpiryLeeway < 0 {
		errs = append(errs, errors.New("invalid storage OIDC duration configuration, values must be >= 0"))
	}
	if c.RetryInitialInterval > 0 && c.RetryMaxInterval > 0 && c.RetryMaxInterval < c.RetryInitialInterval {
		errs = append(errs, errors.New(
			"invalid storage OIDC retry interval configuration, max interval must be >= initial interval",
		))
	}
	if c.RetryBackoffCoefficient != 0 && c.RetryBackoffCoefficient < 1 {
		errs = append(errs, errors.New("invalid storage OIDC retry backoff coefficient, value must be >= 1"))
	}

	return errors.Join(errs...)
}
