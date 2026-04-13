package storage

import (
	"go.artefactual.dev/tools/bucket"

	"github.com/artefactual-sdps/enduro/internal/event"
)

type Config struct {
	TaskQueue   string
	Internal    bucket.Config
	Database    Database
	Event       event.Config
	AIPDeletion AIPDeletionConfig
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

type AIPDeletionConfig struct {
	// ApproveAMSS determines whether AIP deletions are automatically approved in the
	// Archivematica's Storage Service for AMSS locations. When set to false (default),
	// AIP deletions in AMSS locations require manual approval in AMSS. When set to true,
	// they are automatically approved by Enduro, this requires AMSS v0.25.0 or later.
	ApproveAMSS bool

	// ReportTemplatePath specifies the path to the template file used to
	// generate AIP deletion reports. If ReportTemplatePath is empty, AIP
	// deletion reports will not be generated.
	ReportTemplatePath string
}
