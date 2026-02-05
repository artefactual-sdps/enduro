package activities_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/mockutil"
	"go.artefactual.dev/tools/ref"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/entfilter"
	"github.com/artefactual-sdps/enduro/internal/enums"
	ingestfake "github.com/artefactual-sdps/enduro/internal/ingest/fake"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

func TestPollSIPStatusesActivity(t *testing.T) {
	t.Parallel()

	batchUUID := uuid.New()
	sip1UUID := uuid.New()
	sip2UUID := uuid.New()
	aip1UUID := uuid.New()
	aip2UUID := uuid.New()

	payload := &goaingest.ListSipsPayload{
		BatchUUID: ref.New(batchUUID.String()),
		Limit:     ref.New(entfilter.MaxPageSize),
	}

	type test struct {
		name    string
		params  *activities.PollSIPStatusesActivityParams
		mock    func(*ingestfake.MockServiceMockRecorder)
		want    *activities.PollSIPStatusesActivityResult
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Returns true when all SIPs have expected status",
			params: &activities.PollSIPStatusesActivityParams{
				BatchUUID:        batchUUID,
				ExpectedSIPCount: 2,
				ExpectedStatus:   enums.SIPStatusValidated,
			},
			mock: func(r *ingestfake.MockServiceMockRecorder) {
				r.ListSips(mockutil.Context(), payload).
					Return(&goaingest.SIPs{Items: []*goaingest.SIP{
						{UUID: sip1UUID, Status: enums.SIPStatusValidated.String()},
						{UUID: sip2UUID, Status: enums.SIPStatusValidated.String()},
					}}, nil)
			},
			want: &activities.PollSIPStatusesActivityResult{AllExpectedStatus: true},
		},
		{
			name: "Returns true when all SIPs have ingested status",
			params: &activities.PollSIPStatusesActivityParams{
				BatchUUID:        batchUUID,
				ExpectedSIPCount: 2,
				ExpectedStatus:   enums.SIPStatusIngested,
			},
			mock: func(r *ingestfake.MockServiceMockRecorder) {
				r.ListSips(mockutil.Context(), payload).
					Return(&goaingest.SIPs{Items: []*goaingest.SIP{
						{
							UUID:    sip1UUID,
							Status:  enums.SIPStatusIngested.String(),
							AipUUID: ref.New(aip1UUID.String()),
						},
						{
							UUID:    sip2UUID,
							Status:  enums.SIPStatusIngested.String(),
							AipUUID: ref.New(aip2UUID.String()),
						},
					}}, nil)
			},
			want: &activities.PollSIPStatusesActivityResult{
				AllExpectedStatus: true,
				SIPIDstoAIPIDs: map[uuid.UUID]uuid.UUID{
					sip1UUID: aip1UUID,
					sip2UUID: aip2UUID,
				},
			},
		},
		{
			name: "Returns false when some SIPs have failed status",
			params: &activities.PollSIPStatusesActivityParams{
				BatchUUID:        batchUUID,
				ExpectedSIPCount: 2,
				ExpectedStatus:   enums.SIPStatusValidated,
			},
			mock: func(r *ingestfake.MockServiceMockRecorder) {
				r.ListSips(mockutil.Context(), payload).
					Return(&goaingest.SIPs{Items: []*goaingest.SIP{
						{UUID: sip1UUID, Status: enums.SIPStatusValidated.String()},
						{UUID: sip2UUID, Status: enums.SIPStatusFailed.String()},
					}}, nil)
			},
			want: &activities.PollSIPStatusesActivityResult{AllExpectedStatus: false},
		},
		{
			name: "Returns false when some SIPs have error status",
			params: &activities.PollSIPStatusesActivityParams{
				BatchUUID:        batchUUID,
				ExpectedSIPCount: 2,
				ExpectedStatus:   enums.SIPStatusValidated,
			},
			mock: func(r *ingestfake.MockServiceMockRecorder) {
				r.ListSips(mockutil.Context(), payload).
					Return(&goaingest.SIPs{Items: []*goaingest.SIP{
						{UUID: sip1UUID, Status: enums.SIPStatusValidated.String()},
						{UUID: sip2UUID, Status: enums.SIPStatusError.String()},
					}}, nil)
			},
			want: &activities.PollSIPStatusesActivityResult{AllExpectedStatus: false},
		},
		{
			name: "Returns false when some SIPs have canceled status",
			params: &activities.PollSIPStatusesActivityParams{
				BatchUUID:        batchUUID,
				ExpectedSIPCount: 2,
				ExpectedStatus:   enums.SIPStatusValidated,
			},
			mock: func(r *ingestfake.MockServiceMockRecorder) {
				r.ListSips(mockutil.Context(), payload).
					Return(&goaingest.SIPs{Items: []*goaingest.SIP{
						{UUID: sip1UUID, Status: enums.SIPStatusCanceled.String()},
						{UUID: sip2UUID, Status: enums.SIPStatusValidated.String()},
					}}, nil)
			},
			want: &activities.PollSIPStatusesActivityResult{AllExpectedStatus: false},
		},
		{
			name: "Continues polling when SIPs are not in a final status",
			params: &activities.PollSIPStatusesActivityParams{
				BatchUUID:        batchUUID,
				ExpectedSIPCount: 2,
				ExpectedStatus:   enums.SIPStatusValidated,
			},
			mock: func(r *ingestfake.MockServiceMockRecorder) {
				// First call: SIPs in progress.
				r.ListSips(mockutil.Context(), payload).
					Return(&goaingest.SIPs{Items: []*goaingest.SIP{
						{UUID: sip1UUID, Status: enums.SIPStatusProcessing.String()},
						{UUID: sip2UUID, Status: enums.SIPStatusQueued.String()},
					}}, nil)
				// Second call: one SIP validated and the other in progress.
				r.ListSips(mockutil.Context(), payload).
					Return(&goaingest.SIPs{Items: []*goaingest.SIP{
						{UUID: sip1UUID, Status: enums.SIPStatusValidated.String()},
						{UUID: sip2UUID, Status: enums.SIPStatusProcessing.String()},
					}}, nil)
				// Third call: SIPs validated.
				r.ListSips(mockutil.Context(), payload).
					Return(&goaingest.SIPs{Items: []*goaingest.SIP{
						{UUID: sip1UUID, Status: enums.SIPStatusValidated.String()},
						{UUID: sip2UUID, Status: enums.SIPStatusValidated.String()},
					}}, nil)
			},
			want: &activities.PollSIPStatusesActivityResult{AllExpectedStatus: true},
		},
		{
			name: "Fails when SIP count doesn't match expected",
			params: &activities.PollSIPStatusesActivityParams{
				BatchUUID:        batchUUID,
				ExpectedSIPCount: 3,
				ExpectedStatus:   enums.SIPStatusValidated,
			},
			mock: func(r *ingestfake.MockServiceMockRecorder) {
				r.ListSips(mockutil.Context(), payload).
					Return(&goaingest.SIPs{Items: []*goaingest.SIP{
						{UUID: sip1UUID, Status: enums.SIPStatusValidated.String()},
						{UUID: sip2UUID, Status: enums.SIPStatusValidated.String()},
					}}, nil)
			},
			wantErr: "check SIP statuses: expected 3 SIPs but found 2",
		},
		{
			name: "Fails when ListSips returns an error",
			params: &activities.PollSIPStatusesActivityParams{
				BatchUUID:        batchUUID,
				ExpectedSIPCount: 2,
				ExpectedStatus:   enums.SIPStatusValidated,
			},
			mock: func(r *ingestfake.MockServiceMockRecorder) {
				r.ListSips(mockutil.Context(), payload).
					Return(nil, fmt.Errorf("persistence error"))
			},
			wantErr: "check SIP statuses: list SIPs: persistence error",
		},
		{
			name: "Fails on invalid SIP status",
			params: &activities.PollSIPStatusesActivityParams{
				BatchUUID:        batchUUID,
				ExpectedSIPCount: 2,
				ExpectedStatus:   enums.SIPStatusValidated,
			},
			mock: func(r *ingestfake.MockServiceMockRecorder) {
				r.ListSips(mockutil.Context(), payload).
					Return(&goaingest.SIPs{Items: []*goaingest.SIP{
						{UUID: sip1UUID, Status: "unknown"},
						{UUID: sip2UUID, Status: enums.SIPStatusValidated.String()},
					}}, nil)
			},
			wantErr: "check SIP statuses: invalid SIP status: unknown",
		},
		{
			name: "Fails on invalid AIP UUID",
			params: &activities.PollSIPStatusesActivityParams{
				BatchUUID:        batchUUID,
				ExpectedSIPCount: 2,
				ExpectedStatus:   enums.SIPStatusValidated,
			},
			mock: func(r *ingestfake.MockServiceMockRecorder) {
				r.ListSips(mockutil.Context(), payload).
					Return(&goaingest.SIPs{Items: []*goaingest.SIP{
						{
							UUID:    sip1UUID,
							Status:  enums.SIPStatusIngested.String(),
							AipUUID: ref.New("invalid-uuid"),
						},
						{UUID: sip2UUID, Status: enums.SIPStatusIngested.String()},
					}}, nil)
			},
			wantErr: "check SIP statuses: parse AIP UUID: invalid UUID length: 12",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			ingestsvc := ingestfake.NewMockService(gomock.NewController(t))
			if tt.mock != nil {
				tt.mock(ingestsvc.EXPECT())
			}

			env.RegisterActivityWithOptions(
				activities.NewPollSIPStatusesActivity(ingestsvc, time.Microsecond).Execute,
				temporalsdk_activity.RegisterOptions{Name: activities.PollSIPStatusesActivityName},
			)

			enc, err := env.ExecuteActivity(activities.PollSIPStatusesActivityName, tt.params)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			var res activities.PollSIPStatusesActivityResult
			err = enc.Get(&res)
			assert.NilError(t, err)
			assert.DeepEqual(t, &res, tt.want)
		})
	}
}
