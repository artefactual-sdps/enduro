package ingest

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/clientauth"
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
	Enabled                                  bool
	clientauth.OIDCAccessTokenProviderConfig `mapstructure:",squash"`
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

	if err := c.OIDCAccessTokenProviderConfig.Validate(); err != nil {
		return fmt.Errorf("storage OIDC:\n%v", err)
	}

	return nil
}
