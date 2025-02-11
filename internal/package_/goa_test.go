package package_

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
	"github.com/google/uuid"
	"go.artefactual.dev/tools/mockutil"
	"go.artefactual.dev/tools/ref"
	"go.uber.org/mock/gomock"
	"goa.design/goa/v3/security"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	authfake "github.com/artefactual-sdps/enduro/internal/api/auth/fake"
	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/persistence/fake"
	"github.com/artefactual-sdps/enduro/internal/timerange"
)

func TestJWTAuth(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		mock    func(tv *authfake.MockTokenVerifier, claims *auth.Claims)
		claims  *auth.Claims
		scopes  []string
		logged  string
		wantErr error
	}
	for _, tt := range []test{
		{
			name: "Verifies and adds claims to context",
			mock: func(tv *authfake.MockTokenVerifier, claims *auth.Claims) {
				tv.EXPECT().
					Verify(context.Background(), "abc").
					Return(claims, nil)
			},
			claims: &auth.Claims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{"*"},
			},
			scopes: []string{"package:read"},
		},
		{
			name: "Fails with unauthorized error",
			mock: func(tv *authfake.MockTokenVerifier, claims *auth.Claims) {
				tv.EXPECT().
					Verify(context.Background(), "abc").
					Return(nil, auth.ErrUnauthorized)
			},
			wantErr: ErrUnauthorized,
		},
		{
			name: "Fails with unauthorized error (logging)",
			mock: func(tv *authfake.MockTokenVerifier, claims *auth.Claims) {
				tv.EXPECT().
					Verify(context.Background(), "abc").
					Return(nil, fmt.Errorf("fail"))
			},
			logged:  `"level"=1 "msg"="failed to verify token" "err"="fail"`,
			wantErr: ErrUnauthorized,
		},
		{
			name: "Fails with forbidden error",
			mock: func(tv *authfake.MockTokenVerifier, claims *auth.Claims) {
				tv.EXPECT().
					Verify(context.Background(), "abc").
					Return(claims, nil)
			},
			claims: &auth.Claims{
				Email:         "info@artefactual.com",
				EmailVerified: true,
				Attributes:    []string{"package:list"},
			},
			scopes:  []string{"package:read"},
			wantErr: ErrForbidden,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var logged string
			logger := funcr.New(
				func(prefix, args string) { logged = args },
				funcr.Options{Verbosity: 1},
			)

			tvMock := authfake.NewMockTokenVerifier(gomock.NewController(t))
			tt.mock(tvMock, tt.claims)
			gw := &goaWrapper{
				packageImpl: &packageImpl{
					logger:        logger,
					tokenVerifier: tvMock,
				},
			}

			ctx, err := gw.JWTAuth(context.Background(), "abc", &security.JWTScheme{RequiredScopes: tt.scopes})
			assert.Equal(t, logged, tt.logged)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, auth.UserClaimsFromContext(ctx), tt.claims)
		})
	}
}

func nullUUID(s string) uuid.NullUUID {
	return uuid.NullUUID{
		UUID:  uuid.MustParse(s),
		Valid: true,
	}
}

var testSIPs = []*datatypes.SIP{
	{
		ID:         1,
		Name:       "Test package 1",
		WorkflowID: "workflow-1",
		RunID:      "c5f7c35a-d5a6-4e00-b4da-b036ce5b40bc",
		AIPID:      nullUUID("e2ace0da-8697-453d-9ea1-4c9b62309e54"),
		LocationID: nullUUID("146182ff-9923-4869-bca1-0bbc0f822025"),
		Status:     enums.SIPStatusDone,
		CreatedAt:  time.Date(2024, 9, 25, 9, 31, 10, 0, time.UTC),
		StartedAt: sql.NullTime{
			Time:  time.Date(2024, 9, 25, 9, 31, 11, 0, time.UTC),
			Valid: true,
		},
		CompletedAt: sql.NullTime{
			Time:  time.Date(2024, 9, 25, 9, 31, 12, 0, time.UTC),
			Valid: true,
		},
	},
	{
		ID:         2,
		Name:       "Test package 2",
		WorkflowID: "workflow-2",
		RunID:      "d1f172c6-4ec8-4488-8a09-eef422b024cc",
		AIPID:      nullUUID("ffdb12f4-1735-4022-b746-a9bf4a32109b"),
		LocationID: nullUUID("659a93a0-2a6a-4931-a505-f07f71f5b010"),
		Status:     enums.SIPStatusInProgress,
		CreatedAt:  time.Date(2024, 10, 1, 17, 13, 26, 0, time.UTC),
		StartedAt: sql.NullTime{
			Time:  time.Date(2024, 10, 1, 17, 13, 27, 0, time.UTC),
			Valid: true,
		},
		CompletedAt: sql.NullTime{
			Time:  time.Date(2024, 10, 1, 17, 13, 28, 0, time.UTC),
			Valid: true,
		},
	},
}

func TestList(t *testing.T) {
	t.Parallel()

	type test struct {
		name         string
		payload      *goapackage.ListPayload
		mockRecorder func(mr *persistence_fake.MockServiceMockRecorder)
		want         *goapackage.EnduroPackages
		wantErr      string
	}
	for _, tt := range []test{
		{
			name: "Returns all packages",
			mockRecorder: func(mr *persistence_fake.MockServiceMockRecorder) {
				mr.ListSIPs(
					mockutil.Context(),
					&persistence.SIPFilter{
						Sort: persistence.NewSort().AddCol("id", true),
					},
				).Return(
					testSIPs,
					&persistence.Page{Limit: 20, Total: 2},
					nil,
				)
			},
			want: &goapackage.EnduroPackages{
				Items: goapackage.EnduroStoredPackageCollection{
					{
						ID:          1,
						Name:        ref.New("Test package 1"),
						LocationID:  ref.New(uuid.MustParse("146182ff-9923-4869-bca1-0bbc0f822025")),
						Status:      "done",
						WorkflowID:  ref.New("workflow-1"),
						RunID:       ref.New("c5f7c35a-d5a6-4e00-b4da-b036ce5b40bc"),
						AipID:       ref.New("e2ace0da-8697-453d-9ea1-4c9b62309e54"),
						CreatedAt:   "2024-09-25T09:31:10Z",
						StartedAt:   ref.New("2024-09-25T09:31:11Z"),
						CompletedAt: ref.New("2024-09-25T09:31:12Z"),
					},
					{
						ID:          2,
						Name:        ref.New("Test package 2"),
						LocationID:  ref.New(uuid.MustParse("659a93a0-2a6a-4931-a505-f07f71f5b010")),
						Status:      "in progress",
						WorkflowID:  ref.New("workflow-2"),
						RunID:       ref.New("d1f172c6-4ec8-4488-8a09-eef422b024cc"),
						AipID:       ref.New("ffdb12f4-1735-4022-b746-a9bf4a32109b"),
						CreatedAt:   "2024-10-01T17:13:26Z",
						StartedAt:   ref.New("2024-10-01T17:13:27Z"),
						CompletedAt: ref.New("2024-10-01T17:13:28Z"),
					},
				},
				Page: &goapackage.EnduroPage{
					Limit: 20,
					Total: 2,
				},
			},
		},
		{
			name: "Returns filtered packages",
			payload: &goapackage.ListPayload{
				Name:                ref.New("Test package 1"),
				AipID:               ref.New("e2ace0da-8697-453d-9ea1-4c9b62309e54"),
				LocationID:          ref.New("146182ff-9923-4869-bca1-0bbc0f822025"),
				EarliestCreatedTime: ref.New("2024-09-25T09:30:00Z"),
				LatestCreatedTime:   ref.New("2024-09-25T09:40:00Z"),
				Status:              ref.New("done"),
				Limit:               ref.New(10),
				Offset:              ref.New(1),
			},
			mockRecorder: func(mr *persistence_fake.MockServiceMockRecorder) {
				mr.ListSIPs(
					mockutil.Context(),
					&persistence.SIPFilter{
						Name:       ref.New("Test package 1"),
						AIPID:      ref.New(uuid.MustParse("e2ace0da-8697-453d-9ea1-4c9b62309e54")),
						LocationID: ref.New(uuid.MustParse("146182ff-9923-4869-bca1-0bbc0f822025")),
						CreatedAt: &timerange.Range{
							Start: time.Date(2024, 9, 25, 9, 30, 0, 0, time.UTC),
							End:   time.Date(2024, 9, 25, 9, 40, 0, 0, time.UTC),
						},
						Status: ref.New(enums.SIPStatusDone),
						Sort:   persistence.NewSort().AddCol("id", true),
						Page: persistence.Page{
							Limit:  10,
							Offset: 1,
						},
					},
				).Return(
					testSIPs[0:1],
					&persistence.Page{Limit: 10, Total: 1},
					nil,
				)
			},
			want: &goapackage.EnduroPackages{
				Items: goapackage.EnduroStoredPackageCollection{
					{
						ID:          1,
						Name:        ref.New("Test package 1"),
						LocationID:  ref.New(uuid.MustParse("146182ff-9923-4869-bca1-0bbc0f822025")),
						Status:      "done",
						WorkflowID:  ref.New("workflow-1"),
						RunID:       ref.New("c5f7c35a-d5a6-4e00-b4da-b036ce5b40bc"),
						AipID:       ref.New("e2ace0da-8697-453d-9ea1-4c9b62309e54"),
						CreatedAt:   "2024-09-25T09:31:10Z",
						StartedAt:   ref.New("2024-09-25T09:31:11Z"),
						CompletedAt: ref.New("2024-09-25T09:31:12Z"),
					},
				},
				Page: &goapackage.EnduroPage{
					Limit: 10,
					Total: 1,
				},
			},
		},
		{
			name: "Errors on an internal service error",
			payload: &goapackage.ListPayload{
				Name: ref.New("Package 42"),
			},
			mockRecorder: func(mr *persistence_fake.MockServiceMockRecorder) {
				mr.ListSIPs(
					mockutil.Context(),
					&persistence.SIPFilter{
						Name: ref.New("Package 42"),
						Sort: persistence.NewSort().AddCol("id", true),
					},
				).Return(
					[]*datatypes.SIP{},
					&persistence.Page{},
					persistence.ErrNotFound,
				)
			},
			wantErr: "not found error",
		},
		{
			name: "Errors on a bad aip_id",
			payload: &goapackage.ListPayload{
				AipID: ref.New("XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"),
			},
			wantErr: "aip_id: invalid UUID",
		},
		{
			name: "Errors on a bad location_id",
			payload: &goapackage.ListPayload{
				LocationID: ref.New("XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"),
			},
			wantErr: "location_id: invalid UUID",
		},
		{
			name: "Errors on a bad status",
			payload: &goapackage.ListPayload{
				Status: ref.New("meditating"),
			},
			wantErr: "invalid status",
		},
		{
			name: "Errors on a bad earliest_created_time",
			payload: &goapackage.ListPayload{
				EarliestCreatedTime: ref.New("2024-15-15T25:83:52Z"),
			},
			wantErr: "earliest_created_time: invalid time",
		},
		{
			name: "Errors on a bad latest_created_time",
			payload: &goapackage.ListPayload{
				LatestCreatedTime: ref.New("2024-15-15T25:83:52Z"),
			},
			wantErr: "latest_created_time: invalid time",
		},
		{
			name: "Errors on a bad created at range",
			payload: &goapackage.ListPayload{
				EarliestCreatedTime: ref.New("2024-10-01T17:43:52Z"),
				LatestCreatedTime:   ref.New("2023-10-01T17:43:52Z"),
			},
			wantErr: "time range: end cannot be before start",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			ctrl := gomock.NewController(t)
			svc := persistence_fake.NewMockService(ctrl)

			if tt.mockRecorder != nil {
				tt.mockRecorder(svc.EXPECT())
			}

			wrapper := goaWrapper{
				packageImpl: &packageImpl{
					logger: logr.Discard(),
					perSvc: svc,
				},
			}

			got, err := wrapper.List(ctx, tt.payload)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}
