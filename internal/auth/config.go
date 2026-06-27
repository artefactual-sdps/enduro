package auth

import (
	"errors"
)

type Config struct {
	Enabled bool
	OIDC    OIDCConfigs
	Ticket  *TicketConfig
}

type OIDCConfigs []OIDCConfig

type OIDCConfig struct {
	ProviderURL            string
	ClientID               string
	SkipEmailVerifiedCheck bool

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
	if len(c.OIDC) == 0 {
		return errors.New("OIDC configuration required when API auth is enabled")
	}

	for i := range c.OIDC {
		if c.OIDC[i].ProviderURL == "" {
			return errors.New("OIDC provider URL required")
		}
		if c.OIDC[i].ClientID == "" {
			return errors.New("OIDC client ID required")
		}
		if c.OIDC[i].ABAC.Enabled && c.OIDC[i].ABAC.ClaimPath == "" {
			return errors.New("OIDC ABAC claim path required when ABAC is enabled")
		}
		if c.OIDC[i].ABAC.UseRoles && len(c.OIDC[i].ABAC.RolesMapping) == 0 {
			return errors.New("OIDC ABAC roles mapping required when use roles is enabled")
		}
	}

	return nil
}
