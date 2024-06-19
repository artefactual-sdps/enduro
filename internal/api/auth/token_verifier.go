package auth

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

var ErrUnauthorized error = errors.New("unauthorized")

type TokenVerifier interface {
	Verify(ctx context.Context, token string) (*Claims, error)
}

type NoopTokenVerifier struct{}

var _ TokenVerifier = (*NoopTokenVerifier)(nil)

func (t *NoopTokenVerifier) Verify(ctx context.Context, token string) (*Claims, error) {
	return nil, nil
}

type OIDCTokenVerifier struct {
	verifier *oidc.IDTokenVerifier
	cfg      *OIDCConfig
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
		cfg:      cfg,
	}, nil
}

func (t *OIDCTokenVerifier) Verify(ctx context.Context, token string) (*Claims, error) {
	// Verify token.
	idToken, err := t.verifier.Verify(ctx, token)
	if err != nil {
		return nil, err
	}

	// Extract custom claims.
	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	// Check that claims are verified.
	if !claims.EmailVerified {
		return nil, ErrUnauthorized
	}

	// Parse attributes.
	claims.Attributes, err = t.parseAttributes(idToken)
	if err != nil {
		return nil, err
	}

	return &claims, nil
}

// parseAttributes extracts the attributes used for access control from the token
// based on configuration. It finds the claim in the token based on the ClaimPath
// and ClaimPathSeparator values and filters the attrs. based on ClaimValuePrefix.
func (t *OIDCTokenVerifier) parseAttributes(token *oidc.IDToken) ([]string, error) {
	if !t.cfg.ABAC.Enabled {
		return nil, nil
	}

	var data map[string]interface{}
	if err := token.Claims(&data); err != nil {
		return nil, err
	}

	var keys []string
	if t.cfg.ABAC.ClaimPathSeparator != "" {
		keys = strings.Split(t.cfg.ABAC.ClaimPath, t.cfg.ABAC.ClaimPathSeparator)
	} else {
		keys = []string{t.cfg.ABAC.ClaimPath}
	}

	for i, key := range keys {
		value, ok := data[key]
		if !ok {
			return nil, fmt.Errorf("attributes not found in token, claim path: %s", t.cfg.ABAC.ClaimPath)
		}

		if i == len(keys)-1 {
			val := reflect.ValueOf(value)
			if val.Kind() != reflect.Slice {
				return nil, fmt.Errorf(
					"attributes are not part of a multivalue claim, claim path: %s",
					t.cfg.ABAC.ClaimPath,
				)
			}

			var filteredValue []string
			for i := range val.Len() {
				str, ok := val.Index(i).Interface().(string)
				if ok {
					if cutAttr, found := strings.CutPrefix(str, t.cfg.ABAC.ClaimValuePrefix); found {
						filteredValue = append(filteredValue, cutAttr)
					}
				}
			}

			return filteredValue, nil
		}

		nested, ok := value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("attributes not found in token, claim path: %s", t.cfg.ABAC.ClaimPath)
		}

		data = nested
	}

	return nil, fmt.Errorf("unexpected error parsing attributes")
}
