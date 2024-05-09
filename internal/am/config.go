package am

import (
	"time"

	"github.com/google/uuid"

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

	// Capacity sets the maximum number of worker sessions the worker can
	// handle at one time (default: 1).
	Capacity int

	// PollInterval is the time to wait between poll requests to the AM API.
	PollInterval time.Duration

	// TransferDeadline is the maximum time to wait for a transfer to complete.
	// Set to zero for no deadline.
	TransferDeadline time.Duration

	// AMSSLocationID is the local UUID of the Archivematica Storage Service
	// storage location.
	AMSSLocationID uuid.UUID
}
