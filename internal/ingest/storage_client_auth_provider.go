package ingest

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type AccessTokenProvider interface {
	AccessToken(ctx context.Context) (string, error)
}

type OIDCAccessTokenProvider struct {
	mu                      sync.Mutex
	token                   *oauth2.Token
	cc                      clientcredentials.Config
	tokenExpiryLeeway       time.Duration
	retryMaxAttempts        int
	retryInitialInterval    time.Duration
	retryMaxInterval        time.Duration
	retryBackoffCoefficient float64
}

var _ AccessTokenProvider = (*OIDCAccessTokenProvider)(nil)

func NewOIDCAccessTokenProvider(ctx context.Context, cfg StorageOIDCConfig) (*OIDCAccessTokenProvider, error) {
	tokenURL := cfg.TokenURL
	if tokenURL == "" {
		provider, err := oidc.NewProvider(ctx, cfg.ProviderURL)
		if err != nil {
			return nil, fmt.Errorf("discover OIDC provider: %v", err)
		}
		tokenURL = provider.Endpoint().TokenURL
	}
	if tokenURL == "" {
		return nil, errors.New("missing OIDC token endpoint URL")
	}

	cc := clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     tokenURL,
		Scopes:       cfg.Scopes,
	}
	if cfg.Audience != "" {
		cc.EndpointParams = url.Values{"audience": []string{cfg.Audience}}
	}

	return &OIDCAccessTokenProvider{
		cc:                      cc,
		tokenExpiryLeeway:       cfg.TokenExpiryLeeway,
		retryMaxAttempts:        cfg.RetryMaxAttempts,
		retryInitialInterval:    cfg.RetryInitialInterval,
		retryMaxInterval:        cfg.RetryMaxInterval,
		retryBackoffCoefficient: cfg.RetryBackoffCoefficient,
	}, nil
}

func (p *OIDCAccessTokenProvider) AccessToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.tokenNeedsRefresh() {
		token, err := p.requestToken(ctx)
		if err != nil {
			return "", fmt.Errorf("request OIDC token: %v", err)
		}
		p.token = token
	}

	return p.token.AccessToken, nil
}

func (p *OIDCAccessTokenProvider) tokenNeedsRefresh() bool {
	if p.token == nil || p.token.AccessToken == "" {
		return true
	}
	if p.token.Expiry.IsZero() {
		return false
	}

	return p.token.Expiry.Before(time.Now().Add(p.tokenExpiryLeeway))
}

func (p *OIDCAccessTokenProvider) requestToken(ctx context.Context) (*oauth2.Token, error) {
	var err error
	var token *oauth2.Token
	for attempt := 1; attempt <= p.retryMaxAttempts; attempt++ {
		token, err = p.cc.TokenSource(ctx).Token()
		if err == nil {
			return token, nil
		}
		if !isRetryableErr(err) || attempt == p.retryMaxAttempts {
			break
		}

		time.Sleep(p.backoff(attempt))
	}

	return nil, err
}

func (p *OIDCAccessTokenProvider) backoff(attempt int) time.Duration {
	if p.retryInitialInterval <= 0 {
		return 0
	}

	wait := float64(p.retryInitialInterval) * math.Pow(p.retryBackoffCoefficient, float64(attempt-1))
	if p.retryMaxInterval > 0 {
		wait = math.Min(wait, float64(p.retryMaxInterval))
	}

	return time.Duration(wait)
}

func isRetryableErr(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var rerr *oauth2.RetrieveError
	if errors.As(err, &rerr) && rerr.Response != nil {
		return rerr.Response.StatusCode >= http.StatusInternalServerError
	}

	return true
}
