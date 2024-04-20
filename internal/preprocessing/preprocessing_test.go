package preprocessing_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/preprocessing"
)

func TestPreprocessingConfig(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		config  preprocessing.Config
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Validates if not enabled",
			config: preprocessing.Config{
				Enabled: false,
			},
		},
		{
			name: "Validates with all required fields",
			config: preprocessing.Config{
				Enabled:    true,
				SharedPath: "/tmp",
				Temporal: preprocessing.Temporal{
					Namespace:    "default",
					TaskQueue:    "preprocessing",
					WorkflowName: "preprocessing",
				},
			},
		},
		{
			name: "Returns error if shared path is missing",
			config: preprocessing.Config{
				Enabled: true,
			},
			wantErr: "sharedPath is required in the [preprocessing] configuration",
		},
		{
			name: "Returns error if temporal config is missing",
			config: preprocessing.Config{
				Enabled:    true,
				SharedPath: "/tmp",
			},
			wantErr: "namespace, taskQueue and workflowName are required in the [preprocessing.temporal] configuration",
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
