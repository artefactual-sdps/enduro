package config_test

import (
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/pkg/childwf"
)

func TestChildWorkflowConfig_ReadFromTOML(t *testing.T) {
	toml := `
[ingest.storage]
address = "storage-api:9000"
defaultPermanentLocationId = "f2cc963f-c14d-4eaa-b950-bd207189a1f1"

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
	assert.DeepEqual(t, c.ChildWorkflows, config.ChildWorkflowConfigs{
		{
			Type:         childwf.WorkflowTypePreprocessing,
			Namespace:    "default",
			TaskQueue:    "preprocessing",
			WorkflowName: "preprocessing",
			SharedPath:   "/home/enduro/shared",
		},
		{
			Type:         childwf.WorkflowTypePoststorage,
			Namespace:    "default",
			TaskQueue:    "poststorage",
			WorkflowName: "poststorage",
		},
	})
}

func TestChildWorkflowConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  config.ChildWorkflowConfig
		wantErr string
	}{
		{
			name: "Valid config",
			config: config.ChildWorkflowConfig{
				Type:         childwf.WorkflowTypePreprocessing,
				Namespace:    "default",
				TaskQueue:    "preprocessing",
				WorkflowName: "preprocessing",
				SharedPath:   "/home/enduro/shared",
			},
		},
		{
			name: "Errors on missing fields",
			config: config.ChildWorkflowConfig{
				Type: childwf.WorkflowTypePreprocessing,
			},
			wantErr: `missing required value(s): namespace, taskQueue, workflowName, sharedPath`,
		},
		{
			name: "Errors on invalid type",
			config: config.ChildWorkflowConfig{
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

func TestChildWorkflowConfigs_ByType(t *testing.T) {
	configs := config.ChildWorkflowConfigs{
		{
			Type:         childwf.WorkflowTypePreprocessing,
			Namespace:    "default",
			TaskQueue:    "preprocessing",
			WorkflowName: "preprocessing",
			SharedPath:   "/home/enduro/shared",
		},
		{
			Type:         childwf.WorkflowTypePoststorage,
			Namespace:    "default",
			TaskQueue:    "poststorage",
			WorkflowName: "poststorage",
		},
	}

	cfg := configs.ByType(childwf.WorkflowTypePreprocessing)
	assert.DeepEqual(t, cfg, &config.ChildWorkflowConfig{
		Type:         childwf.WorkflowTypePreprocessing,
		Namespace:    "default",
		TaskQueue:    "preprocessing",
		WorkflowName: "preprocessing",
		SharedPath:   "/home/enduro/shared",
	})

	cfg = configs.ByType("nonexistent")
	assert.Equal(t, cfg, (*config.ChildWorkflowConfig)(nil))
}

func TestChildWorkflowConfigs_Validate(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		configs config.ChildWorkflowConfigs
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Valid configs",
			configs: config.ChildWorkflowConfigs{
				{
					Type:         childwf.WorkflowTypePreprocessing,
					Namespace:    "default",
					TaskQueue:    "preprocessing",
					WorkflowName: "preprocessing",
					SharedPath:   "/home/enduro/shared",
				},
				{
					Type:         childwf.WorkflowTypePoststorage,
					Namespace:    "default",
					TaskQueue:    "poststorage",
					WorkflowName: "poststorage",
					SharedPath:   "/home/enduro/shared",
				},
			},
		},
		{
			name: "Errors on duplicate type",
			configs: config.ChildWorkflowConfigs{
				{
					Type:         childwf.WorkflowTypePreprocessing,
					Namespace:    "default",
					TaskQueue:    "preprocessing",
					WorkflowName: "preprocessing",
					SharedPath:   "/home/enduro/shared",
				},
				{
					Type:         childwf.WorkflowTypePreprocessing,
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
			configs: config.ChildWorkflowConfigs{
				{
					Type:         childwf.WorkflowTypePreprocessing,
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
