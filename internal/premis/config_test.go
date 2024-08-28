package premis_test

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/premis"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		config  *premis.Config
		wantErr string
	}

	// Non-existent XSD path.
	badXSDfs := fs.NewDir(t, "", fs.WithFile("missing.xsd", ""))
	badXSDPath := badXSDfs.Join("missing.xsd")
	badXSDfs.Remove()

	for _, tt := range []test{
		{
			name: "Passes validation (disabled)",
			config: &premis.Config{
				Enabled: false,
			},
		},
		{
			name: "Fails validation (missing XSD path)",
			config: &premis.Config{
				Enabled: true,
			},
			wantErr: "xsdPath is required in the [validatePremis] configuration when enabled",
		},
		{
			name: "Fails validation (missing XSD file)",
			config: &premis.Config{
				Enabled: true,
				XSDPath: badXSDPath,
			},
			wantErr: fmt.Sprintf("xsdPath in [validatePremis] not found: stat %s: no such file or directory", badXSDPath),
		},
		{
			name: "Passes validation (enabled)",
			config: &premis.Config{
				Enabled: true,
				XSDPath: fs.NewDir(t, "", fs.WithFile("empty.xsd", "")).Join("empty.xsd"),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
		})
	}
}
