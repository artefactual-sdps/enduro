package am

import "github.com/artefactual-sdps/enduro/internal/sftp"

type Config struct {
	// Archivematica server address.
	Address string

	// Archivematica API user.
	User string

	// Archivematica API key.
	APIKey string

	// Archivematica processing configuration to use (default: "automated").
	ProcessingConfig string

	// SFTP configuration for uploading transfers to Archivematica.
	SFTP sftp.Config
}
