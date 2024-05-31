package package_

import (
	"time"

	"go.artefactual.dev/tools/ref"

	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/db"
)

// preservationActionToGoa returns the API representation of a preservation task.
func preservationActionToGoa(pt *datatypes.PreservationAction) *goapackage.EnduroPackagePreservationAction {
	var startedAt string
	if pt.StartedAt.Valid {
		startedAt = pt.StartedAt.Time.Format(time.RFC3339)
	}

	return &goapackage.EnduroPackagePreservationAction{
		ID:          pt.ID,
		WorkflowID:  pt.WorkflowID,
		Type:        pt.Type.String(),
		Status:      pt.Status.String(),
		StartedAt:   startedAt,
		CompletedAt: db.FormatOptionalTime(pt.CompletedAt),
		PackageID:   &pt.PackageID,
	}
}

// preservationTaskToGoa returns the API representation of a preservation task.
func preservationTaskToGoa(pt *datatypes.PreservationTask) *goapackage.EnduroPackagePreservationTask {
	return &goapackage.EnduroPackagePreservationTask{
		ID:     pt.ID,
		TaskID: pt.TaskID,
		Name:   pt.Name,
		Status: pt.Status.String(),

		// TODO: Make Goa StartedAt a pointer to a string to avoid having to
		// convert a null time to an empty (zero value) string.
		StartedAt: ref.DerefZero(db.FormatOptionalTime(pt.CompletedAt)),

		CompletedAt:          db.FormatOptionalTime(pt.CompletedAt),
		Note:                 &pt.Note,
		PreservationActionID: &pt.PreservationActionID,
	}
}
