package ingest

import (
	"fmt"
	"net/http"
)

type BearerTransport struct {
	Base          http.RoundTripper
	TokenProvider AccessTokenProvider
}

func NewBearerTransport(base http.RoundTripper, provider AccessTokenProvider) *BearerTransport {
	return &BearerTransport{Base: base, TokenProvider: provider}
}

func (t *BearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.TokenProvider.AccessToken(req.Context())
	if err != nil {
		return nil, fmt.Errorf("get ingest storage API access token: %v", err)
	}

	clonedReq := req.Clone(req.Context())
	clonedReq.Header.Set("Authorization", "Bearer "+token)

	return t.Base.RoundTrip(clonedReq)
}
