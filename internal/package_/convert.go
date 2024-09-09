package package_

import (
	"time"

	"go.artefactual.dev/tools/ref"

	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/db"
)

func packageToGoaPackageCreatedEvent(p *datatypes.Package) *goapackage.PackageCreatedEvent {
	var id uint
	if p.ID > 0 {
		id = uint(p.ID) // #nosec G115 -- range validated.
	}

	return &goapackage.PackageCreatedEvent{
		ID:   id,
		Item: p.Goa(),
	}
}

// preservationActionToGoa returns the API representation of a preservation task.
func preservationActionToGoa(pa *datatypes.PreservationAction) *goapackage.EnduroPackagePreservationAction {
	var startedAt string
	if pa.StartedAt.Valid {
		startedAt = pa.StartedAt.Time.Format(time.RFC3339)
	}

	var id uint
	if pa.ID > 0 {
		id = uint(pa.ID) // #nosec G115 -- range validated.
	}

	var packageID uint
	if pa.PackageID > 0 {
		packageID = uint(pa.PackageID) // #nosec G115 -- range validated.
	}

	return &goapackage.EnduroPackagePreservationAction{
		ID:          uint(id),
		WorkflowID:  pa.WorkflowID,
		Type:        pa.Type.String(),
		Status:      pa.Status.String(),
		StartedAt:   startedAt,
		CompletedAt: db.FormatOptionalTime(pa.CompletedAt),
		PackageID:   ref.New(packageID),
	}
}

// preservationTaskToGoa returns the API representation of a preservation task.
func preservationTaskToGoa(pt *datatypes.PreservationTask) *goapackage.EnduroPackagePreservationTask {
	var id uint
	if pt.ID > 0 {
		id = uint(pt.ID) // #nosec G115 -- range validated.
	}

	var paID uint
	if pt.PreservationActionID > 0 {
		paID = uint(pt.PreservationActionID) // #nosec G115 -- range validated.
	}

	return &goapackage.EnduroPackagePreservationTask{
		ID:     id,
		TaskID: pt.TaskID,
		Name:   pt.Name,
		Status: pt.Status.String(),

		// TODO: Make Goa StartedAt a pointer to a string to avoid having to
		// convert a null time to an empty (zero value) string.
		StartedAt: ref.DerefZero(db.FormatOptionalTime(pt.CompletedAt)),

		CompletedAt:          db.FormatOptionalTime(pt.CompletedAt),
		Note:                 &pt.Note,
		PreservationActionID: ref.New(paID),
	}
}
