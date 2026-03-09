package datatypes

import (
	"time"

	"github.com/google/uuid"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/db"
	"github.com/artefactual-sdps/enduro/internal/enums"
)

// Batch represents a Batch.
type Batch struct {
	ID          int
	UUID        uuid.UUID
	Identifier  string
	Status      enums.BatchStatus
	SIPSCount   int
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time

	// Uploader is the user that uploaded the Batch.
	Uploader *User
}

// Goa returns the API representation of the Batch.
func (b *Batch) Goa() *goaingest.Batch {
	if b == nil {
		return nil
	}

	col := goaingest.Batch{
		UUID:       b.UUID,
		Identifier: b.Identifier,
		Status:     b.Status.String(),
		SipsCount:  b.SIPSCount,
		CreatedAt:  db.FormatTime(b.CreatedAt),
	}
	if !b.StartedAt.IsZero() {
		col.StartedAt = new(b.StartedAt.Format(time.RFC3339))
	}
	if !b.CompletedAt.IsZero() {
		col.CompletedAt = new(b.CompletedAt.Format(time.RFC3339))
	}
	if b.Uploader != nil {
		col.UploaderUUID = new(b.Uploader.UUID)
		if b.Uploader.Email != "" {
			col.UploaderEmail = &b.Uploader.Email
		}
		if b.Uploader.Name != "" {
			col.UploaderName = &b.Uploader.Name
		}
	}

	return &col
}
