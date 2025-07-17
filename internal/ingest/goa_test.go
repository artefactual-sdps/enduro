package ingest

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
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/entfilter"
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
			scopes: []string{auth.IngestSIPSReadAttr},
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
				Attributes:    []string{auth.IngestSIPSListAttr},
			},
			scopes:  []string{auth.IngestSIPSReadAttr},
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
				ingestImpl: &ingestImpl{
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

var (
	sipUUID1 = uuid.New()
	sipUUID2 = uuid.New()
	sipUUID3 = uuid.New()
	testSIPs = []*datatypes.SIP{
		{
			ID:        1,
			UUID:      sipUUID1,
			Name:      "Test SIP 1",
			AIPID:     nullUUID("e2ace0da-8697-453d-9ea1-4c9b62309e54"),
			Status:    enums.SIPStatusIngested,
			CreatedAt: time.Date(2024, 9, 25, 9, 31, 10, 0, time.UTC),
			StartedAt: sql.NullTime{
				Time:  time.Date(2024, 9, 25, 9, 31, 11, 0, time.UTC),
				Valid: true,
			},
			CompletedAt: sql.NullTime{
				Time:  time.Date(2024, 9, 25, 9, 31, 12, 0, time.UTC),
				Valid: true,
			},
			Uploader: &datatypes.Uploader{
				UUID:  uuid.MustParse("0b075937-458c-43d9-b46c-222a072d62a9"),
				Email: "uploader@example.com",
				Name:  "Test Uploader",
			},
		},
		{
			ID:        2,
			UUID:      sipUUID2,
			Name:      "Test SIP 2",
			AIPID:     nullUUID("ffdb12f4-1735-4022-b746-a9bf4a32109b"),
			Status:    enums.SIPStatusProcessing,
			CreatedAt: time.Date(2024, 10, 1, 17, 13, 26, 0, time.UTC),
			StartedAt: sql.NullTime{
				Time:  time.Date(2024, 10, 1, 17, 13, 27, 0, time.UTC),
				Valid: true,
			},
			CompletedAt: sql.NullTime{
				Time:  time.Date(2024, 10, 1, 17, 13, 28, 0, time.UTC),
				Valid: true,
			},
		},
		{
			ID:        3,
			UUID:      sipUUID3,
			Name:      "Test SIP 3",
			Status:    enums.SIPStatusError,
			CreatedAt: time.Date(2024, 10, 1, 17, 13, 26, 0, time.UTC),
			StartedAt: sql.NullTime{
				Time:  time.Date(2024, 10, 1, 17, 13, 27, 0, time.UTC),
				Valid: true,
			},
			CompletedAt: sql.NullTime{
				Time:  time.Date(2024, 10, 1, 17, 13, 28, 0, time.UTC),
				Valid: true,
			},
			FailedAs:  enums.SIPFailedAsSIP,
			FailedKey: "failed-key",
		},
	}
)

func TestListSIPs(t *testing.T) {
	t.Parallel()

	type test struct {
		name         string
		payload      *goaingest.ListSipsPayload
		mockRecorder func(mr *persistence_fake.MockServiceMockRecorder)
		want         *goaingest.SIPs
		wantErr      string
	}
	for _, tt := range []test{
		{
			name: "Returns all SIPs",
			mockRecorder: func(mr *persistence_fake.MockServiceMockRecorder) {
				mr.ListSIPs(
					mockutil.Context(),
					&persistence.SIPFilter{
						Sort: entfilter.NewSort().AddCol("id", true),
					},
				).Return(
					testSIPs,
					&persistence.Page{Limit: 20, Total: 3},
					nil,
				)
			},
			want: &goaingest.SIPs{
				Items: goaingest.SIPCollection{
					{
						UUID:          sipUUID1,
						Name:          ref.New("Test SIP 1"),
						Status:        enums.SIPStatusIngested.String(),
						AipID:         ref.New("e2ace0da-8697-453d-9ea1-4c9b62309e54"),
						CreatedAt:     "2024-09-25T09:31:10Z",
						StartedAt:     ref.New("2024-09-25T09:31:11Z"),
						CompletedAt:   ref.New("2024-09-25T09:31:12Z"),
						UploaderUUID:  ref.New(uuid.MustParse("0b075937-458c-43d9-b46c-222a072d62a9")),
						UploaderEmail: ref.New("uploader@example.com"),
						UploaderName:  ref.New("Test Uploader"),
					},
					{
						UUID:        sipUUID2,
						Name:        ref.New("Test SIP 2"),
						Status:      enums.SIPStatusProcessing.String(),
						AipID:       ref.New("ffdb12f4-1735-4022-b746-a9bf4a32109b"),
						CreatedAt:   "2024-10-01T17:13:26Z",
						StartedAt:   ref.New("2024-10-01T17:13:27Z"),
						CompletedAt: ref.New("2024-10-01T17:13:28Z"),
					},
					{
						UUID:        sipUUID3,
						Name:        ref.New("Test SIP 3"),
						Status:      enums.SIPStatusError.String(),
						CreatedAt:   "2024-10-01T17:13:26Z",
						StartedAt:   ref.New("2024-10-01T17:13:27Z"),
						CompletedAt: ref.New("2024-10-01T17:13:28Z"),
						FailedAs:    ref.New(enums.SIPFailedAsSIP.String()),
						FailedKey:   ref.New("failed-key"),
					},
				},
				Page: &goaingest.EnduroPage{
					Limit: 20,
					Total: 3,
				},
			},
		},
		{
			name: "Returns filtered SIPs",
			payload: &goaingest.ListSipsPayload{
				Name:                ref.New("Test SIP 1"),
				AipID:               ref.New("e2ace0da-8697-453d-9ea1-4c9b62309e54"),
				EarliestCreatedTime: ref.New("2024-09-25T09:30:00Z"),
				LatestCreatedTime:   ref.New("2024-09-25T09:40:00Z"),
				Status:              ref.New(enums.SIPStatusIngested.String()),
				UploaderID:          ref.New("0b075937-458c-43d9-b46c-222a072d62a9"),
				Limit:               ref.New(10),
				Offset:              ref.New(1),
			},
			mockRecorder: func(mr *persistence_fake.MockServiceMockRecorder) {
				mr.ListSIPs(
					mockutil.Context(),
					&persistence.SIPFilter{
						Name:  ref.New("Test SIP 1"),
						AIPID: ref.New(uuid.MustParse("e2ace0da-8697-453d-9ea1-4c9b62309e54")),
						CreatedAt: &timerange.Range{
							Start: time.Date(2024, 9, 25, 9, 30, 0, 0, time.UTC),
							End:   time.Date(2024, 9, 25, 9, 40, 0, 0, time.UTC),
						},
						Status:     ref.New(enums.SIPStatusIngested),
						UploaderID: ref.New(uuid.MustParse("0b075937-458c-43d9-b46c-222a072d62a9")),
						Sort:       entfilter.NewSort().AddCol("id", true),
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
			want: &goaingest.SIPs{
				Items: goaingest.SIPCollection{
					{
						UUID:          sipUUID1,
						Name:          ref.New("Test SIP 1"),
						Status:        enums.SIPStatusIngested.String(),
						AipID:         ref.New("e2ace0da-8697-453d-9ea1-4c9b62309e54"),
						CreatedAt:     "2024-09-25T09:31:10Z",
						StartedAt:     ref.New("2024-09-25T09:31:11Z"),
						CompletedAt:   ref.New("2024-09-25T09:31:12Z"),
						UploaderUUID:  ref.New(uuid.MustParse("0b075937-458c-43d9-b46c-222a072d62a9")),
						UploaderEmail: ref.New("uploader@example.com"),
						UploaderName:  ref.New("Test Uploader"),
					},
				},
				Page: &goaingest.EnduroPage{
					Limit: 10,
					Total: 1,
				},
			},
		},
		{
			name: "Errors on an internal service error",
			payload: &goaingest.ListSipsPayload{
				Name: ref.New("SIP 42"),
			},
			mockRecorder: func(mr *persistence_fake.MockServiceMockRecorder) {
				mr.ListSIPs(
					mockutil.Context(),
					&persistence.SIPFilter{
						Name: ref.New("SIP 42"),
						Sort: entfilter.NewSort().AddCol("id", true),
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
			payload: &goaingest.ListSipsPayload{
				AipID: ref.New("XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"),
			},
			wantErr: "aip_id: invalid UUID",
		},
		{
			name: "Errors on a bad status",
			payload: &goaingest.ListSipsPayload{
				Status: ref.New("meditating"),
			},
			wantErr: "status: invalid value",
		},
		{
			name: "Errors on a bad earliest_created_time",
			payload: &goaingest.ListSipsPayload{
				EarliestCreatedTime: ref.New("2024-15-15T25:83:52Z"),
			},
			wantErr: "created at: time range: cannot parse start time",
		},
		{
			name: "Errors on a bad latest_created_time",
			payload: &goaingest.ListSipsPayload{
				LatestCreatedTime: ref.New("2024-15-15T25:83:52Z"),
			},
			wantErr: "created at: time range: cannot parse end time",
		},
		{
			name: "Errors on a bad created at range",
			payload: &goaingest.ListSipsPayload{
				EarliestCreatedTime: ref.New("2024-10-01T17:43:52Z"),
				LatestCreatedTime:   ref.New("2023-10-01T17:43:52Z"),
			},
			wantErr: "created at: time range: end cannot be before start",
		},
		{
			name: "Errors on a bad uploader_id",
			payload: &goaingest.ListSipsPayload{
				UploaderID: ref.New("invalid"),
			},
			wantErr: "uploader_id: invalid UUID",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			ctrl := gomock.NewController(t)
			svc := persistence_fake.NewMockService(ctrl)

			if tt.mockRecorder != nil {
				tt.mockRecorder(svc.EXPECT())
			}

			wrapper := goaWrapper{
				ingestImpl: &ingestImpl{
					logger: logr.Discard(),
					perSvc: svc,
				},
			}

			got, err := wrapper.ListSips(ctx, tt.payload)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}

func TestListUsers(t *testing.T) {
	t.Parallel()

	userID1 := uuid.MustParse("0b075937-458c-43d9-b46c-222a072d62a9")
	userID2 := uuid.MustParse("a4400d29-6ba9-4843-aeb9-1039d68a3a5f")

	longStr := "XzesALZdoIEVAHleapPGvSMeAmTYrzUVoKDSavobUoegYMhpFGJPadSSjjujSSfasXVAmqhzcCRzHJTbmFxnkeCJbSfLfudPlTKndTVdnYCJpuAahELOuRzISmpVJRAZYTcaeRvlXmTwnPkyvCYOqYkFUyaEcmHbzaUkcOnSxJsxDTmeiCrGsJWMvUoxbbNpbgzrTkbauzDamhQivGbcFoKCaZruMiPXCwnWJxLLyMNHIIjhEHXMgQLwFCKnQViN"

	testUsers := []*datatypes.User{
		{
			UUID:      userID1,
			Email:     "user1@example.com",
			Name:      "User One",
			CreatedAt: time.Date(2025, 6, 22, 11, 33, 44, 0, time.UTC),
		},
		{
			UUID:      userID2,
			Email:     "user2@example.com",
			Name:      "User Two",
			CreatedAt: time.Date(2025, 6, 24, 12, 36, 48, 0, time.UTC),
		},
	}

	type test struct {
		name         string
		payload      *goaingest.ListUsersPayload
		mockRecorder func(mr *persistence_fake.MockServiceMockRecorder)
		want         *goaingest.Users
		wantErr      string
	}
	for _, tt := range []test{
		{
			name: "Returns all users",
			mockRecorder: func(mr *persistence_fake.MockServiceMockRecorder) {
				mr.ListUsers(
					mockutil.Context(),
					&persistence.UserFilter{},
				).Return(
					testUsers,
					&persistence.Page{Limit: 20, Total: 2},
					nil,
				)
			},
			want: &goaingest.Users{
				Items: goaingest.UserCollection{
					{
						UUID:      userID1,
						Email:     "user1@example.com",
						Name:      "User One",
						CreatedAt: "2025-06-22T11:33:44Z",
					},
					{
						UUID:      userID2,
						Email:     "user2@example.com",
						Name:      "User Two",
						CreatedAt: "2025-06-24T12:36:48Z",
					},
				},
				Page: &goaingest.EnduroPage{
					Limit: 20,
					Total: 2,
				},
			},
		},
		{
			name: "Filters users",
			payload: &goaingest.ListUsersPayload{
				Email:  ref.New("user1@example.com"),
				Name:   ref.New("User One"),
				Limit:  ref.New(10),
				Offset: ref.New(0),
			},
			mockRecorder: func(mr *persistence_fake.MockServiceMockRecorder) {
				mr.ListUsers(
					mockutil.Context(),
					&persistence.UserFilter{
						Email: ref.New("user1@example.com"),
						Name:  ref.New("User One"),
						Page: persistence.Page{
							Limit:  10,
							Offset: 0,
						},
					},
				).Return(
					testUsers[0:1],
					&persistence.Page{Limit: 10, Total: 1},
					nil,
				)
			},
			want: &goaingest.Users{
				Items: goaingest.UserCollection{
					{
						UUID:      userID1,
						Email:     "user1@example.com",
						Name:      "User One",
						CreatedAt: "2025-06-22T11:33:44Z",
					},
				},
				Page: &goaingest.EnduroPage{
					Limit: 10,
					Total: 1,
				},
			},
		},
		{
			name: "Returns error on email validation error",
			payload: &goaingest.ListUsersPayload{
				Email: ref.New(longStr),
			},
			wantErr: "email: exceeds maximum length of 255",
		},
		{
			name: "Returns error on name validation error",
			payload: &goaingest.ListUsersPayload{
				Name: ref.New(longStr),
			},
			wantErr: "name: exceeds maximum length of 255",
		},
		{
			name: "Returns error on internal service error",
			payload: &goaingest.ListUsersPayload{
				Email: ref.New("user1@example.com"),
			},
			mockRecorder: func(mr *persistence_fake.MockServiceMockRecorder) {
				mr.ListUsers(
					mockutil.Context(),
					&persistence.UserFilter{
						Email: ref.New("user1@example.com"),
					},
				).Return(
					nil,
					nil,
					persistence.ErrInternal,
				)
			},
			want: &goaingest.Users{
				Items: goaingest.UserCollection{},
				Page: &goaingest.EnduroPage{
					Limit: 20,
					Total: 1,
				},
			},
			wantErr: "internal error",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			ctrl := gomock.NewController(t)
			svc := persistence_fake.NewMockService(ctrl)

			if tt.mockRecorder != nil {
				tt.mockRecorder(svc.EXPECT())
			}

			wrapper := goaWrapper{
				ingestImpl: &ingestImpl{
					logger: logr.Discard(),
					perSvc: svc,
				},
			}

			got, err := wrapper.ListUsers(ctx, tt.payload)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}
