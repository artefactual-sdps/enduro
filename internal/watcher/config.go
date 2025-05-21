package watcher

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/artefactual-sdps/enduro/internal/enums"
)

const defaultPollInterval = 200 * time.Millisecond

type Config struct {
	Filesystem []*FilesystemConfig
	Minio      []*MinioConfig
	Embedded   *MinioConfig
}

func (c Config) setDefaults() {
	if c.Embedded != nil && c.Embedded.WorkflowType == "" {
		c.Embedded.WorkflowType = enums.WorkflowTypeCreateAip
	}
	for _, fs := range c.Filesystem {
		if fs != nil && fs.WorkflowType == "" {
			fs.WorkflowType = enums.WorkflowTypeCreateAip
		}
	}
	for _, minio := range c.Minio {
		if minio != nil && minio.WorkflowType == "" {
			minio.WorkflowType = enums.WorkflowTypeCreateAip
		}
	}
}

func (c Config) Validate() error {
	c.setDefaults()

	var err error
	if c.Embedded != nil && !c.Embedded.WorkflowType.IsValid() {
		err = fmt.Errorf("invalid workflowType in [watcher.embedded] config: %q", c.Embedded.WorkflowType)
	}
	for _, fs := range c.Filesystem {
		if fs != nil && !fs.WorkflowType.IsValid() {
			err = errors.Join(
				err,
				fmt.Errorf("invalid workflowType in [watcher.filesystem] config: %q", fs.WorkflowType),
			)
		}
	}
	for _, minio := range c.Minio {
		if minio != nil && !minio.WorkflowType.IsValid() {
			err = errors.Join(
				err,
				fmt.Errorf("invalid workflowType in [watcher.minio] config: %q", minio.WorkflowType),
			)
		}
	}

	return err
}

func (c Config) CompletedDirs() []string {
	dirs := []string{}
	for _, item := range c.Filesystem {
		if item == nil {
			continue
		}
		if item.CompletedDir == "" {
			continue
		}
		if abs, err := filepath.Abs(item.CompletedDir); err == nil {
			dirs = append(dirs, abs)
		}
	}
	return dirs
}

// See filesystem.go for more.
type FilesystemConfig struct {
	Name         string
	Path         string
	CompletedDir string
	Ignore       string
	Inotify      bool

	RetentionPeriod  *time.Duration
	StripTopLevelDir bool

	// PollInterval sets the length of time between filesystem polls (default:
	// 200ms). If Inotify is true then PollInterval is ignored.
	PollInterval time.Duration

	// WorkflowType specifies which workflow this watcher should execute
	// (default: "create aip").
	WorkflowType enums.WorkflowType
}

func (cfg *FilesystemConfig) setDefaults() {
	if cfg.PollInterval == 0 {
		cfg.PollInterval = defaultPollInterval
	}
}

// See minio.go for more.
type MinioConfig struct {
	Name            string
	RedisAddress    string
	RedisList       string
	RedisFailedList string
	RedisPopTimeout time.Duration
	Region          string
	Endpoint        string
	PathStyle       bool
	Profile         string
	Key             string
	Secret          string
	Token           string
	Bucket          string
	URL             string

	RetentionPeriod  *time.Duration
	StripTopLevelDir bool

	// PollInterval sets the length of time between Redis polls (default: 1s).
	PollInterval time.Duration

	// WorkflowType specifies which workflow this watcher should execute
	// (default: "create aip").
	WorkflowType enums.WorkflowType
}
