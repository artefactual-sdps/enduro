package activities_test

import (
	"errors"
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
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/enums"
	ingestfake "github.com/artefactual-sdps/enduro/internal/ingest/fake"
	storage_enums "github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

func TestClearIngestedSIPsActivity(t *testing.T) {
	t.Parallel()

	batchUUID := uuid.New()
	sipUUID1 := uuid.New()
	sipUUID2 := uuid.New()
	sipUUID3 := uuid.New()
	sipUUID4 := uuid.New()
	aipID1 := uuid.New().String()
	aipID2 := uuid.New().String()
	aipID3 := uuid.New().String()
	aipID4 := uuid.New().String()

	listPayload := func(offset int) *goaingest.ListSipsPayload {
		return &goaingest.ListSipsPayload{
			BatchUUID: ref.New(batchUUID.String()),
			Limit:     ref.New(1000),
			Offset:    ref.New(offset),
		}
	}

	aipDeletionPayload := func(aipID string) *goastorage.AipDeletionAutoPayload {
		return &goastorage.AipDeletionAutoPayload{
			UUID:       aipID,
			Reason:     fmt.Sprintf("Batch %s canceled", batchUUID),
			SkipReport: ref.New(true),
		}
	}

	type test struct {
		name    string
		params  *activities.ClearIngestedSIPsActivityParams
		mock    func(*ingestfake.MockServiceMockRecorder, *ingestfake.MockStorageClientMockRecorder)
		wantErr []string
	}
	for _, tt := range []test{
		{
			name:   "Clears ingested and canceled SIPs with pagination",
			params: &activities.ClearIngestedSIPsActivityParams{BatchUUID: batchUUID},
			mock: func(i *ingestfake.MockServiceMockRecorder, s *ingestfake.MockStorageClientMockRecorder) {
				i.ListSips(mockutil.Context(), listPayload(0)).
					Return(&goaingest.SIPs{
						Items: []*goaingest.SIP{
							{
								UUID:    sipUUID1,
								Status:  enums.SIPStatusIngested.String(),
								AipUUID: ref.New(aipID1),
							},
							{
								UUID:   sipUUID2,
								Status: enums.SIPStatusProcessing.String(),
							},
						},
						Page: &goaingest.EnduroPage{Total: 3},
					}, nil)
				i.SetStatus(mockutil.Context(), sipUUID1, enums.SIPStatusCanceled).Return(nil)
				s.AipDeletionAuto(mockutil.Context(), aipDeletionPayload(aipID1)).Return(nil)

				i.ListSips(mockutil.Context(), listPayload(2)).
					Return(&goaingest.SIPs{
						Items: []*goaingest.SIP{
							{
								UUID:    sipUUID3,
								Status:  enums.SIPStatusCanceled.String(),
								AipUUID: ref.New(aipID2),
							},
						},
						Page: &goaingest.EnduroPage{Total: 3},
					}, nil)
				s.AipDeletionAuto(mockutil.Context(), aipDeletionPayload(aipID2)).Return(nil)

				s.ShowAip(mockutil.Context(), &goastorage.ShowAipPayload{UUID: aipID1}).
					Return(&goastorage.AIP{Status: storage_enums.AIPStatusProcessing.String()}, nil)
				s.ShowAip(mockutil.Context(), &goastorage.ShowAipPayload{UUID: aipID1}).
					Return(&goastorage.AIP{Status: storage_enums.AIPStatusDeleted.String()}, nil)
				s.ShowAip(mockutil.Context(), &goastorage.ShowAipPayload{UUID: aipID2}).
					Return(&goastorage.AIP{Status: storage_enums.AIPStatusDeleted.String()}, nil)
			},
		},
		{
			name:   "Returns an error when listing SIPs fails",
			params: &activities.ClearIngestedSIPsActivityParams{BatchUUID: batchUUID},
			mock: func(i *ingestfake.MockServiceMockRecorder, _ *ingestfake.MockStorageClientMockRecorder) {
				i.ListSips(mockutil.Context(), listPayload(0)).Return(nil, errors.New("persistence error"))
			},
			wantErr: []string{"list SIPs: persistence error"},
		},
		{
			name:   "Continues when deletion request fails for an already deleted AIP",
			params: &activities.ClearIngestedSIPsActivityParams{BatchUUID: batchUUID},
			mock: func(i *ingestfake.MockServiceMockRecorder, s *ingestfake.MockStorageClientMockRecorder) {
				i.ListSips(mockutil.Context(), listPayload(0)).
					Return(&goaingest.SIPs{
						Items: []*goaingest.SIP{
							{
								UUID:    sipUUID1,
								Status:  enums.SIPStatusCanceled.String(),
								AipUUID: ref.New(aipID1),
							},
						},
						Page: &goaingest.EnduroPage{Total: 1},
					}, nil)
				s.AipDeletionAuto(mockutil.Context(), aipDeletionPayload(aipID1)).
					Return(errors.New("deletion already requested"))
				s.ShowAip(mockutil.Context(), &goastorage.ShowAipPayload{UUID: aipID1}).
					Return(&goastorage.AIP{Status: storage_enums.AIPStatusDeleted.String()}, nil).
					Times(2)
			},
		},
		{
			name:   "Aggregates parse, status update, and deletion wait errors",
			params: &activities.ClearIngestedSIPsActivityParams{BatchUUID: batchUUID},
			mock: func(i *ingestfake.MockServiceMockRecorder, s *ingestfake.MockStorageClientMockRecorder) {
				i.ListSips(mockutil.Context(), listPayload(0)).
					Return(&goaingest.SIPs{
						Items: []*goaingest.SIP{
							{
								UUID:   sipUUID1,
								Status: "invalid",
							},
							{
								UUID:    sipUUID2,
								Status:  enums.SIPStatusIngested.String(),
								AipUUID: ref.New(aipID1),
							},
							{
								UUID:    sipUUID3,
								Status:  enums.SIPStatusCanceled.String(),
								AipUUID: ref.New(aipID2),
							},
						},
						Page: &goaingest.EnduroPage{Total: 3},
					}, nil)
				i.SetStatus(mockutil.Context(), sipUUID2, enums.SIPStatusCanceled).
					Return(errors.New("status update failed"))
				s.AipDeletionAuto(mockutil.Context(), aipDeletionPayload(aipID2)).Return(nil)
				s.ShowAip(mockutil.Context(), &goastorage.ShowAipPayload{UUID: aipID2}).
					Return(&goastorage.AIP{Status: storage_enums.AIPStatusStored.String()}, nil)
			},
			wantErr: []string{
				fmt.Sprintf("parse SIP %q status: invalid is not a valid SIPStatus", sipUUID1.String()),
				fmt.Sprintf("set SIP %q status to canceled: status update failed", sipUUID2.String()),
				fmt.Sprintf("AIP %q could not be deleted", aipID2),
			},
		},
		{
			name:   "Aggregates deletion and AIP status check errors",
			params: &activities.ClearIngestedSIPsActivityParams{BatchUUID: batchUUID},
			mock: func(i *ingestfake.MockServiceMockRecorder, s *ingestfake.MockStorageClientMockRecorder) {
				i.ListSips(mockutil.Context(), listPayload(0)).
					Return(&goaingest.SIPs{
						Items: []*goaingest.SIP{
							{
								UUID:    sipUUID1,
								Status:  enums.SIPStatusCanceled.String(),
								AipUUID: ref.New(aipID1),
							},
							{
								UUID:    sipUUID2,
								Status:  enums.SIPStatusCanceled.String(),
								AipUUID: ref.New(aipID2),
							},
							{
								UUID:    sipUUID3,
								Status:  enums.SIPStatusCanceled.String(),
								AipUUID: ref.New(aipID3),
							},
							{
								UUID:    sipUUID4,
								Status:  enums.SIPStatusCanceled.String(),
								AipUUID: ref.New(aipID4),
							},
						},
						Page: &goaingest.EnduroPage{Total: 4},
					}, nil)
				s.AipDeletionAuto(mockutil.Context(), aipDeletionPayload(aipID1)).
					Return(errors.New("deletion request failed"))
				s.ShowAip(mockutil.Context(), &goastorage.ShowAipPayload{UUID: aipID1}).
					Return(nil, errors.New("lookup failed"))
				s.AipDeletionAuto(mockutil.Context(), aipDeletionPayload(aipID2)).
					Return(errors.New("deletion request failed"))
				s.ShowAip(mockutil.Context(), &goastorage.ShowAipPayload{UUID: aipID2}).
					Return(&goastorage.AIP{Status: storage_enums.AIPStatusStored.String()}, nil)
				s.AipDeletionAuto(mockutil.Context(), aipDeletionPayload(aipID3)).Return(nil)
				s.AipDeletionAuto(mockutil.Context(), aipDeletionPayload(aipID4)).Return(nil)
				s.ShowAip(mockutil.Context(), &goastorage.ShowAipPayload{UUID: aipID3}).
					Return(nil, errors.New("status lookup failed"))
				s.ShowAip(mockutil.Context(), &goastorage.ShowAipPayload{UUID: aipID4}).
					Return(&goastorage.AIP{Status: storage_enums.AIPStatusStored.String()}, nil)
			},
			wantErr: []string{
				fmt.Sprintf("request AIP %q deletion:", aipID1),
				"deletion request failed",
				"lookup failed",
				fmt.Sprintf("request AIP %q deletion:", aipID2),
				fmt.Sprintf("show AIP %q: status lookup failed", aipID3),
				fmt.Sprintf("AIP %q could not be deleted", aipID4),
			},
		},
		{
			name:   "Skips canceled SIP without AIP UUID",
			params: &activities.ClearIngestedSIPsActivityParams{BatchUUID: batchUUID},
			mock: func(i *ingestfake.MockServiceMockRecorder, _ *ingestfake.MockStorageClientMockRecorder) {
				i.ListSips(mockutil.Context(), listPayload(0)).
					Return(&goaingest.SIPs{
						Items: []*goaingest.SIP{
							{
								UUID:   sipUUID1,
								Status: enums.SIPStatusCanceled.String(),
							},
						},
						Page: &goaingest.EnduroPage{Total: 1},
					}, nil)
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			ingestsvc := ingestfake.NewMockService(gomock.NewController(t))
			storageClient := ingestfake.NewMockStorageClient(gomock.NewController(t))
			if tt.mock != nil {
				tt.mock(ingestsvc.EXPECT(), storageClient.EXPECT())
			}

			env.RegisterActivityWithOptions(
				activities.NewClearIngestedSIPsActivity(ingestsvc, storageClient, time.Microsecond).Execute,
				temporalsdk_activity.RegisterOptions{Name: activities.ClearIngestedSIPsActivityName},
			)

			enc, err := env.ExecuteActivity(activities.ClearIngestedSIPsActivityName, tt.params)
			if len(tt.wantErr) > 0 {
				assert.Assert(t, err != nil)
				for _, wantErr := range tt.wantErr {
					assert.ErrorContains(t, err, wantErr)
				}
				return
			}
			assert.NilError(t, err)

			var res activities.ClearIngestedSIPsActivityResult
			err = enc.Get(&res)
			assert.NilError(t, err)
			assert.DeepEqual(t, &res, &activities.ClearIngestedSIPsActivityResult{})
		})
	}
}
