package bagit

import (
	"errors"
)

type ValidatorConfig struct {
	// # cacheDir sets the cache directory used to write runtime artifacts for
	// the BagIt validator's runtime and runners validator's runtime and
	// runners. If CacheDir is an empty string or omitted, the validator will
	// attempt to create a cache directory in the process user's home directory
	// (e.g. /home/enduro/.cache/bagit-gython). If the user's home directory is
	// not available (e.g. because the process has no home directory), the
	// validator will fall back to using a unique temporary directory
	// (e.g. /tmp/bagit-gython-12345) that will be deleted at shutdown.
	CacheDir string `mapstructure:"cacheDir"`

	// PoolSize sets the number of bag validation runners available to
	// concurrently validate bags. PoolSize must be 1 (the default value) or
	// greater. If the number of requested validation jobs exceeds the available
	// runners, the extra jobs will be queued and run when a runner becomes
	// available. See
	// https://github.com/artefactual-labs/bagit-gython/blob/main/README.md for
	// more details on how the validator pool works.
	PoolSize int `mapstructure:"poolSize"`
}

func (c *ValidatorConfig) Validate() error {
	if c.PoolSize < 1 {
		return errors.New("bagit.validator.poolSize must be 1 or greater")
	}

	return nil
}
