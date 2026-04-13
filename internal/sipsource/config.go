package sipsource

import (
	"errors"
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

	return errs
}

func (c *Config) IsEmpty() bool {
	return c == nil || (c.ID == uuid.Nil && c.Name == "" && c.Bucket == nil)
}
