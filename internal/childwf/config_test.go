package childwf_test

import (
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/childwf"
	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/enums"
)

func TestConfig_ReadFromTOML(t *testing.T) {
	toml := `
[[childWorkflows]]
type = "preprocessing"
namespace = "default"
taskQueue = "preprocessing"
workflowName = "preprocessing"
sharedPath = "/home/enduro/shared"

[[childWorkflows]]
type = "poststorage"
namespace = "default"
taskQueue = "poststorage"
workflowName = "poststorage"
`

	tmpDir := fs.NewDir(t, "",
		fs.WithFile("enduro.toml", toml),
	)
	configFile := tmpDir.Join("enduro.toml")

	var c config.Configuration
	_, _, err := config.Read(&c, configFile)

	assert.NilError(t, err)
	assert.DeepEqual(t, c.ChildWorkflows, childwf.Configs{
		{
			Type:         enums.ChildWorkflowTypePreprocessing,
			Namespace:    "default",
			TaskQueue:    "preprocessing",
			WorkflowName: "preprocessing",
			SharedPath:   "/home/enduro/shared",
		},
		{
			Type:         enums.ChildWorkflowTypePoststorage,
			Namespace:    "default",
			TaskQueue:    "poststorage",
			WorkflowName: "poststorage",
		},
	})
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  childwf.Config
		wantErr string
	}{
		{
			name: "Valid config",
			config: childwf.Config{
				Type:         enums.ChildWorkflowTypePreprocessing,
				Namespace:    "default",
				TaskQueue:    "preprocessing",
				WorkflowName: "preprocessing",
				SharedPath:   "/home/enduro/shared",
			},
		},
		{
			name: "Errors on missing fields",
			config: childwf.Config{
				Type: enums.ChildWorkflowTypePreprocessing,
			},
			wantErr: `missing required value(s): namespace, taskQueue, workflowName, sharedPath`,
		},
		{
			name: "Errors on invalid type",
			config: childwf.Config{
				Type:         "invalid_type",
				Namespace:    "default",
				TaskQueue:    "taskqueue",
				WorkflowName: "workflowname",
			},
			wantErr: `invalid type: invalid_type`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func TestConfigs_ByType(t *testing.T) {
	configs := childwf.Configs{
		{
			Type:         enums.ChildWorkflowTypePreprocessing,
			Namespace:    "default",
			TaskQueue:    "preprocessing",
			WorkflowName: "preprocessing",
			SharedPath:   "/home/enduro/shared",
		},
		{
			Type:         enums.ChildWorkflowTypePoststorage,
			Namespace:    "default",
			TaskQueue:    "poststorage",
			WorkflowName: "poststorage",
		},
	}

	cfg := configs.ByType(enums.ChildWorkflowTypePreprocessing)
	assert.DeepEqual(t, cfg, &childwf.Config{
		Type:         enums.ChildWorkflowTypePreprocessing,
		Namespace:    "default",
		TaskQueue:    "preprocessing",
		WorkflowName: "preprocessing",
		SharedPath:   "/home/enduro/shared",
	})

	cfg = configs.ByType("nonexistent")
	assert.Equal(t, cfg, (*childwf.Config)(nil))
}

func TestConfigs_Validate(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		configs childwf.Configs
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Valid configs",
			configs: childwf.Configs{
				{
					Type:         enums.ChildWorkflowTypePreprocessing,
					Namespace:    "default",
					TaskQueue:    "preprocessing",
					WorkflowName: "preprocessing",
					SharedPath:   "/home/enduro/shared",
				},
				{
					Type:         enums.ChildWorkflowTypePoststorage,
					Namespace:    "default",
					TaskQueue:    "poststorage",
					WorkflowName: "poststorage",
					SharedPath:   "/home/enduro/shared",
				},
			},
		},
		{
			name: "Errors on duplicate type",
			configs: childwf.Configs{
				{
					Type:         enums.ChildWorkflowTypePreprocessing,
					Namespace:    "default",
					TaskQueue:    "preprocessing",
					WorkflowName: "preprocessing",
					SharedPath:   "/home/enduro/shared",
				},
				{
					Type:         enums.ChildWorkflowTypePreprocessing,
					Namespace:    "default",
					TaskQueue:    "preprocessing-2",
					WorkflowName: "preprocessing-2",
					SharedPath:   "/home/enduro/shared-2",
				},
			},
			wantErr: "child workflow[1]: duplicate type: preprocessing",
		},
		{
			name: "Errors on missing config values",
			configs: childwf.Configs{
				{
					Type:         enums.ChildWorkflowTypePreprocessing,
					Namespace:    "default",
					TaskQueue:    "preprocessing",
					WorkflowName: "preprocessing",
				},
				{
					Namespace:    "default",
					TaskQueue:    "poststorage",
					WorkflowName: "poststorage",
				},
			},
			wantErr: `child workflow[0]: missing required value(s): sharedPath
child workflow[1]: missing required value(s): type`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.configs.Validate()
			if tt.wantErr == "" {
				assert.NilError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
