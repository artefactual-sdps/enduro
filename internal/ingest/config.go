package ingest

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/clientauth"
)

type Config struct {
	// AllowDuplicates toggles whether a SIP can be ingested more than once.
	// The default value (false) will stop ingest with a content error when a
	// SIP archive file (e.g. zip) is submitted that has the same checksum as a
	// previously ingested SIP. SIPs submitted as a directory are not checked
	// for duplicate contents.
	//
	// A SIP is only considered a duplicate if the checksum matches an existing
	// SIP with a status of: "ingested", "pending", "processing", "queued", or
	// "validated". If the SIP status is "error", "failed" or "canceled" the
	// SIP will be ignored when checking for duplicates.
	//
	// A checksum is calculated and stored for every SIP archive ingested by
	// Enduro, regardless of this setting. When `allowDuplicates` is false, a
	// new ingest's checksum will be checked against all previously ingested SIP
	// checksums, even if `allowDuplicates` was true when the old SIPs were
	// ingested.
	AllowDuplicates bool

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
