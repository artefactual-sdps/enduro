package premis

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	Enabled bool
	XSDPath string
}

// Validate implements config.ConfigurationValidator.
func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.XSDPath == "" {
		return errors.New("xsdPath is required in the [validatePremis] configuration when enabled")
	}
	if _, err := os.Stat(c.XSDPath); err != nil {
		return fmt.Errorf("xsdPath in [validatePremis] not found: %v", err)
	}
	return nil
}
