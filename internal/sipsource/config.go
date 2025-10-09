package sipsource

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/bucket"
)

var (
	// ErrMissingID is returned when the SIP source ID is missing.
	ErrMissingID = errors.New("SIP source: missing ID")
	// ErrMissingName is returned when the SIP source name is empty.
	ErrMissingName = errors.New("SIP source: missing name")
	// ErrMissingBucket is returned when no bucket is configured.
	ErrMissingBucket = errors.New("SIP source: missing bucket")
	// ErrInvalidConfig is returned when the SIP source configuration is invalid.
	ErrInvalidConfig = errors.New("SIP source: invalid configuration")
)

type Config struct {
	// ID is the UUID of the SIP source.
	ID uuid.UUID

	// Name is the human readable name of the SIP source.
	Name string

	// Bucket is the configuration for the bucket to be used as the SIP source.
	Bucket *bucket.Config

	// RetentionPeriod is the duration for which SIPs should be retained after
	// a successful ingest. If negative, SIPs will be retained indefinitely.
	RetentionPeriod time.Duration
}

func (c *Config) Validate() error {
	// Allow empty SIP source configurations, for installations where no SIP
	// source is needed.
	if c.IsEmpty() {
		return nil
	}

	var errs error
	if c.ID == uuid.Nil {
		errs = errors.Join(errs, ErrMissingID)
	}
	if c.Name == "" {
		errs = errors.Join(errs, ErrMissingName)
	}
	if c.Bucket == nil {
		errs = errors.Join(errs, ErrMissingBucket)
	}
	if err := validateBucketConfig(c.Bucket); err != nil {
		errs = errors.Join(errs, err)
	}

	return errs
}

func (c *Config) IsEmpty() bool {
	return c == nil || (c.ID == uuid.Nil && c.Name == "" && c.Bucket == nil)
}

func validateBucketConfig(cfg *bucket.Config) error {
	if cfg == nil {
		return nil
	}
	if cfg.URL == "" && cfg.Endpoint == "" {
		return fmt.Errorf(
			"%w: [sipsource.bucket]: either a URL or S3 style (endpoint) bucket configuration must be provided",
			ErrInvalidConfig,
		)
	}
	if cfg.URL != "" && cfg.Endpoint != "" {
		return fmt.Errorf(
			"%w: [sipsource.bucket]: the URL and S3 style (endpoint) configurations are mutually exclusive",
			ErrInvalidConfig,
		)
	}

	return nil
}
