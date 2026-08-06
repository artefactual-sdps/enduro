package am_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.artefactual.dev/amclient"
	"go.artefactual.dev/amclient/amclienttest"
	"go.artefactual.dev/tools/mockutil"
	"go.artefactual.dev/tools/temporal"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/am"
)

func TestStartTransferActivity(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name            string
		statusCode      int
		serverVersion   string
		wantErr         string
		wantNonRetryErr bool
	}{
		{
			name:       "Returns transfer ID",
			statusCode: http.StatusAccepted,
		},
		{
			name:            "Returns an invalid credentials error",
			statusCode:      http.StatusUnauthorized,
			wantErr:         "invalid Archivematica credentials",
			wantNonRetryErr: true,
		},
		{
			name:            "Returns an insufficient permissions error",
			statusCode:      http.StatusForbidden,
			wantErr:         "insufficient Archivematica permissions",
			wantNonRetryErr: true,
		},
		{
			name:            "Returns a not found error",
			statusCode:      http.StatusNotFound,
			wantErr:         "Archivematica resource not found",
			wantNonRetryErr: true,
		},
		{
			name:       "Retries a request in progress",
			statusCode: http.StatusConflict,
			wantErr:    "transfer request is still in progress",
		},
		{
			name:            "Rejects a changed idempotent request",
			statusCode:      http.StatusUnprocessableEntity,
			wantErr:         "idempotency key was reused with a different transfer request",
			wantNonRetryErr: true,
		},
		{
			name:            "Preserves ordinary conflict handling without a key",
			statusCode:      http.StatusConflict,
			serverVersion:   "1.18.9",
			wantErr:         "Archivematica error: 409 Conflict",
			wantNonRetryErr: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			serverVersion := tt.serverVersion
			if serverVersion == "" {
				serverVersion = "1.19.0"
			}

			ctrl := gomock.NewController(t)
			packageService := amclienttest.NewMockPackageService(ctrl)
			var gotReq *amclient.PackageCreateRequest
			packageService.EXPECT().
				Create(mockutil.Context(), gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					req *amclient.PackageCreateRequest,
				) (*amclient.PackageCreateResponse, *amclient.Response, error) {
					gotReq = req
					resp := packageCreateResponse(tt.statusCode)
					if tt.statusCode == http.StatusAccepted {
						return &amclient.PackageCreateResponse{ID: "transfer-id"}, resp, nil
					}
					return nil, resp, &amclient.ErrorResponse{Response: resp.Response}
				})

			client := newStartTransferTestClient(
				t,
				packageService,
				func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("X-Archivematica-Version", serverVersion)
					w.WriteHeader(http.StatusNotImplemented)
				},
			)
			activity := am.NewStartTransferActivity(&am.Config{}, client)

			result, err := executeStartTransferActivity(t, activity)

			if tt.wantErr == "" {
				assert.NilError(t, err)
				assert.DeepEqual(t, result, &am.StartTransferActivityResult{TransferID: "transfer-id"})
			} else {
				assert.ErrorContains(t, err, tt.wantErr)
				assert.Equal(t, temporal.NonRetryableError(err), tt.wantNonRetryErr)
			}

			assert.Assert(t, gotReq != nil)
			if serverVersion == "1.19.0" {
				assert.Assert(t, strings.HasPrefix(gotReq.IdempotencyKey, "enduro-transfer-"))
			} else {
				assert.Equal(t, gotReq.IdempotencyKey, "")
			}
			if tt.statusCode == http.StatusAccepted {
				assert.DeepEqual(t, gotReq, &amclient.PackageCreateRequest{
					Name:             "Testing",
					Type:             "zipped bag",
					Path:             "/tmp",
					ProcessingConfig: "automated",
					AutoApprove:      new(true),
					IdempotencyKey:   gotReq.IdempotencyKey,
				})
			}
		})
	}
}

func TestStartTransferActivityVersionGate(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name          string
		serverVersion string
		wantKey       bool
	}{
		{
			name:          "Older server",
			serverVersion: "1.18.9",
		},
		{
			name:          "Minimum supported server",
			serverVersion: "1.19.0",
			wantKey:       true,
		},
		{
			name: "Unavailable server info",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			packageService := amclienttest.NewMockPackageService(ctrl)
			var gotKey string
			packageService.EXPECT().
				Create(mockutil.Context(), gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					req *amclient.PackageCreateRequest,
				) (*amclient.PackageCreateResponse, *amclient.Response, error) {
					gotKey = req.IdempotencyKey
					return &amclient.PackageCreateResponse{ID: "transfer-id"},
						packageCreateResponse(http.StatusAccepted), nil
				})

			client := newStartTransferTestClient(
				t,
				packageService,
				func(w http.ResponseWriter, _ *http.Request) {
					if tt.serverVersion != "" {
						w.Header().Set("X-Archivematica-Version", tt.serverVersion)
					}
					w.WriteHeader(http.StatusNotImplemented)
				},
			)
			activity := am.NewStartTransferActivity(&am.Config{}, client)

			result, err := executeStartTransferActivity(t, activity)

			assert.NilError(t, err)
			assert.DeepEqual(t, result, &am.StartTransferActivityResult{TransferID: "transfer-id"})
			assert.Equal(t, gotKey != "", tt.wantKey)
		})
	}
}

func TestStartTransferActivityChecksVersionEveryExecution(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	packageService := amclienttest.NewMockPackageService(ctrl)
	keys := make([]string, 0, 3)
	packageService.EXPECT().
		Create(mockutil.Context(), gomock.Any()).
		DoAndReturn(func(
			_ context.Context,
			req *amclient.PackageCreateRequest,
		) (*amclient.PackageCreateResponse, *amclient.Response, error) {
			keys = append(keys, req.IdempotencyKey)
			return &amclient.PackageCreateResponse{ID: "transfer-id"},
				packageCreateResponse(http.StatusAccepted), nil
		}).
		Times(3)

	versionRequests := 0
	client := newStartTransferTestClient(
		t,
		packageService,
		func(w http.ResponseWriter, _ *http.Request) {
			versionRequests++
			version := "1.19.0"
			if versionRequests == 1 {
				version = "1.18.9"
			}
			w.Header().Set("X-Archivematica-Version", version)
			w.WriteHeader(http.StatusNotImplemented)
		},
	)
	activity := am.NewStartTransferActivity(&am.Config{}, client)

	ts := &temporalsdk_testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()
	env.RegisterActivityWithOptions(activity.Execute, temporalsdk_activity.RegisterOptions{
		Name: am.StartTransferActivityName,
	})
	params := startTransferParams()
	for range 3 {
		encoded, err := env.ExecuteActivity(am.StartTransferActivityName, params)
		assert.NilError(t, err)
		var result am.StartTransferActivityResult
		assert.NilError(t, encoded.Get(&result))
	}

	assert.Equal(t, versionRequests, 3)
	assert.Equal(t, len(keys), 3)
	assert.Equal(t, keys[0], "")
	assert.Assert(t, strings.HasPrefix(keys[1], "enduro-transfer-"))
	assert.Equal(t, keys[2], keys[1])
}

func newStartTransferTestClient(
	t *testing.T,
	packageService amclient.PackageService,
	serverInfoHandler http.HandlerFunc,
) *amclient.Client {
	t.Helper()

	server := httptest.NewServer(serverInfoHandler)
	t.Cleanup(server.Close)
	client := amclient.NewClient(server.Client(), server.URL+"/", "test", "test")
	client.Package = packageService

	return client
}

func packageCreateResponse(statusCode int) *amclient.Response {
	return &amclient.Response{
		Response: &http.Response{
			StatusCode: statusCode,
			Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		},
	}
}

func executeStartTransferActivity(
	t *testing.T,
	activity *am.StartTransferActivity,
) (*am.StartTransferActivityResult, error) {
	t.Helper()

	ts := &temporalsdk_testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()
	env.RegisterActivityWithOptions(activity.Execute, temporalsdk_activity.RegisterOptions{
		Name: am.StartTransferActivityName,
	})

	encoded, err := env.ExecuteActivity(am.StartTransferActivityName, startTransferParams())
	if err != nil {
		return nil, err
	}

	var result am.StartTransferActivityResult
	if err := encoded.Get(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func startTransferParams() *am.StartTransferActivityParams {
	return &am.StartTransferActivityParams{
		Name:         "Testing",
		ZipPIP:       true,
		RelativePath: "/tmp",
	}
}
