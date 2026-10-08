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
	if c.Embedded != nil {
		setWorkflowTypeDefault(&c.Embedded.WorkflowType)
		setRetentionDefaults(&c.Embedded.RetentionPeriod, &c.Embedded.DeleteFailedSIPAfter)
	}
	for _, fs := range c.Filesystem {
		if fs != nil {
			setWorkflowTypeDefault(&fs.WorkflowType)
			setRetentionDefaults(&fs.RetentionPeriod, &fs.DeleteFailedSIPAfter)
		}
	}
	for _, minio := range c.Minio {
		if minio != nil {
			setWorkflowTypeDefault(&minio.WorkflowType)
			setRetentionDefaults(&minio.RetentionPeriod, &minio.DeleteFailedSIPAfter)
		}
	}
}

func setWorkflowTypeDefault(t *enums.WorkflowType) {
	if *t == "" {
		*t = enums.WorkflowTypeCreateAip
	}
}

// setRetentionDefaults sets unset (nil) retention periods to zero, so that
// watched location SIPs are deleted immediately after ingest by default. A
// negative value must be configured explicitly to retain SIPs indefinitely.
func setRetentionDefaults(retention, deleteFailed **time.Duration) {
	if *retention == nil {
		*retention = new(time.Duration)
	}
	if *deleteFailed == nil {
		*deleteFailed = new(time.Duration)
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

	// RetentionPeriod is the duration for which SIPs should be retained after
	// a successful ingest. If zero or unset (default), SIPs will be deleted
	// immediately after ingest. If negative, SIPs will be retained
	// indefinitely.
	RetentionPeriod *time.Duration

	// DeleteFailedSIPAfter is the duration to retain SIPs after an unsuccessful
	// ingest. If zero or unset (default), the SIP is deleted immediately.
	// If negative, the SIP is retained indefinitely.
	DeleteFailedSIPAfter *time.Duration

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

	// RetentionPeriod is the duration for which SIPs should be retained after
	// a successful ingest. If zero or unset (default), SIPs will be deleted
	// immediately after ingest. If negative, SIPs will be retained
	// indefinitely.
	RetentionPeriod *time.Duration

	// DeleteFailedSIPAfter is the duration to retain SIPs after an unsuccessful
	// ingest. A SIP is considered unsuccessful if its ingest failed, errored or
	// was canceled. If zero or unset (default), the SIP is deleted immediately.
	// If negative, the SIP is retained indefinitely.
	DeleteFailedSIPAfter *time.Duration

	// WatchTimeout sets the maximum time the Watch() method will wait for a
	// Redis event when the queue is empty before returning a timeout error
	// (default: 1 minute).
	WatchTimeout time.Duration

	// WorkflowType specifies which workflow this watcher should execute
	// (default: "create aip").
	WorkflowType enums.WorkflowType
}
