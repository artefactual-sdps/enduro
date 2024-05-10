package storage

import "github.com/google/uuid"

type Config struct {
	TaskQueue                  string
	EnduroAddress              string
	DefaultPermanentLocationID uuid.UUID
	Internal                   LocationConfig
	Database                   Database
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
