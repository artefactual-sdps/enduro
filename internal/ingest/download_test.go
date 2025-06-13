package ingest_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	goahttp "goa.design/goa/v3/http"
	"gocloud.dev/blob"
	"gocloud.dev/blob/memblob"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	persistence_fake "github.com/artefactual-sdps/enduro/internal/persistence/fake"
)

func TestDownloadSIP(t *testing.T) {
	sipUUID := uuid.New()
	key := "failed.zip"
	content := []byte("zipcontent")
	contentType := "application/zip"

	bucket := memblob.OpenBucket(nil)
	t.Cleanup(func() {
		if err := bucket.Close(); err != nil {
			t.Fatalf("close bucket: %v", err)
		}
	})

	w, err := bucket.NewWriter(t.Context(), key, &blob.WriterOptions{ContentType: contentType})
	assert.NilError(t, err)
	_, err = w.Write(content)
	assert.NilError(t, err)
	assert.NilError(t, w.Close())

	tests := []struct {
		name        string
		payloadUUID string
		sip         *datatypes.SIP
		perErr      error
		wantCode    int
		wantBody    []byte
		wantHeaders map[string]string
	}{
		{
			name:        "Fails to download a SIP (invalid UUID)",
			payloadUUID: "invalid-uuid",
			wantCode:    http.StatusBadRequest,
			wantBody:    []byte(`{"message":"invalid request"}` + "\n"),
			wantHeaders: map[string]string{"Content-Type": "application/json"},
		},
		{
			name:        "Fails to download a SIP (SIP not found)",
			payloadUUID: sipUUID.String(),
			perErr:      persistence.ErrNotFound,
			wantCode:    http.StatusNotFound,
			wantBody:    []byte(`{"message":"SIP not found"}` + "\n"),
			wantHeaders: map[string]string{"Content-Type": "application/json"},
		},
		{
			name:        "Fails to download a SIP (persistence error)",
			payloadUUID: sipUUID.String(),
			perErr:      persistence.ErrInternal,
			wantCode:    http.StatusInternalServerError,
			wantBody:    []byte(`{"message":"error reading SIP"}` + "\n"),
			wantHeaders: map[string]string{"Content-Type": "application/json"},
		},
		{
			name:        "Fails to download a SIP (missing failed values)",
			payloadUUID: sipUUID.String(),
			sip:         &datatypes.SIP{UUID: sipUUID},
			wantCode:    http.StatusBadRequest,
			wantBody:    []byte(`{"message":"SIP has no failed values"}` + "\n"),
			wantHeaders: map[string]string{"Content-Type": "application/json"},
		},
		{
			name:        "Fails to download a SIP (SIP file not found)",
			payloadUUID: sipUUID.String(),
			sip:         &datatypes.SIP{UUID: sipUUID, FailedAs: enums.SIPFailedAsSIP, FailedKey: "missing"},
			wantCode:    http.StatusNotFound,
			wantBody:    []byte(`{"message":"SIP file not found"}` + "\n"),
			wantHeaders: map[string]string{"Content-Type": "application/json"},
		},
		{
			name:        "Downloads a SIP",
			payloadUUID: sipUUID.String(),
			sip:         &datatypes.SIP{UUID: sipUUID, FailedAs: enums.SIPFailedAsSIP, FailedKey: key},
			wantCode:    http.StatusOK,
			wantBody:    content,
			wantHeaders: map[string]string{
				"Content-Disposition": fmt.Sprintf("attachment; filename=%q", key),
				"Content-Type":        contentType,
				"Content-Length":      strconv.Itoa(len(content)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := persistence_fake.NewMockService(gomock.NewController(t))
			if tt.sip != nil || tt.perErr != nil {
				u, err := uuid.Parse(tt.payloadUUID)
				assert.NilError(t, err)
				ctrl.EXPECT().ReadSIP(gomock.Any(), u).Return(tt.sip, tt.perErr)
			}

			svc := ingest.NewService(logr.Discard(), nil, nil, nil, ctrl, nil, nil, "", bucket, 0, nil)
			pattern := "/sips/%s/download"
			mux := goahttp.NewMuxer()
			handler := svc.DownloadSIP(mux, goahttp.RequestDecoder)
			mux.Handle("GET", fmt.Sprintf(pattern, "{uuid}"), handler)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf(pattern, tt.payloadUUID), nil)

			mux.ServeHTTP(rec, req)
			assert.Equal(t, rec.Code, tt.wantCode)
			assert.DeepEqual(t, rec.Body.Bytes(), tt.wantBody)
			for k, v := range tt.wantHeaders {
				assert.Equal(t, rec.Header().Get(k), v)
			}
		})
	}
}
