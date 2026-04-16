package datatypes

import (
	"time"

	"github.com/google/uuid"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/db"
	"github.com/artefactual-sdps/enduro/internal/enums"
)

// SIP represents a SIP.
type SIP struct {
	ID     int
	UUID   uuid.UUID
	Name   string
	AIPID  uuid.NullUUID // Nullable.
	Status enums.SIPStatus

	// It defaults to CURRENT_TIMESTAMP(6) so populated as soon as possible.
	CreatedAt time.Time

	// Zero until processing starts.
	StartedAt time.Time

	// Zero until ingest completes.
	CompletedAt time.Time

	// Set if there is a failure in workflow, it can be empty.
	FailedAs enums.SIPFailedAs

	// Object key from the failed SIP/PIP in the internal bucket.
	FailedKey string

	// Uploader is the user that uploaded the SIP.
	Uploader *User

	// Batch is the batch this SIP belongs to.
	Batch *Batch

	// FileCount is the number of files in the SIP.
	FileCount int32
}

// Goa returns the API representation of the SIP.
func (s *SIP) Goa() *goaingest.SIP {
	if s == nil {
		return nil
	}

	col := goaingest.SIP{
		UUID:        s.UUID,
		Name:        db.FormatOptionalString(s.Name),
		Status:      s.Status.String(),
		CreatedAt:   db.FormatTime(s.CreatedAt),
		StartedAt:   db.FormatOptionalZeroTime(s.StartedAt),
		CompletedAt: db.FormatOptionalZeroTime(s.CompletedAt),
	}
	if s.AIPID.Valid {
		col.AipUUID = new(s.AIPID.UUID.String())
	}
	if s.FailedAs != "" {
		col.FailedAs = new(s.FailedAs.String())
	}
	if s.FailedKey != "" {
		col.FailedKey = new(s.FailedKey)
	}
	if s.Uploader != nil {
		col.UploaderUUID = new(s.Uploader.UUID)
		if s.Uploader.Email != "" {
			col.UploaderEmail = &s.Uploader.Email
		}
		if s.Uploader.Name != "" {
			col.UploaderName = &s.Uploader.Name
		}
	}
	if s.Batch != nil {
		col.BatchUUID = new(s.Batch.UUID)
		col.BatchIdentifier = new(s.Batch.Identifier)
		col.BatchStatus = new(s.Batch.Status.String())
	}

	return &col
}
