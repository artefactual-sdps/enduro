package watcher_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/watcher"
)

func TestCompletedDirs(t *testing.T) {
	t.Parallel()

	c := watcher.Config{
		Filesystem: []*watcher.FilesystemConfig{
			{CompletedDir: ""},
			nil,
			{CompletedDir: "/tmp/test-1"},
			{CompletedDir: "/tmp/test-2"},
			{CompletedDir: "./test-3"},
		},
	}

	wd, _ := os.Getwd()
	assert.DeepEqual(t, c.CompletedDirs(), []string{
		"/tmp/test-1",
		"/tmp/test-2",
		filepath.Join(wd, "test-3"),
	})
}

func TestValidate(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		config     watcher.Config
		wantConfig watcher.Config
		wantErr    string
	}{
		{
			name: "Sets defaults",
			config: watcher.Config{
				Filesystem: []*watcher.FilesystemConfig{
					{Name: "fs", WorkflowType: ""},
				},
				Minio: []*watcher.MinioConfig{
					{Name: "minio", WorkflowType: ""},
				},
				Embedded: &watcher.MinioConfig{WorkflowType: ""},
			},
			wantConfig: watcher.Config{
				Filesystem: []*watcher.FilesystemConfig{
					{
						Name:                 "fs",
						WorkflowType:         enums.WorkflowTypeCreateAip,
						RetentionPeriod:      new(time.Duration),
						DeleteFailedSIPAfter: new(time.Duration),
					},
				},
				Minio: []*watcher.MinioConfig{
					{
						Name:                 "minio",
						WorkflowType:         enums.WorkflowTypeCreateAip,
						RetentionPeriod:      new(time.Duration),
						DeleteFailedSIPAfter: new(time.Duration),
					},
				},
				Embedded: &watcher.MinioConfig{
					WorkflowType:         enums.WorkflowTypeCreateAip,
					RetentionPeriod:      new(time.Duration),
					DeleteFailedSIPAfter: new(time.Duration),
				},
			},
		},
		{
			name: "Validates",
			config: watcher.Config{
				Filesystem: []*watcher.FilesystemConfig{
					{
						Name:                 "fs",
						WorkflowType:         "create aip",
						RetentionPeriod:      new(time.Hour),
						DeleteFailedSIPAfter: new(24 * time.Hour),
					},
				},
				Minio: []*watcher.MinioConfig{
					{
						Name:                 "minio",
						WorkflowType:         enums.WorkflowTypeCreateAndReviewAip,
						RetentionPeriod:      new(time.Duration),
						DeleteFailedSIPAfter: new(time.Duration),
					},
				},
				Embedded: &watcher.MinioConfig{
					WorkflowType:         enums.WorkflowTypeCreateAip,
					RetentionPeriod:      new(-1 * time.Second),
					DeleteFailedSIPAfter: new(-1 * time.Second),
				},
			},
			wantConfig: watcher.Config{
				Filesystem: []*watcher.FilesystemConfig{
					{
						Name:                 "fs",
						WorkflowType:         enums.WorkflowTypeCreateAip,
						RetentionPeriod:      new(time.Hour),
						DeleteFailedSIPAfter: new(24 * time.Hour),
					},
				},
				Minio: []*watcher.MinioConfig{
					{
						Name:                 "minio",
						WorkflowType:         enums.WorkflowTypeCreateAndReviewAip,
						RetentionPeriod:      new(time.Duration),
						DeleteFailedSIPAfter: new(time.Duration),
					},
				},
				Embedded: &watcher.MinioConfig{
					WorkflowType:         enums.WorkflowTypeCreateAip,
					RetentionPeriod:      new(-1 * time.Second),
					DeleteFailedSIPAfter: new(-1 * time.Second),
				},
			},
		},
		{
			name: "Invalidates",
			config: watcher.Config{
				Filesystem: []*watcher.FilesystemConfig{
					{Name: "fs1", WorkflowType: "invalid"},
				},
				Minio: []*watcher.MinioConfig{
					{Name: "minio1", WorkflowType: "invalid"},
				},
				Embedded: &watcher.MinioConfig{WorkflowType: "invalid"},
			},
			wantErr: "invalid workflowType in [watcher.embedded] config: \"invalid\"\ninvalid workflowType in [watcher.filesystem] config: \"invalid\"\ninvalid workflowType in [watcher.minio] config: \"invalid\"",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.config.Validate()
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, tc.config, tc.wantConfig)
		})
	}
}
