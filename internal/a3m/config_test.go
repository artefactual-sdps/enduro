package a3m_test

import (
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/config"
)

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		config  string
		wantErr string
	}

	for _, tc := range []test{
		{
			name: "Valid config passes validation",
			config: `[a3m.processing]
aipCompressionLevel = 0
`,
		},
		{
			name: "Error if AipCompressionLevel is less than minimum",
			config: `[a3m.processing]
aipCompressionLevel = -1`,
			wantErr: "failed to validate the provided config: AipCompressionLevel: -1 is outside valid range (0 to 9)",
		},
		{
			name: "Error if AipCompressionLevel is greater than maximum",
			config: `[a3m.processing]
aipCompressionLevel = 10`,
			wantErr: "failed to validate the provided config: AipCompressionLevel: 10 is outside valid range (0 to 9)",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := fs.NewDir(t, "",
				fs.WithFile("enduro.toml", tc.config),
			)
			configFile := tmpDir.Join("enduro.toml")

			var c config.Configuration
			found, configFileUsed, err := config.Read(&c, configFile)
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, found, true)
			assert.Equal(t, configFileUsed, configFile)
		})
	}
}
