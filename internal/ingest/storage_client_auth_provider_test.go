package ingest_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/ingest"
)

type tokenResponse struct {
	statusCode  int
	accessToken string
	expiresIn   int
}

// Some oauth2 client credentials paths can issue two HTTP requests for a
// single token attempt. Test cases must provide enough responses for the
// full retry budget to avoid indexing past this slice.
func tokenServer(t *testing.T, responses []tokenResponse) *httptest.Server {
	t.Helper()

	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		res := responses[calls-1]
		if res.statusCode != http.StatusOK {
			w.WriteHeader(res.statusCode)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]any{
			"token_type":   "Bearer",
			"access_token": res.accessToken,
			"expires_in":   res.expiresIn,
		})
		assert.NilError(t, err)
	}))

	t.Cleanup(srv.Close)

	return srv
}

func TestOIDCAccessTokenProviderAccessToken(t *testing.T) {
	t.Parallel()

	type test struct {
		name       string
		responses  []tokenResponse
		calls      int
		wantTokens []string
		wantErr    string
	}

	for _, tc := range []test{
		{
			name: "Caches token without expiry",
			responses: []tokenResponse{
				{statusCode: http.StatusOK, accessToken: "token-1"},
			},
			calls:      2,
			wantTokens: []string{"token-1", "token-1"},
		},
		{
			name: "Refreshes token that is within leeway window",
			responses: []tokenResponse{
				{statusCode: http.StatusOK, accessToken: "token-1", expiresIn: 1},
				{statusCode: http.StatusOK, accessToken: "token-2", expiresIn: 3600},
			},
			calls:      2,
			wantTokens: []string{"token-1", "token-2"},
		},
		{
			name: "Retries transient token endpoint failures",
			responses: []tokenResponse{
				{statusCode: http.StatusBadGateway},
				{statusCode: http.StatusBadGateway},
				{statusCode: http.StatusOK, accessToken: "token-ok", expiresIn: 3600},
			},
			calls:      1,
			wantTokens: []string{"token-ok"},
		},
		{
			name: "Returns error if token endpoint returns non-retryable error",
			responses: []tokenResponse{
				{statusCode: http.StatusBadRequest},
				{statusCode: http.StatusBadRequest},
			},
			calls:   1,
			wantErr: "request OIDC token: oauth2: cannot fetch token: 400 Bad Request",
		},
		{
			name: "Returns error after retry attempts are exhausted",
			responses: []tokenResponse{
				{statusCode: http.StatusBadGateway},
				{statusCode: http.StatusBadGateway},
				{statusCode: http.StatusBadGateway},
				{statusCode: http.StatusBadGateway},
				{statusCode: http.StatusBadGateway},
				{statusCode: http.StatusBadGateway},
			},
			calls:   1,
			wantErr: "request OIDC token: oauth2: cannot fetch token: 502 Bad Gateway",
		},
		{
			name: "Returns error if token endpoint returns an empty token",
			responses: []tokenResponse{
				{statusCode: http.StatusOK, accessToken: ""},
				{statusCode: http.StatusOK, accessToken: ""},
				{statusCode: http.StatusOK, accessToken: ""},
				{statusCode: http.StatusOK, accessToken: ""},
				{statusCode: http.StatusOK, accessToken: ""},
				{statusCode: http.StatusOK, accessToken: ""},
			},
			calls:   1,
			wantErr: "request OIDC token: oauth2: server response missing access_token",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := tokenServer(t, tc.responses)

			cfg := ingest.StorageOIDCConfig{
				Enabled:                 true,
				TokenURL:                srv.URL,
				ClientID:                "client-id",
				ClientSecret:            "client-secret",
				RetryMaxAttempts:        3,
				RetryInitialInterval:    time.Microsecond,
				RetryMaxInterval:        time.Microsecond,
				RetryBackoffCoefficient: 1.0,
				TokenExpiryLeeway:       30 * time.Second,
			}

			provider, err := ingest.NewOIDCAccessTokenProvider(t.Context(), cfg)
			assert.NilError(t, err)

			var tokens []string
			for range tc.calls {
				token, err := provider.AccessToken(t.Context())
				if tc.wantErr != "" {
					assert.ErrorContains(t, err, tc.wantErr)
					return
				}

				assert.NilError(t, err)
				tokens = append(tokens, token)
			}

			assert.DeepEqual(t, tokens, tc.wantTokens)
		})
	}
}
