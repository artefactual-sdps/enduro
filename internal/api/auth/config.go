package auth

import (
	"errors"
)

type Config struct {
	Enabled bool
	OIDC    *OIDCConfig
	Ticket  *TicketConfig
}

type OIDCConfig struct {
	ProviderURL  string
	ClientID     string
	ClientSecret string
}

type TicketConfig struct {
	Redis *RedisConfig
}

type RedisConfig struct {
	Address string
	Prefix  string
}

// Validate implements config.ConfigurationValidator.
func (c Config) Validate() error {
	if c.Enabled && c.OIDC == nil {
		return errors.New("Missing OIDC configuration with API auth. enabled.")
	}
	return nil
}
