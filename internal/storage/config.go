package storage

import (
	"errors"
	"fmt"
)

type Config struct {
	// TODO: internal location to use its own type.
	Region    string
	Endpoint  string
	PathStyle bool
	Profile   string
	Key       string
	Secret    string
	Token     string
	Bucket    string

	EnduroAddress string

	Locations []LocationConfig `mapstructure:"location"`
}

// Validate implements config.ConfigurationValidator.
func (c Config) Validate() error {
	index := map[string]bool{}
	for _, item := range c.Locations {
		if item.Name == "" {
			return errors.New("location name is undefined")
		}
		if _, ok := index[item.Name]; ok {
			return fmt.Errorf("location with name %s already defined", item.Name)
		}
		index[item.Name] = true
	}
	return nil
}

type LocationConfig struct {
	Name      string
	Region    string
	Endpoint  string
	PathStyle bool
	Profile   string
	Key       string
	Secret    string
	Token     string
	Bucket    string
}
