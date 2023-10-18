package am

import "github.com/artefactual-sdps/enduro/internal/sftp"

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

	// SFTP configuration for uploading transfers to Archivematica.
	SFTP sftp.Config
}
