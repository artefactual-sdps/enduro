package am

import "github.com/artefactual-sdps/enduro/internal/sftp"

type Config struct {
	// SFTP configuration for uploading transfers to Archivematica.
	SFTP sftp.Config
}
