package am

import (
	"time"

	"github.com/artefactual-sdps/enduro/internal/sftp"
)

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

	// TransferSourcePath is the path to an Archivematica transfer source
	// directory. It is used in the POST /api/v2beta/package "path" parameter
	// to start a transfer via the API. TransferSourcePath must be prefixed with
	// the UUID of an AMSS transfer source directory, optionally followed by a
	// relative path from the source dir (e.g.
	// "749ef452-fbed-4d50-9072-5f98bc01e52e:sftp_upload").
	TransferSourcePath string

	// Capacity sets the maximum number of worker sessions the worker can
	// handle at one time (default: 1).
	Capacity int

	// PollInterval is the time to wait between poll requests to the AM API.
	PollInterval time.Duration

	// TransferDeadline is the maximum time to wait for a transfer to complete.
	// Set to zero for no deadline.
	TransferDeadline time.Duration
}
