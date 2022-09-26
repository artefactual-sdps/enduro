package auth_test

import (
	"context"
	"testing"
	"time"

	"chainguard.dev/go-oidctest/pkg/oidctest"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
)

const audience = "test-audience"

func token(t *testing.T, signer jose.Signer, iss string, claims interface{}) (token string) {
	t.Helper()

	// Use signed builder to generate token with given claims.
	builder := jwt.Signed(signer).
		Claims(jwt.Claims{
			Issuer:   iss,
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Expiry:   jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
			Subject:  "test-subject",
			Audience: jwt.Audience{audience},
		})

	// Include extra claims.
	if claims != nil {
		builder = builder.Claims(claims)
	}

	// Serialize token.
	token, err := builder.CompactSerialize()
	assert.NilError(t, err)

	return token
}

func TestOIDCTokenVerifier(t *testing.T) {
	t.Run("Verifies tokens with email verified", func(t *testing.T) {
		signer, iss := oidctest.NewIssuer(t)
		token := token(t, signer, iss, auth.Claims{EmailVerified: true})

		ctx := context.Background()
		v, err := auth.NewOIDCTokenVerifier(ctx, &auth.OIDCConfig{
			ProviderURL: iss,
			ClientID:    audience,
		})
		assert.NilError(t, err)

		verified, err := v.Verify(ctx, token)
		assert.NilError(t, err)
		assert.Assert(t, verified == true)
	})

	t.Run("Rejects tokens without email verified", func(t *testing.T) {
		signer, iss := oidctest.NewIssuer(t)
		token := token(t, signer, iss, auth.Claims{EmailVerified: false})

		ctx := context.Background()
		v, err := auth.NewOIDCTokenVerifier(ctx, &auth.OIDCConfig{
			ProviderURL: iss,
			ClientID:    audience,
		})
		assert.NilError(t, err)

		verified, err := v.Verify(ctx, token)
		assert.NilError(t, err)
		assert.Assert(t, verified == false)
	})

	t.Run("Rejects tokens under other errorful conditions", func(t *testing.T) {
		signer, iss := oidctest.NewIssuer(t)
		token := token(t, signer, iss, auth.Claims{EmailVerified: false})

		ctx := context.Background()
		v, err := auth.NewOIDCTokenVerifier(ctx, &auth.OIDCConfig{
			ProviderURL: iss,
			ClientID:    "--- wrong-audience ---",
		})
		assert.NilError(t, err)

		verified, err := v.Verify(ctx, token)
		assert.Error(t, err, "oidc: expected audience \"--- wrong-audience ---\" got [\"test-audience\"]")
		assert.Assert(t, verified == false)
	})

	t.Run("Constructor fails when context is canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := auth.NewOIDCTokenVerifier(ctx, &auth.OIDCConfig{
			ProviderURL: "http://test",
		})
		assert.Error(t, err, "Get \"http://test/.well-known/openid-configuration\": context canceled")
	})
}

func TestNoopTokenVerifier(t *testing.T) {
	t.Run("Verifies tokens", func(t *testing.T) {
		ctx := context.Background()
		v := &auth.NoopTokenVerifier{}

		verified, err := v.Verify(ctx, "")
		assert.NilError(t, err)
		assert.Assert(t, verified == true)
	})
}
