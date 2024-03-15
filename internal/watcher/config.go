package watcher

import (
	"path/filepath"
	"time"
)

const defaultPollInterval = 200 * time.Millisecond

type Config struct {
	Filesystem []*FilesystemConfig
	Minio      []*MinioConfig
	Embedded   *MinioConfig
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
}
