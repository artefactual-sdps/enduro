package bagit

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestValidatorConfig(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		config  ValidatorConfig
		wantErr string
	}
	tests := []test{
		{
			name: "valid config",
			config: ValidatorConfig{
				CacheDir: "/home/enduro/bagvalidator_cache",
				PoolSize: 2,
			},
		},
		{
			name: "invalid pool size",
			config: ValidatorConfig{
				PoolSize: 0,
			},
			wantErr: "bagit.validator.poolSize must be 1 or greater",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.config.Validate()
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}
			assert.NilError(t, err)
		})
	}
}
