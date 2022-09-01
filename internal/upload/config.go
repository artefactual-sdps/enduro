package upload

import (
	"errors"
)

type Config struct {
	URL       string
	Region    string
	Endpoint  string
	PathStyle bool
	Profile   string
	Key       string
	Secret    string
	Token     string
	Bucket    string
}

// Validate implements config.ConfigurationValidator.
func (c Config) Validate() error {
	if c.URL != "" && (c.Bucket != "" || c.Region != "") {
		return errors.New("URL and rest of the [upload] configuration options are mutually exclusive")
	}
	return nil
}
