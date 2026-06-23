package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"goa.design/goa/v3/security"
	"gotest.tools/v3/assert"

	intabout "github.com/artefactual-sdps/enduro/internal/about"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/auth"
	intingest "github.com/artefactual-sdps/enduro/internal/ingest"
	ingestfake "github.com/artefactual-sdps/enduro/internal/ingest/fake"
	storagefake "github.com/artefactual-sdps/enduro/internal/storage/fake"
)

type testStorageService struct {
	*storagefake.MockService
	authCalls int
	claims    *auth.Claims
	token     string
}

func (s *testStorageService) BearerAuth(
	ctx context.Context,
	token string,
	_ *security.BearerScheme,
) (context.Context, error) {
	s.authCalls++
	s.token = token
	if s.claims != nil {
		ctx = auth.WithUserClaims(ctx, s.claims)
	}
	return ctx, nil
}

type testIngestService struct {
	*ingestfake.MockService
	authCalls int
	claims    *auth.Claims
	token     string
}

func (s *testIngestService) BearerAuth(
	ctx context.Context,
	token string,
	_ *security.BearerScheme,
) (context.Context, error) {
	s.authCalls++
	s.token = token
	if s.claims != nil {
		ctx = auth.WithUserClaims(ctx, s.claims)
	}
	return ctx, nil
}

type testAPI struct {
	storage *testStorageService
	ingest  *testIngestService
	handler http.Handler
}

func newTestAPI(t *testing.T) *testAPI {
	t.Helper()
	t.Setenv("ENDURO_API_CORS_ORIGIN", "http://example.com")

	ctrl := gomock.NewController(t)
	storageSvc := &testStorageService{MockService: storagefake.NewMockService(ctrl)}
	ingestSvc := &testIngestService{MockService: ingestfake.NewMockService(ctrl)}

	server := HTTPServer(
		logr.Discard(),
		slog.New(slog.DiscardHandler),
		nil,
		&Config{Listen: ":0"},
		ingestSvc,
		storageSvc,
		intabout.NewService(logr.Discard(), "", nil, intingest.UploadConfig{}, nil),
	)

	return &testAPI{
		storage: storageSvc,
		ingest:  ingestSvc,
		handler: server.Handler,
	}
}

func TestHTTPServer(t *testing.T) {
	t.Run("Storage", func(t *testing.T) {
		api := newTestAPI(t)

		locationID := uuid.MustParse("7fd0bb89-df4a-4aeb-a1bd-6db3907bb832")
		api.storage.EXPECT().
			ShowLocation(gomock.Any(), gomock.Any()).
			Return(&goastorage.Location{
				Name:      "Configured location",
				Source:    "s3",
				Purpose:   "aip_store",
				UUID:      locationID,
				CreatedAt: "2025-01-01T00:00:00Z",
				Config: goastorage.NewConfigS3(&goastorage.S3Config{
					Bucket: "archive",
					Region: "eu-west-1",
				}),
			}, nil)

		req := httptest.NewRequest(http.MethodGet, "/storage/locations/"+locationID.String(), nil)
		req.Header.Set("Authorization", "Bearer token")
		rec := httptest.NewRecorder()
		api.handler.ServeHTTP(rec, req)
		assert.Equal(t, rec.Code, http.StatusOK)

		var body map[string]any
		assert.NilError(t, json.NewDecoder(rec.Body).Decode(&body))
		assert.DeepEqual(t, body, map[string]any{
			"name":       "Configured location",
			"source":     "s3",
			"purpose":    "aip_store",
			"uuid":       locationID.String(),
			"created_at": "2025-01-01T00:00:00Z",
		})
	})

	t.Run("Ingest", func(t *testing.T) {
		api := newTestAPI(t)

		sipID := "d1845cb6-a5ea-474a-9ab8-26f9bcd919f5"
		sourceID := "58eb3d17-5678-4137-ad4f-471c9d9b207f"

		api.ingest.EXPECT().
			AddSip(gomock.Any(), gomock.Any()).
			Return(&goaingest.AddSipResult{UUID: sipID}, nil)

		req := httptest.NewRequest(
			http.MethodPost,
			"/ingest/sips?source_id="+sourceID+"&key=test-object.zip",
			nil,
		)
		req.Header.Set("Authorization", "Bearer token")
		rec := httptest.NewRecorder()
		api.handler.ServeHTTP(rec, req)
		assert.Equal(t, rec.Code, http.StatusCreated)

		var body map[string]any
		assert.NilError(t, json.NewDecoder(rec.Body).Decode(&body))
		assert.DeepEqual(t, body, map[string]any{
			"uuid": sipID,
		})
	})
}

func TestHTTPServerMonitorAuth(t *testing.T) {
	const token = "monitor-token"

	tests := []struct {
		name      string
		path      string
		setup     func(*testing.T, *testAPI, *auth.Claims)
		authCalls func(*testAPI) int
	}{
		{
			name: "Ingest",
			path: "/ingest/monitor",
			setup: func(t *testing.T, api *testAPI, claims *auth.Claims) {
				t.Helper()

				api.ingest.claims = claims
				api.ingest.EXPECT().
					Monitor(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, payload *goaingest.MonitorPayload, stream goaingest.MonitorServerStream) error {
						assert.Equal(t, api.ingest.token, token)
						assert.Assert(t, payload.Token != nil)
						assert.Equal(t, *payload.Token, token)
						assert.Equal(t, auth.UserClaimsFromContext(ctx), claims)
						assert.Assert(t, stream != nil)
						return nil
					})
			},
			authCalls: func(api *testAPI) int { return api.ingest.authCalls },
		},
		{
			name: "Storage",
			path: "/storage/monitor",
			setup: func(t *testing.T, api *testAPI, claims *auth.Claims) {
				t.Helper()

				api.storage.claims = claims
				api.storage.EXPECT().
					Monitor(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, payload *goastorage.MonitorPayload, stream goastorage.MonitorServerStream) error {
						assert.Equal(t, api.storage.token, token)
						assert.Assert(t, payload.Token != nil)
						assert.Equal(t, *payload.Token, token)
						assert.Equal(t, auth.UserClaimsFromContext(ctx), claims)
						assert.Assert(t, stream != nil)
						return nil
					})
			},
			authCalls: func(api *testAPI) int { return api.storage.authCalls },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := newTestAPI(t)
			claims := &auth.Claims{
				Email:      "monitor@example.com",
				Attributes: []string{"*"},
			}
			tt.setup(t, api, claims)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			api.handler.ServeHTTP(rec, req)

			assert.Equal(t, rec.Code, http.StatusOK)
			assert.Equal(t, tt.authCalls(api), 1)
		})
	}
}
