package db

type Config struct {
	// Driver specifies the database driver (e.g. "mysql" or "sqlite3").
	Driver string

	// DSN (Data Source Name) specifies the database connection information.
	DSN string

	// Migrate specifies whether to run migrations (true) to upgrade the
	// database schema or not (false).
	Migrate bool
}
