package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr/funcr"
	"go.uber.org/mock/gomock"
	"goa.design/goa/v3/security"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	authfake "github.com/artefactual-sdps/enduro/internal/api/auth/fake"
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
			scopes: []string{auth.StorageAIPSReadAttr},
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
				Attributes:    []string{auth.StorageAIPSListAttr},
			},
			scopes:  []string{auth.StorageAIPSReadAttr},
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
			svc := &serviceImpl{
				logger:        logger,
				tokenVerifier: tvMock,
			}

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
