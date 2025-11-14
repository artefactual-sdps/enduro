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

// Batch represents a Batch.
type Batch struct {
	ID         int
	UUID       uuid.UUID
	Identifier string
	Status     enums.BatchStatus
	SIPSCount  int

	// It defaults to CURRENT_TIMESTAMP(6) so populated as soon as possible.
	CreatedAt time.Time

	// Nullable, populated as soon as processing starts.
	StartedAt sql.NullTime

	// Nullable, populated as soon as ingest completes.
	CompletedAt sql.NullTime

	// Uploader is the user that uploaded the Batch.
	Uploader *User
}

// Goa returns the API representation of the Batch.
func (b *Batch) Goa() *goaingest.Batch {
	if b == nil {
		return nil
	}

	col := goaingest.Batch{
		UUID:        b.UUID,
		Identifier:  b.Identifier,
		Status:      b.Status.String(),
		SipsCount:   b.SIPSCount,
		CreatedAt:   db.FormatTime(b.CreatedAt),
		StartedAt:   db.FormatOptionalTime(b.StartedAt),
		CompletedAt: db.FormatOptionalTime(b.CompletedAt),
	}
	if b.Uploader != nil {
		col.UploaderUUID = ref.New(b.Uploader.UUID)
		if b.Uploader.Email != "" {
			col.UploaderEmail = &b.Uploader.Email
		}
		if b.Uploader.Name != "" {
			col.UploaderName = &b.Uploader.Name
		}
	}

	return &col
}
