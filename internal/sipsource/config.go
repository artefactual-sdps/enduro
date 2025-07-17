package sipsource

import (
	"errors"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/bucket"
)

var (
	// ErrInvalidID is returned when the SIP source ID is invalid.
	ErrInvalidID = errors.New("SIP source: invalid ID")
	// ErrInvalidBucket is returned when a bucket is not configured.
	ErrInvalidBucket = errors.New("SIP source: invalid bucket")
	// ErrInvalidName is returned when the SIP source name is empty.
	ErrInvalidName = errors.New("SIP source: invalid name")
)

type Config struct {
	// ID is the UUID of the SIP source.
	ID uuid.UUID

	// Bucket is the configuration for the bucket where SIPs are stored.
	Bucket *bucket.Config

	// Name is the human readable name of the SIP source.
	Name string
}

func (c *Config) Validate() error {
	var errs error
	if c.ID == uuid.Nil {
		errs = errors.Join(errs, ErrInvalidID)
	}
	if c.Bucket == nil {
		errs = errors.Join(errs, ErrInvalidBucket)
	}
	if c.Name == "" {
		errs = errors.Join(errs, ErrInvalidName)
	}
	return errs
}
