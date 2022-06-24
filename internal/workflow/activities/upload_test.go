package activities

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	goastorage "github.com/artefactual-labs/enduro/internal/api/gen/storage"
)

// StorageService implements goastorage.Service.
type StorageService struct {
	SubmitHandler     func(ctx context.Context, req *goastorage.SubmitPayload) (res *goastorage.SubmitResult, err error)
	UpdateHandler     func(ctx context.Context, req *goastorage.UpdatePayload) (err error)
	DownloadHandler   func(ctx context.Context, req *goastorage.DownloadPayload) (res []byte, err error)
	ListHandler       func(ctx context.Context) (res goastorage.StoredLocationCollection, err error)
	MoveHandler       func(ctx context.Context, req *goastorage.MovePayload) (err error)
	MoveStatusHandler func(ctx context.Context, req *goastorage.MoveStatusPayload) (res *goastorage.MoveStatusResult, err error)
}

func (s StorageService) Submit(ctx context.Context, req *goastorage.SubmitPayload) (res *goastorage.SubmitResult, err error) {
	return s.SubmitHandler(ctx, req)
}

func (s StorageService) Update(ctx context.Context, req *goastorage.UpdatePayload) (err error) {
	return s.UpdateHandler(ctx, req)
}

func (s StorageService) Download(ctx context.Context, req *goastorage.DownloadPayload) (res []byte, err error) {
	return s.DownloadHandler(ctx, req)
}

func (s StorageService) List(ctx context.Context) (res goastorage.StoredLocationCollection, err error) {
	return s.ListHandler(ctx)
}

func (s StorageService) Move(ctx context.Context, req *goastorage.MovePayload) (err error) {
	return s.MoveHandler(ctx, req)
}

func (s StorageService) MoveStatus(ctx context.Context, req *goastorage.MoveStatusPayload) (res *goastorage.MoveStatusResult, err error) {
	return s.MoveStatusHandler(ctx, req)
}

func MinIOUploadPreSignedURLHandler(t *testing.T) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		bytes, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()

		assert.NilError(t, err)
		assert.DeepEqual(t, bytes, []byte("contents-of-the-aip"))
	}
}

func TestUploadActivity(t *testing.T) {
	t.Run("Activity runs successfully", func(t *testing.T) {
		minioTestServer := httptest.NewServer(http.HandlerFunc(MinIOUploadPreSignedURLHandler(t)))
		defer minioTestServer.Close()

		fakeStorageService := StorageService{}
		fakeStorageService.SubmitHandler = func(ctx context.Context, req *goastorage.SubmitPayload) (res *goastorage.SubmitResult, err error) {
			return &goastorage.SubmitResult{
				URL: minioTestServer.URL + "/aips/foobar.7z",
			}, nil
		}
		fakeStorageService.UpdateHandler = func(ctx context.Context, req *goastorage.UpdatePayload) (err error) {
			return nil
		}

		endpoints := goastorage.NewEndpoints(fakeStorageService)
		storageClient := goastorage.NewClient(
			endpoints.Submit,
			endpoints.Update,
			endpoints.Download,
			endpoints.List,
			endpoints.Move,
			endpoints.MoveStatus,
		)

		tmpDir := fs.NewDir(t, "", fs.WithFile("aip.7z", "contents-of-the-aip"))
		defer tmpDir.Remove()

		activity := NewUploadActivity(storageClient)

		err := activity.Execute(context.Background(), &UploadActivityParams{
			AIPPath: tmpDir.Join("aip.7z"),
			AIPID:   uuid.New().String(),
			Name:    "aip.7z",
		})
		assert.NilError(t, err)
	})

	t.Run("Activity returns an error if final Update call fails", func(t *testing.T) {
		minioTestServer := httptest.NewServer(http.HandlerFunc(MinIOUploadPreSignedURLHandler(t)))
		defer minioTestServer.Close()

		fakeStorageService := StorageService{}
		fakeStorageService.SubmitHandler = func(ctx context.Context, req *goastorage.SubmitPayload) (res *goastorage.SubmitResult, err error) {
			return &goastorage.SubmitResult{
				URL: minioTestServer.URL + "/aips/foobar.7z",
			}, nil
		}
		fakeStorageService.UpdateHandler = func(ctx context.Context, req *goastorage.UpdatePayload) (err error) {
			return errors.New("update failed")
		}

		endpoints := goastorage.NewEndpoints(fakeStorageService)
		storageClient := goastorage.NewClient(
			endpoints.Submit,
			endpoints.Update,
			endpoints.Download,
			endpoints.List,
			endpoints.Move,
			endpoints.MoveStatus,
		)

		tmpDir := fs.NewDir(t, "", fs.WithFile("aip.7z", "contents-of-the-aip"))
		defer tmpDir.Remove()

		activity := NewUploadActivity(storageClient)

		err := activity.Execute(context.Background(), &UploadActivityParams{
			AIPPath: tmpDir.Join("aip.7z"),
			AIPID:   uuid.New().String(),
			Name:    "aip.7z",
		})
		assert.Error(t, err, "update failed")
	})
}
