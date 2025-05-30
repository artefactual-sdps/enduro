package about_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/about"
	"github.com/artefactual-sdps/enduro/internal/api/auth"
	authfake "github.com/artefactual-sdps/enduro/internal/api/auth/fake"
	goaabout "github.com/artefactual-sdps/enduro/internal/api/gen/about"
	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/poststorage"
	"github.com/artefactual-sdps/enduro/internal/preprocessing"
	"github.com/artefactual-sdps/enduro/internal/pres"
)

func TestJWTAuth(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		mock    func(tv *authfake.MockTokenVerifier, claims *auth.Claims)
		claims  *auth.Claims
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
				Attributes:    []string{},
			},
		},
		{
			name: "Fails with unauthorized error",
			mock: func(tv *authfake.MockTokenVerifier, claims *auth.Claims) {
				tv.EXPECT().
					Verify(context.Background(), "abc").
					Return(nil, auth.ErrUnauthorized)
			},
			wantErr: about.ErrUnauthorized,
		},
		{
			name: "Fails with unauthorized error (logging)",
			mock: func(tv *authfake.MockTokenVerifier, claims *auth.Claims) {
				tv.EXPECT().
					Verify(context.Background(), "abc").
					Return(nil, fmt.Errorf("fail"))
			},
			logged:  `"level"=1 "msg"="failed to verify token" "err"="fail"`,
			wantErr: about.ErrUnauthorized,
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
			srv := about.NewService(
				logger,
				"",
				preprocessing.Config{},
				[]poststorage.Config{},
				ingest.UploadConfig{},
				tvMock,
			)

			ctx, err := srv.JWTAuth(context.Background(), "abc", nil)
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

func TestAbout(t *testing.T) {
	t.Parallel()

	versionRegExp := regexp.MustCompile(`^\d+\.\d+\.\d+-dev$`)

	type test struct {
		name   string
		config config.Configuration
		want   *goaabout.EnduroAbout
	}
	for _, tt := range []test{
		{
			name:   "Empty config",
			config: config.Configuration{},
			want: &goaabout.EnduroAbout{
				Version:            "",
				PreservationSystem: "Unknown",
				Preprocessing: &goaabout.EnduroPreprocessing{
					Enabled:      false,
					WorkflowName: "",
					TaskQueue:    "",
				},
				Poststorage:   goaabout.EnduroPoststorageCollection{},
				UploadMaxSize: 0,
			},
		},
		{
			name:   "Preservation system: Archivematica",
			config: config.Configuration{Preservation: pres.Config{TaskQueue: "am"}},
			want: &goaabout.EnduroAbout{
				Version:            "",
				PreservationSystem: "Archivematica",
				Preprocessing: &goaabout.EnduroPreprocessing{
					Enabled:      false,
					WorkflowName: "",
					TaskQueue:    "",
				},
				Poststorage: goaabout.EnduroPoststorageCollection{},
			},
		},
		{
			name:   "Preservation system: a3m",
			config: config.Configuration{Preservation: pres.Config{TaskQueue: "a3m"}},
			want: &goaabout.EnduroAbout{
				Version:            "",
				PreservationSystem: "a3m",
				Preprocessing: &goaabout.EnduroPreprocessing{
					Enabled:      false,
					WorkflowName: "",
					TaskQueue:    "",
				},
				Poststorage: goaabout.EnduroPoststorageCollection{},
			},
		},
		{
			name: "Full config",
			config: config.Configuration{
				Preservation: pres.Config{TaskQueue: "a3m"},
				Preprocessing: preprocessing.Config{
					Enabled:    true,
					Extract:    true,
					SharedPath: "/tmp",
					Temporal: preprocessing.Temporal{
						Namespace:    "default",
						TaskQueue:    "preprocessing",
						WorkflowName: "preprocessing",
					},
				},
				Poststorage: []poststorage.Config{
					{
						Namespace:    "default",
						TaskQueue:    "poststorage",
						WorkflowName: "poststorage_1",
					},
					{
						Namespace:    "default",
						TaskQueue:    "poststorage",
						WorkflowName: "poststorage_2",
					},
				},
				Upload: ingest.UploadConfig{
					MaxSize: 12345678,
				},
			},
			want: &goaabout.EnduroAbout{
				Version:            "",
				PreservationSystem: "a3m",
				Preprocessing: &goaabout.EnduroPreprocessing{
					Enabled:      true,
					TaskQueue:    "preprocessing",
					WorkflowName: "preprocessing",
				},
				Poststorage: goaabout.EnduroPoststorageCollection{
					&goaabout.EnduroPoststorage{
						TaskQueue:    "poststorage",
						WorkflowName: "poststorage_1",
					},
					&goaabout.EnduroPoststorage{
						TaskQueue:    "poststorage",
						WorkflowName: "poststorage_2",
					},
				},
				UploadMaxSize: 12345678,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := about.NewService(
				logr.Discard(),
				tt.config.Preservation.TaskQueue,
				tt.config.Preprocessing,
				tt.config.Poststorage,
				tt.config.Upload,
				&auth.NoopTokenVerifier{},
			)
			res, err := srv.About(context.Background(), &goaabout.AboutPayload{})
			assert.NilError(t, err)
			assert.DeepEqual(t, res, tt.want, cmpopts.IgnoreFields(goaabout.EnduroAbout{}, "Version"))
			assert.Assert(t, versionRegExp.MatchString(res.Version))
		})
	}
}
