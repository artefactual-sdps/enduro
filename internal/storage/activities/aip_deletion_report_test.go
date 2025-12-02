package activities_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"go.artefactual.dev/tools/mockutil"
	"go.artefactual.dev/tools/ref"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/memblob"
	"gotest.tools/v3/assert"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/activities"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/fake"
	pdf_fake "github.com/artefactual-sdps/enduro/internal/storage/pdfs/fake"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func expectReadAIP(msvc *fake.MockService, id uuid.UUID) {
	msvc.EXPECT().
		ReadAip(mockutil.Context(), id).
		Return(&goastorage.AIP{
			UUID:         id,
			Name:         "Test AIP",
			LocationUUID: ref.New(uuid.MustParse("223e4567-e89b-12d3-a456-426614174000")),
		}, nil)
}

func expectListDeletionRequests(msvc *fake.MockService, aipID uuid.UUID) {
	msvc.EXPECT().
		ListDeletionRequests(mockutil.Context(), &persistence.DeletionRequestFilter{
			AIPUUID: ref.New(aipID),
			Status:  ref.New(enums.DeletionRequestStatusApproved),
		}).
		Return([]*types.DeletionRequest{
			{
				DBID:         1,
				UUID:         uuid.MustParse("323e4567-e89b-12d3-a456-426614174000"),
				WorkflowDBID: 1,
				Reason:       "Test reason for deletion",
				RequestedAt:  time.Date(2025, 10, 26, 8, 20, 40, 0, time.UTC),
				Requester:    "requester@example.com",
				ReviewedAt:   time.Date(2025, 10, 27, 8, 20, 40, 0, time.UTC),
				Reviewer:     "reviewer@example.com",
				Status:       enums.DeletionRequestStatusApproved,
				AIPUUID:      aipID,
			},
		}, nil)
}

func expectReadWorkflows(msvc *fake.MockService, id int) {
	msvc.EXPECT().
		ReadWorkflow(mockutil.Context(), id).
		Return(&types.Workflow{
			DBID:        id,
			CompletedAt: time.Date(2025, 10, 28, 9, 30, 50, 0, time.UTC),
		}, nil)
}

func expectLocation(t *testing.T, msvc *fake.MockService) {
	t.Helper()

	loc, err := storage.NewInternalLocation(&storage.LocationConfig{URL: "mem://"})
	if err != nil {
		t.Fatalf("couldn't create internal location: %v", err)
	}

	msvc.EXPECT().
		Location(mockutil.Context(), uuid.Nil).
		Return(loc, nil)
}

func expectUpdateAIP(t *testing.T, msvc *fake.MockService, aipID uuid.UUID) {
	msvc.EXPECT().
		UpdateAIP(
			mockutil.Context(),
			aipID,
			mockutil.Func(
				"Updates AIP deletion report key",
				func(updater persistence.AIPUpdater) error {
					got, err := updater(&types.AIP{})
					assert.NilError(t, err)
					assert.DeepEqual(t, got, &types.AIP{
						DeletionReportKey: ref.New(
							storage.ReportPrefix + "aip_deletion_report_123e4567-e89b-12d3-a456-426614174000.pdf",
						),
					})
					return nil
				},
			),
		).
		Return(&types.AIP{}, nil)
}

func expectSvc(t *testing.T, msvc *fake.MockService, aipID uuid.UUID) {
	t.Helper()

	expectReadAIP(msvc, aipID)
	expectListDeletionRequests(msvc, aipID)
	expectReadWorkflows(msvc, 1)
	expectLocation(t, msvc)
	expectUpdateAIP(t, msvc, aipID)
}

func expectFormFill(t *testing.T, mff *pdf_fake.MockFormFiller, aipID uuid.UUID) {
	t.Helper()

	mff.EXPECT().
		FillForm(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)
}

func TestAIPDeletionReportActivity(t *testing.T) {
	t.Parallel()

	aipID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	templatePath := "../../../assets/Enduro_AIP_deletion_report_v3.tmpl.pdf"

	type test struct {
		name             string
		templatePath     string
		expectedSvc      func(*testing.T, *fake.MockService, uuid.UUID)
		expectedFormFill func(*testing.T, *pdf_fake.MockFormFiller, uuid.UUID)
		params           activities.AIPDeletionReportActivityParams
		want             activities.AIPDeletionReportActivityResult
		wantErr          string
	}
	for _, tc := range []test{
		{
			name:             "Generate AIP Deletion Report",
			templatePath:     templatePath,
			expectedSvc:      expectSvc,
			expectedFormFill: expectFormFill,
			params: activities.AIPDeletionReportActivityParams{
				AIPID:          aipID,
				LocationSource: enums.LocationSourceAmss,
			},
			want: activities.AIPDeletionReportActivityResult{
				Key: storage.ReportPrefix + "aip_deletion_report_123e4567-e89b-12d3-a456-426614174000.pdf",
			},
		},
		{
			name: "Errors if report template is empty",
			params: activities.AIPDeletionReportActivityParams{
				AIPID:          aipID,
				LocationSource: enums.LocationSourceAmss,
			},
			wantErr: "AIP deletion report: template path is not configured",
		},
		{
			name:         "Errors if report template doesn't exist",
			templatePath: "non_existent.tmpl.pdf",
			params: activities.AIPDeletionReportActivityParams{
				AIPID:          aipID,
				LocationSource: enums.LocationSourceAmss,
			},
			wantErr: "AIP deletion report: template file not found: non_existent.tmpl.pdf",
		},
		{
			name:         "Errors if AIP is not found",
			templatePath: templatePath,
			expectedSvc: func(t *testing.T, msvc *fake.MockService, aipID uuid.UUID) {
				msvc.EXPECT().
					ReadAip(mockutil.Context(), aipID).
					Return(nil, &goastorage.AIPNotFound{
						Message: "AIP not found",
						UUID:    aipID,
					})
			},
			params: activities.AIPDeletionReportActivityParams{
				AIPID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			wantErr: "AIP deletion report: load data: ReadAip: AIP not found",
		},
		{
			name:         "Errors if ListDeletionRequests fails",
			templatePath: templatePath,
			expectedSvc: func(t *testing.T, msvc *fake.MockService, aipID uuid.UUID) {
				expectReadAIP(msvc, aipID)
				msvc.EXPECT().
					ListDeletionRequests(mockutil.Context(), &persistence.DeletionRequestFilter{
						AIPUUID: ref.New(aipID),
						Status:  ref.New(enums.DeletionRequestStatusApproved),
					}).
					Return(nil, errors.New("internal error"))
			},
			params: activities.AIPDeletionReportActivityParams{
				AIPID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			wantErr: "AIP deletion report: load data: ListDeletionRequests: internal error",
		},
		{
			name:         "Errors if no approved deletion requests found",
			templatePath: templatePath,
			expectedSvc: func(t *testing.T, msvc *fake.MockService, aipID uuid.UUID) {
				expectReadAIP(msvc, aipID)
				msvc.EXPECT().
					ListDeletionRequests(mockutil.Context(), &persistence.DeletionRequestFilter{
						AIPUUID: ref.New(aipID),
						Status:  ref.New(enums.DeletionRequestStatusApproved),
					}).
					Return([]*types.DeletionRequest{}, nil)
			},
			params: activities.AIPDeletionReportActivityParams{
				AIPID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			wantErr: "AIP deletion report: no approved deletion request found for AIP 123e4567-e89b-12d3-a456-426614174000",
		},
		{
			name:         "Errors if ReadWorkflow fails",
			templatePath: templatePath,
			expectedSvc: func(t *testing.T, msvc *fake.MockService, aipID uuid.UUID) {
				expectReadAIP(msvc, aipID)
				expectListDeletionRequests(msvc, aipID)
				msvc.EXPECT().
					ReadWorkflow(mockutil.Context(), 1).
					Return(nil, errors.New("internal error"))
			},
			params: activities.AIPDeletionReportActivityParams{
				AIPID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			wantErr: "AIP deletion report: load data: ReadWorkflow: internal error",
		},
		{
			name:         "Errors if updating AIP fails",
			templatePath: templatePath,
			expectedSvc: func(t *testing.T, msvc *fake.MockService, aipID uuid.UUID) {
				expectReadAIP(msvc, aipID)
				expectListDeletionRequests(msvc, aipID)
				expectReadWorkflows(msvc, 1)
				expectLocation(t, msvc)
				msvc.EXPECT().
					UpdateAIP(
						mockutil.Context(),
						aipID,
						mockutil.Func(
							"Updates deletion report key",
							func(updater persistence.AIPUpdater) error {
								_, err := updater(&types.AIP{})
								return err
							},
						),
					).
					Return(nil, errors.New("internal error"))
			},
			expectedFormFill: expectFormFill,
			params: activities.AIPDeletionReportActivityParams{
				AIPID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			wantErr: "AIP deletion report: update AIP: internal error",
		},
		{
			name:         "Errors if form filling fails",
			templatePath: templatePath,
			expectedSvc: func(t *testing.T, msvc *fake.MockService, aipID uuid.UUID) {
				expectReadAIP(msvc, aipID)
				expectListDeletionRequests(msvc, aipID)
				expectReadWorkflows(msvc, 1)
				expectLocation(t, msvc)
			},
			expectedFormFill: func(t *testing.T, mff *pdf_fake.MockFormFiller, aipID uuid.UUID) {
				mff.EXPECT().
					FillForm(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("invalid json data"))
			},
			params: activities.AIPDeletionReportActivityParams{
				AIPID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			wantErr: "AIP deletion report: fill form: invalid json data",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			msvc := fake.NewMockService(gomock.NewController(t))
			if tc.expectedSvc != nil {
				tc.expectedSvc(t, msvc, aipID)
			}

			mff := pdf_fake.NewMockFormFiller(gomock.NewController(t))
			if tc.expectedFormFill != nil {
				tc.expectedFormFill(t, mff, aipID)
			}

			bucket, err := blob.OpenBucket(t.Context(), "mem://")
			if err != nil {
				t.Fatalf("failed to open in-memory bucket: %v", err)
			}
			defer bucket.Close()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				activities.NewAIPDeletionReportActivity(
					clockwork.NewFakeClockAt(time.Date(2025, 10, 30, 11, 15, 16, 0, time.UTC)),
					storage.AIPDeletionConfig{ReportTemplatePath: tc.templatePath},
					msvc,
					mff,
				).Execute,
				temporalsdk_activity.RegisterOptions{
					Name: activities.AIPDeletionReportActivityName,
				},
			)

			enc, err := env.ExecuteActivity(
				activities.AIPDeletionReportActivityName,
				&tc.params,
			)
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			assert.NilError(t, err)

			var res activities.AIPDeletionReportActivityResult
			_ = enc.Get(&res)
			assert.DeepEqual(t, &res, &tc.want)
		})
	}
}
