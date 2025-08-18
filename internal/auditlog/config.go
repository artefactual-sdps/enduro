package auditlog

type Config struct {
	// Filepath specifies the location of the audit log file.  If Filepath is
	// not set, audit logging will be disabled.
	Filepath string

	// Verbosity sets the minimum log level that will be logged (Default: INFO).
	Verbosity int
}
