package auth

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
)

type Claims struct {
	EmailVerified bool `json:"email_verified,omitempty"`
}

type TokenVerifier interface {
	Verify(ctx context.Context, token string) (bool, error)
}

type NoopTokenVerifier struct{}

var _ TokenVerifier = (*NoopTokenVerifier)(nil)

func (t *NoopTokenVerifier) Verify(ctx context.Context, token string) (bool, error) {
	return true, nil
}

type OIDCTokenVerifier struct {
	verifier *oidc.IDTokenVerifier
}

var _ TokenVerifier = (*OIDCTokenVerifier)(nil)

func NewOIDCTokenVerifier(ctx context.Context, cfg *OIDCConfig) (*OIDCTokenVerifier, error) {
	// Initialize an OIDC provider.
	provider, err := oidc.NewProvider(ctx, cfg.ProviderURL)
	if err != nil {
		return nil, err
	}

	// Create an ID token parser, but only trust ID tokens issued to this client id.
	return &OIDCTokenVerifier{
		verifier: provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
	}, nil
}

func (t *OIDCTokenVerifier) Verify(ctx context.Context, token string) (bool, error) {
	// Verify token.
	idToken, err := t.verifier.Verify(ctx, token)
	if err != nil {
		return false, err
	}

	// Extract custom claims.
	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		return false, err
	}

	// Check that claims are verified.
	if !claims.EmailVerified {
		return false, nil
	}

	return true, nil
}
