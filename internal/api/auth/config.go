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
	ProviderURL string
	ClientID    string

	// Attribute Based Access Control configuration.
	ABAC OIDCABACConfig
}

type OIDCABACConfig struct {
	Enabled            bool
	ClaimPath          string
	ClaimPathSeparator string
	ClaimValuePrefix   string
	UseRoles           bool
	RolesMapping       map[string][]string
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
	if !c.Enabled {
		return nil
	}
	if c.OIDC == nil || c.OIDC.ProviderURL == "" || c.OIDC.ClientID == "" {
		return errors.New("missing OIDC configuration with API auth. enabled")
	}
	if c.OIDC.ABAC.Enabled && c.OIDC.ABAC.ClaimPath == "" {
		return errors.New("missing OIDC ABAC claim path with ABAC enabled")
	}
	if c.OIDC.ABAC.UseRoles && len(c.OIDC.ABAC.RolesMapping) == 0 {
		return errors.New("missing OIDC ABAC roles mapping with use roles enabled")
	}
	return nil
}
