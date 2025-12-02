package activities

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/mockutil"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/fake"
)

func MinIOUploadPreSignedURLHandler(t *testing.T) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		bytes, err := io.ReadAll(req.Body)
		defer req.Body.Close()

		assert.NilError(t, err)
		assert.DeepEqual(t, bytes, []byte("contents-of-the-aip"))
	}
}

func TestUploadActivity(t *testing.T) {
	t.Run("Activity runs successfully", func(t *testing.T) {
		minioTestServer := httptest.NewServer(http.HandlerFunc(MinIOUploadPreSignedURLHandler(t)))
		defer minioTestServer.Close()

		aipUUID := uuid.New().String()
		aipName := "aip.7z"

		mockClient := fake.NewMockClient(gomock.NewController(t))
		mockClient.EXPECT().
			SubmitAip(
				mockutil.Context(),
				&goastorage.SubmitAipPayload{
					UUID: aipUUID,
					Name: aipName,
				},
			).
			Return(
				&goastorage.SubmitAIPResult{
					URL: minioTestServer.URL + "/" + storage.AIPPrefix + "foobar.7z",
				},
				nil,
			)
		mockClient.EXPECT().
			SubmitAipComplete(
				mockutil.Context(),
				&goastorage.SubmitAipCompletePayload{UUID: aipUUID},
			).
			Return(nil)

		tmpDir := fs.NewDir(t, "", fs.WithFile("aip.7z", "contents-of-the-aip"))
		defer tmpDir.Remove()

		activity := NewUploadActivity(mockClient)

		_, err := activity.Execute(context.Background(), &UploadActivityParams{
			AIPPath: tmpDir.Join(aipName),
			AIPID:   aipUUID,
			Name:    aipName,
		})
		assert.NilError(t, err)
	})

	t.Run("Activity returns an error if final Update call fails", func(t *testing.T) {
		minioTestServer := httptest.NewServer(http.HandlerFunc(MinIOUploadPreSignedURLHandler(t)))
		defer minioTestServer.Close()

		aipUUID := uuid.New().String()
		aipName := "aip.7z"

		mockClient := fake.NewMockClient(gomock.NewController(t))
		mockClient.EXPECT().
			SubmitAip(
				mockutil.Context(),
				&goastorage.SubmitAipPayload{
					UUID: aipUUID,
					Name: aipName,
				},
			).
			Return(
				&goastorage.SubmitAIPResult{
					URL: minioTestServer.URL + "/" + storage.AIPPrefix + "foobar.7z",
				},
				nil,
			)
		mockClient.EXPECT().
			SubmitAipComplete(
				mockutil.Context(),
				&goastorage.SubmitAipCompletePayload{UUID: aipUUID},
			).
			Return(errors.New("update failed"))

		tmpDir := fs.NewDir(t, "", fs.WithFile("aip.7z", "contents-of-the-aip"))
		defer tmpDir.Remove()

		activity := NewUploadActivity(mockClient)

		_, err := activity.Execute(context.Background(), &UploadActivityParams{
			AIPPath: tmpDir.Join(aipName),
			AIPID:   aipUUID,
			Name:    aipName,
		})
		assert.Error(t, err, "update failed")
	})
}
