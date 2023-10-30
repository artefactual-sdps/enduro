package am

type Config struct {
	// Archivematica server address.
	Address string

	// Archivematica API user.
	User string

	// Archivematica API key.
	Key string

	// Directory where transfers are deposited for processing (must be readable
	// by Archivematica).
	ShareDir string
}
