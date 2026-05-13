package activities_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/mockutil"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	ingest_fake "github.com/artefactual-sdps/enduro/internal/ingest/fake"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

func TestCheckSIPChecksumActivity(t *testing.T) {
	t.Parallel()

	sipID := uuid.New()       // Current SIP ID.
	duplicateID := uuid.New() // Duplicate SIP ID.
	hash := "5d96420c54e8c2664a7142ed1d681db98861263a5c248077ba46423ad1f66d35"

	type test struct {
		name      string
		params    activities.CheckDuplicateSIPActivityParams
		mockCalls func(m *ingest_fake.MockServiceMockRecorder)
		want      *activities.CheckDuplicateSIPActivityResult
		wantErr   string
	}
	for _, tt := range []test{
		{
			name: "Checksum matches ingested SIP",
			params: activities.CheckDuplicateSIPActivityParams{
				SIPID: sipID,
				Checksum: datatypes.Checksum{
					Algorithm: datatypes.ChecksumAlgoSHA256,
					Hash:      hash,
				},
			},
			mockCalls: func(m *ingest_fake.MockServiceMockRecorder) {
				m.FindDuplicateSIP(
					mockutil.Context(),
					sipID,
					datatypes.Checksum{
						Algorithm: datatypes.ChecksumAlgoSHA256,
						Hash:      hash,
					},
				).Return(
					&datatypes.SIP{
						UUID:              duplicateID,
						Status:            enums.SIPStatusIngested,
						ChecksumAlgorithm: string(datatypes.ChecksumAlgoSHA256),
						ChecksumHash:      hash,
					},
					nil,
				)
			},
			want: &activities.CheckDuplicateSIPActivityResult{
				Duplicate: &datatypes.SIP{
					UUID:              duplicateID,
					Status:            enums.SIPStatusIngested,
					ChecksumAlgorithm: string(datatypes.ChecksumAlgoSHA256),
					ChecksumHash:      hash,
				},
			},
		},
		{
			name: "Checksum not found",
			params: activities.CheckDuplicateSIPActivityParams{
				SIPID: sipID,
				Checksum: datatypes.Checksum{
					Algorithm: datatypes.ChecksumAlgoSHA256,
					Hash:      hash,
				},
			},
			mockCalls: func(m *ingest_fake.MockServiceMockRecorder) {
				m.FindDuplicateSIP(
					mockutil.Context(),
					sipID,
					datatypes.Checksum{
						Algorithm: datatypes.ChecksumAlgoSHA256,
						Hash:      hash,
					},
				).Return(nil, nil)
			},
			want: &activities.CheckDuplicateSIPActivityResult{},
		},
		{
			name: "Errors when ingest service returns an error",
			params: activities.CheckDuplicateSIPActivityParams{
				SIPID: sipID,
				Checksum: datatypes.Checksum{
					Algorithm: datatypes.ChecksumAlgoSHA256,
					Hash:      hash,
				},
			},
			mockCalls: func(m *ingest_fake.MockServiceMockRecorder) {
				m.FindDuplicateSIP(
					mockutil.Context(),
					sipID,
					datatypes.Checksum{
						Algorithm: datatypes.ChecksumAlgoSHA256,
						Hash:      hash,
					},
				).Return(nil, errors.New("an error occurred"))
			},
			wantErr: "check duplicate SIP: an error occurred",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockSvc := ingest_fake.NewMockService(ctrl)
			tt.mockCalls(mockSvc.EXPECT())

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				activities.NewCheckDuplicateSIPActivity(mockSvc).Execute,
				temporalsdk_activity.RegisterOptions{Name: activities.CheckDuplicateSIPActivityName},
			)

			enc, err := env.ExecuteActivity(
				activities.CheckDuplicateSIPActivityName,
				tt.params,
			)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			var got *activities.CheckDuplicateSIPActivityResult
			err = enc.Get(&got)

			assert.NilError(t, err)
			assert.DeepEqual(t, tt.want, got)
		})
	}
}
