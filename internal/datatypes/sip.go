package datatypes

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"

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

	// Nullable, populated as soon as processing starts.
	StartedAt sql.NullTime

	// Nullable, populated as soon as ingest completes.
	CompletedAt sql.NullTime

	// Set if there is a failure in workflow, it can be empty.
	FailedAs enums.SIPFailedAs

	// Object key from the failed SIP/PIP in the internal bucket.
	FailedKey string

	// Uploader is the user that uploaded the SIP.
	Uploader *User

	// Batch is the batch this SIP belongs to.
	Batch *Batch
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
		StartedAt:   db.FormatOptionalTime(s.StartedAt),
		CompletedAt: db.FormatOptionalTime(s.CompletedAt),
	}
	if s.AIPID.Valid {
		col.AipUUID = ref.New(s.AIPID.UUID.String())
	}
	if s.FailedAs != "" {
		col.FailedAs = ref.New(s.FailedAs.String())
	}
	if s.FailedKey != "" {
		col.FailedKey = ref.New(s.FailedKey)
	}
	if s.Uploader != nil {
		col.UploaderUUID = ref.New(s.Uploader.UUID)
		if s.Uploader.Email != "" {
			col.UploaderEmail = &s.Uploader.Email
		}
		if s.Uploader.Name != "" {
			col.UploaderName = &s.Uploader.Name
		}
	}
	if s.Batch != nil {
		col.BatchUUID = ref.New(s.Batch.UUID)
		col.BatchIdentifier = ref.New(s.Batch.Identifier)
		col.BatchStatus = ref.New(s.Batch.Status.String())
	}

	return &col
}
