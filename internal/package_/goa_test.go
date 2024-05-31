package package__test

import (
	"context"
	"errors"
	"testing"

	"go.artefactual.dev/tools/mockutil"
	"go.artefactual.dev/tools/ref"
	"gotest.tools/v3/assert"

	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/package_"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/persistence/fake"
)

func TestGoaCreatePreservationAction(t *testing.T) {
	t.Parallel()

	workflowID := "processing-workflow-720db1d4-825c-4911-9a20-61c212cf23ff"
	startedAt := "2024-06-03T08:51:35Z"
	completedAt := "2024-06-03T08:52:18Z"

	type test struct {
		name    string
		payload goapackage.CreatePreservationActionPayload
		mock    func(*persistence_fake.MockService, *datatypes.PreservationAction) *persistence_fake.MockService
		want    goapackage.EnduroPackagePreservationAction
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Creates a preservation action",
			payload: goapackage.CreatePreservationActionPayload{
				PackageID:   1,
				WorkflowID:  workflowID,
				Type:        enums.PreservationActionTypeCreateAIP.String(),
				Status:      enums.PreservationActionStatusDone.String(),
				StartedAt:   ref.New(startedAt),
				CompletedAt: ref.New(completedAt),
			},
			mock: func(svc *persistence_fake.MockService, pa *datatypes.PreservationAction) *persistence_fake.MockService {
				svc.EXPECT().
					CreatePreservationAction(mockutil.Context(), pa).
					DoAndReturn(
						func(ctx context.Context, pa *datatypes.PreservationAction) error {
							pa.ID = 11
							return nil
						},
					)
				return svc
			},
			want: goapackage.EnduroPackagePreservationAction{
				ID:          11,
				WorkflowID:  workflowID,
				Type:        enums.PreservationActionTypeCreateAIP.String(),
				Status:      enums.PreservationActionStatusDone.String(),
				StartedAt:   startedAt,
				CompletedAt: &completedAt,
				PackageID:   ref.New(uint(1)),
			},
		},
		{
			name: "Errors if PackageID is missing",
			payload: goapackage.CreatePreservationActionPayload{
				WorkflowID:  workflowID,
				Type:        enums.PreservationActionTypeCreateAIP.String(),
				Status:      enums.PreservationActionStatusDone.String(),
				StartedAt:   ref.New(startedAt),
				CompletedAt: ref.New(completedAt),
			},
			mock: func(svc *persistence_fake.MockService, pa *datatypes.PreservationAction) *persistence_fake.MockService {
				svc.EXPECT().
					CreatePreservationAction(mockutil.Context(), pa).
					DoAndReturn(
						func(ctx context.Context, pa *datatypes.PreservationAction) error {
							return errors.New("invalid data error: field \"PackageID\" is required")
						},
					)
				return svc
			},
			wantErr: "preservation action: create: invalid data error: field \"PackageID\" is required",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			payload := tt.payload
			pkgSvc, perSvc := testSvc(t)
			if tt.mock != nil {
				tt.mock(perSvc, package_.GoaToPreservationAction(&payload))
			}

			pa, err := pkgSvc.Goa().CreatePreservationAction(context.Background(), &payload)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, pa, &tt.want) // #nosec G601
		})
	}
}
