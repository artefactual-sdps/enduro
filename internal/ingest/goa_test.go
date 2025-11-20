package ingest_test

import (
	"context"
	"database/sql"
	"errors"
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
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/persistence/fake"
	"github.com/artefactual-sdps/enduro/internal/sipsource"
	sipsource_fake "github.com/artefactual-sdps/enduro/internal/sipsource/fake"
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
			wantErr: ingest.ErrUnauthorized,
		},
		{
			name: "Fails with unauthorized error (logging)",
			mock: func(tv *authfake.MockTokenVerifier, claims *auth.Claims) {
				tv.EXPECT().
					Verify(context.Background(), "abc").
					Return(nil, fmt.Errorf("fail"))
			},
			logged:  `"level"=1 "msg"="failed to verify token" "err"="fail"`,
			wantErr: ingest.ErrUnauthorized,
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
			wantErr: ingest.ErrForbidden,
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
			svc := ingest.NewService(ingest.ServiceParams{
				Logger:        logger,
				TokenVerifier: tvMock,
			})

			ctx, err := svc.JWTAuth(context.Background(), "abc", &security.JWTScheme{RequiredScopes: tt.scopes})
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
			Uploader: &datatypes.User{
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
						AipUUID:       ref.New("e2ace0da-8697-453d-9ea1-4c9b62309e54"),
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
						AipUUID:     ref.New("ffdb12f4-1735-4022-b746-a9bf4a32109b"),
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
				AipUUID:             ref.New("e2ace0da-8697-453d-9ea1-4c9b62309e54"),
				EarliestCreatedTime: ref.New("2024-09-25T09:30:00Z"),
				LatestCreatedTime:   ref.New("2024-09-25T09:40:00Z"),
				Status:              ref.New(enums.SIPStatusIngested.String()),
				UploaderUUID:        ref.New("0b075937-458c-43d9-b46c-222a072d62a9"),
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
						AipUUID:       ref.New("e2ace0da-8697-453d-9ea1-4c9b62309e54"),
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
				AipUUID: ref.New("XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"),
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
				UploaderUUID: ref.New("invalid"),
			},
			wantErr: "uploader_id: invalid UUID",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			ctrl := gomock.NewController(t)
			perSvc := persistence_fake.NewMockService(ctrl)

			if tt.mockRecorder != nil {
				tt.mockRecorder(perSvc.EXPECT())
			}

			svc := ingest.NewService(ingest.ServiceParams{
				PersistenceService: perSvc,
			})

			got, err := svc.ListSips(ctx, tt.payload)
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
			perSvc := persistence_fake.NewMockService(ctrl)

			if tt.mockRecorder != nil {
				tt.mockRecorder(perSvc.EXPECT())
			}

			svc := ingest.NewService(ingest.ServiceParams{
				PersistenceService: perSvc,
			})

			got, err := svc.ListUsers(ctx, tt.payload)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}

func TestListSIPSourceObjects(t *testing.T) {
	t.Parallel()

	sourceID := uuid.MustParse("cc6a61cd-ce26-4338-890a-8a4393f63eed")
	modTime := time.Now()

	type test struct {
		name         string
		payload      *goaingest.ListSipSourceObjectsPayload
		mockRecorder func(mr *sipsource_fake.MockSIPSourceMockRecorder)
		want         *goaingest.SIPSourceObjects
		wantErr      string
	}
	for _, tt := range []test{
		{
			name: "Returns SIP source objects with a next page value",
			payload: &goaingest.ListSipSourceObjectsPayload{
				UUID:  sourceID.String(),
				Limit: ref.New(10),
			},
			mockRecorder: func(mr *sipsource_fake.MockSIPSourceMockRecorder) {
				mr.ListObjects(
					mockutil.Context(),
					sipsource.ListOptions{
						Limit: 10,
						Sort:  sipsource.SortByModTime().Desc(),
					},
				).Return(
					&sipsource.Page{
						Objects: []*sipsource.Object{
							{Key: "object1", Size: 1234, ModTime: modTime},
						},
						Limit:     10,
						NextToken: []byte("next-token"),
					},
					nil,
				)
			},
			want: &goaingest.SIPSourceObjects{
				Objects: goaingest.SIPSourceObjectCollection{
					{
						Key:     "object1",
						Size:    ref.New(int64(1234)),
						ModTime: ref.New(modTime.Format(time.RFC3339)),
					},
				},
				Limit: 10,
				Next:  ref.New("next-token"),
			},
		},
		{
			name: "Returns SIP source objects when a cursor value is provided",
			payload: &goaingest.ListSipSourceObjectsPayload{
				UUID:   sourceID.String(),
				Limit:  ref.New(10),
				Cursor: ref.New("page-token"),
			},
			mockRecorder: func(mr *sipsource_fake.MockSIPSourceMockRecorder) {
				mr.ListObjects(
					mockutil.Context(),
					sipsource.ListOptions{
						Limit: 10,
						Token: []byte("page-token"),
						Sort:  sipsource.SortByModTime().Desc(),
					},
				).Return(
					&sipsource.Page{
						Objects: []*sipsource.Object{
							{Key: "object2", Size: 5678, ModTime: modTime},
						},
						Limit: 10,
					},
					nil,
				)
			},
			want: &goaingest.SIPSourceObjects{
				Objects: goaingest.SIPSourceObjectCollection{
					{
						Key:     "object2",
						Size:    ref.New(int64(5678)),
						ModTime: ref.New(modTime.Format(time.RFC3339)),
					},
				},
				Limit: 10,
			},
		},
		{
			name: "Returns a not found error when SIP source does not exist",
			mockRecorder: func(mr *sipsource_fake.MockSIPSourceMockRecorder) {
				mr.ListObjects(
					mockutil.Context(),
					sipsource.ListOptions{
						Sort: sipsource.SortByModTime().Desc(),
					},
				).Return(
					nil,
					sipsource.ErrInvalidSource,
				)
			},
			wantErr: "SIP Source not found",
		},
		{
			name: "Returns an error when a bad cursor token is provided",
			payload: &goaingest.ListSipSourceObjectsPayload{
				UUID:   sourceID.String(),
				Limit:  ref.New(10),
				Cursor: ref.New("bad-token"),
			},
			mockRecorder: func(mr *sipsource_fake.MockSIPSourceMockRecorder) {
				mr.ListObjects(
					mockutil.Context(),
					sipsource.ListOptions{
						Limit: 10,
						Token: []byte("bad-token"),
						Sort:  sipsource.SortByModTime().Desc(),
					},
				).Return(
					nil,
					sipsource.ErrInvalidToken,
				)
			},
			wantErr: "invalid cursor",
		},
		{
			name: "Returns an internal error",
			mockRecorder: func(mr *sipsource_fake.MockSIPSourceMockRecorder) {
				mr.ListObjects(
					mockutil.Context(),
					sipsource.ListOptions{
						Sort: sipsource.SortByModTime().Desc(),
					},
				).Return(
					nil,
					errors.New("internal error"),
				)
			},
			wantErr: "internal error",
		},
		{
			name: "Returns an empty page when no objects found",
			mockRecorder: func(mr *sipsource_fake.MockSIPSourceMockRecorder) {
				mr.ListObjects(
					mockutil.Context(),
					sipsource.ListOptions{
						Sort: sipsource.SortByModTime().Desc(),
					},
				).Return(
					&sipsource.Page{
						Objects: []*sipsource.Object{},
						Limit:   100,
					},
					nil,
				)
			},
			want: &goaingest.SIPSourceObjects{
				Objects: goaingest.SIPSourceObjectCollection{},
				Limit:   100,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			src := sipsource_fake.NewMockSIPSource(ctrl)
			if tt.mockRecorder != nil {
				tt.mockRecorder(src.EXPECT())
			}

			svc := ingest.NewService(ingest.ServiceParams{
				Logger:        logr.Discard(),
				UploadMaxSize: 1000000,
				SIPSource:     src,
			})

			got, err := svc.ListSipSourceObjects(t.Context(), tt.payload)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}
