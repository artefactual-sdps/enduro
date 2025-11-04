package storage

import (
	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/event"
)

type Config struct {
	TaskQueue                  string
	EnduroAddress              string
	DefaultPermanentLocationID uuid.UUID
	Internal                   LocationConfig
	Database                   Database
	Event                      event.Config
	AIPDeletion                AIPDeletionConfig
}

type Database struct {
	// Driver specifies the database driver (e.g. "mysql" or "sqlite3").
	Driver string

	// DSN (Data Source Name) specifies the database connection information.
	DSN string

	// Migrate specifies whether to run migrations (true) to upgrade the
	// database schema or not (false).
	Migrate bool
}

type LocationConfig struct {
	// URL specifies the location's driver and address by URL (e.g.
	// "s3://my-bucket?region=us-west-1", "file:///tmp/my-bucket").
	URL string

	// S3 compatible location configuration. If URL has a value then these
	// fields are ignored.
	Name      string
	Region    string
	Endpoint  string
	PathStyle bool
	Profile   string
	Key       string
	Secret    string
	Token     string
	Bucket    string
}

type AIPDeletionConfig struct {
	// ApproveAMSS determines whether AIP deletions are automatically approved in the
	// Archivematica's Storage Service for AMSS locations. When set to false (default),
	// AIP deletions in AMSS locations require manual approval in AMSS. When set to true,
	// they are automatically approved by Enduro, this requires AMSS v0.25.0 or later.
	ApproveAMSS bool
}
