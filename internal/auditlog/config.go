package auditlog

type Config struct {
	// Filepath specifies the location of the audit log file. If Filepath is
	// not set, audit logging will be disabled.
	Filepath string

	// MaxSize sets the maximum size of the audit log file in megabytes before
	// it is rotated (default: 100 MB).
	MaxSize int

	// Compress determines if the rotated log files are compressed using gzip
	// (default: false).
	Compress bool
}
