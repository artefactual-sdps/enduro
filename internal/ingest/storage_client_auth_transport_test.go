package ingest_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/ingest"
)

type roundTripper struct {
	req *http.Request
	err error
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.req = req
	if rt.err != nil {
		return nil, rt.err
	}

	return &http.Response{}, nil
}

type tokenProvider struct {
	token string
	err   error
}

func (p *tokenProvider) AccessToken(context.Context) (string, error) {
	return p.token, p.err
}

func TestBearerTransportRoundTrip(t *testing.T) {
	t.Parallel()

	type test struct {
		name          string
		providerToken string
		providerErr   error
		baseErr       error
		want          string
		wantErr       string
	}

	for _, tc := range []test{
		{
			name:          "Adds authorization header",
			providerToken: "token-1",
			want:          "Bearer token-1",
		},
		{
			name:        "Returns token provider errors",
			providerErr: errors.New("token error"),
			wantErr:     "get ingest storage API access token: token error",
		},
		{
			name:          "Returns base transport errors",
			providerToken: "token-3",
			baseErr:       errors.New("base transport error"),
			want:          "Bearer token-3",
			wantErr:       "base transport error",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			base := &roundTripper{err: tc.baseErr}
			transport := ingest.NewBearerTransport(base, &tokenProvider{
				token: tc.providerToken,
				err:   tc.providerErr,
			})

			req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
			assert.NilError(t, err)

			res, err := transport.RoundTrip(req)
			if res != nil && res.Body != nil {
				_ = res.Body.Close()
			}

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, base.req.Header.Get("Authorization"), tc.want)
		})
	}
}
