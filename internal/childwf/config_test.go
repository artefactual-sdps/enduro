package childwf_test

import (
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/childwf"
	"github.com/artefactual-sdps/enduro/internal/config"
)

func TestConfig_ReadFromTOML(t *testing.T) {
	toml := `
[[childworkflows]]
enabled = true
namespace = "default"
taskQueue = "preprocessing"
workflowName = "preprocessing"

[[childworkflows]]
enabled = false
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
	assert.DeepEqual(t, c.Childworkflows, childwf.Configs{
		{
			Enabled:      true,
			Namespace:    "default",
			TaskQueue:    "preprocessing",
			WorkflowName: "preprocessing",
		},
		{
			Enabled:      false,
			Namespace:    "default",
			TaskQueue:    "poststorage",
			WorkflowName: "poststorage",
		},
	})
}

func TestConfig_GetByName(t *testing.T) {
	configs := childwf.Configs{
		{
			Enabled:      true,
			Namespace:    "default",
			TaskQueue:    "preprocessing",
			WorkflowName: "preprocessing",
		},
		{
			Enabled:      true,
			Namespace:    "default",
			TaskQueue:    "poststorage",
			WorkflowName: "poststorage",
		},
	}

	cfg := configs.GetByName("preprocessing")
	assert.DeepEqual(t, cfg, &childwf.Config{
		Enabled:      true,
		Namespace:    "default",
		TaskQueue:    "preprocessing",
		WorkflowName: "preprocessing",
	})

	cfg = configs.GetByName("nonexistent")
	assert.Equal(t, cfg, (*childwf.Config)(nil))
}

func TestConfig_IsEnabled(t *testing.T) {
	configs := childwf.Configs{
		{
			Enabled:      true,
			Namespace:    "default",
			TaskQueue:    "preprocessing",
			WorkflowName: "preprocessing",
		},
		{
			Enabled:      false,
			Namespace:    "default",
			TaskQueue:    "poststorage",
			WorkflowName: "poststorage",
		},
	}

	assert.Equal(t, true, configs.IsEnabled("preprocessing"))
	assert.Equal(t, false, configs.IsEnabled("poststorage"))
	assert.Equal(t, false, configs.IsEnabled("nonexistent"))
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		configs childwf.Configs
		wantErr string
	}{
		{
			name: "Valid configs",
			configs: childwf.Configs{
				{
					Enabled:      true,
					Namespace:    "default",
					TaskQueue:    "preprocessing",
					WorkflowName: "preprocessing",
				},
				{
					Enabled:      true,
					Namespace:    "default",
					TaskQueue:    "poststorage",
					WorkflowName: "poststorage",
				},
			},
		},
		{
			name: "Invalid configs - missing fields",
			configs: childwf.Configs{
				{
					Enabled:      true,
					Namespace:    "default",
					TaskQueue:    "preprocessing",
					WorkflowName: "preprocessing",
				},
				{
					Enabled: true,
				},
			},
			wantErr: "child workflows[1]: namespace is required\n" +
				"child workflows[1]: task queue is required\n" +
				"child workflows[1]: workflow name is required",
		},
		{
			name: "Disabled workflow - no validation",
			configs: childwf.Configs{
				{
					Enabled: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.configs.Validate()
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}
