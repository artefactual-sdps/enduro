package sftp

import "path/filepath"

type Config struct {
	Host string
	Port string

	KnownHostsFile string
	PrivateKey     PrivateKey
}

type PrivateKey struct {
	Path       string
	Passphrase string
}

// SetDefaults sets default values for some configs.
func (c *Config) SetDefaults() {
	if c.Host == "" {
		c.Host = "localhost"
	}

	if c.Port == "" {
		c.Port = "22"
	}

	if c.KnownHostsFile == "" {
		c.KnownHostsFile = filepath.Join("$HOME", ".ssh", "known_hosts")
	}

	if c.PrivateKey.Path == "" {
		c.PrivateKey.Path = filepath.Join("$HOME", ".ssh", "id_rsa")
	}
}
