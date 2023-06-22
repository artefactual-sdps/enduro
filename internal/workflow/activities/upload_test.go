package activities

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"goa.design/goa/v3/security"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

// StorageService implements goastorage.Service.
type StorageService struct {
	OAuth2AuthHandler func(ctx context.Context, token string, scheme *security.OAuth2Scheme) (ctx2 context.Context, err error)
	SubmitHandler     func(ctx context.Context, req *goastorage.SubmitPayload) (
		res *goastorage.SubmitResult, err error)
	UpdateHandler    func(ctx context.Context, req *goastorage.UpdatePayload) (err error)
	DownloadHandler  func(ctx context.Context, req *goastorage.DownloadPayload) (res []byte, err error)
	LocationsHandler func(ctx context.Context, req *goastorage.LocationsPayload) (
		res goastorage.LocationCollection, err error)
	MoveHandler       func(ctx context.Context, req *goastorage.MovePayload) (err error)
	MoveStatusHandler func(ctx context.Context, req *goastorage.MoveStatusPayload) (
		res *goastorage.MoveStatusResult, err error)
	RejectHandler      func(ctx context.Context, req *goastorage.RejectPayload) (err error)
	ShowHandler        func(ctx context.Context, req *goastorage.ShowPayload) (res *goastorage.Package, err error)
	AddLocationHandler func(ctx context.Context, req *goastorage.AddLocationPayload) (
		res *goastorage.AddLocationResult, err error)
	ShowLocationHandler     func(ctx context.Context, req *goastorage.ShowLocationPayload) (res *goastorage.Location, err error)
	LocationPackagesHandler func(ctx context.Context, req *goastorage.LocationPackagesPayload) (
		res goastorage.PackageCollection, err error)
}

func (s StorageService) OAuth2Auth(ctx context.Context, token string, scheme *security.OAuth2Scheme) (ctx2 context.Context, err error) {
	return s.OAuth2AuthHandler(ctx, token, scheme)
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

func (s StorageService) Locations(ctx context.Context, req *goastorage.LocationsPayload) (res goastorage.LocationCollection, err error) {
	return s.LocationsHandler(ctx, req)
}

func (s StorageService) Move(ctx context.Context, req *goastorage.MovePayload) (err error) {
	return s.MoveHandler(ctx, req)
}

func (s StorageService) MoveStatus(ctx context.Context, req *goastorage.MoveStatusPayload) (res *goastorage.MoveStatusResult, err error) {
	return s.MoveStatusHandler(ctx, req)
}

func (s StorageService) Reject(ctx context.Context, req *goastorage.RejectPayload) (err error) {
	return s.RejectHandler(ctx, req)
}

func (s StorageService) Show(ctx context.Context, req *goastorage.ShowPayload) (res *goastorage.Package, err error) {
	return s.ShowHandler(ctx, req)
}

func (s StorageService) AddLocation(ctx context.Context, req *goastorage.AddLocationPayload) (res *goastorage.AddLocationResult, err error) {
	return s.AddLocationHandler(ctx, req)
}

func (s StorageService) ShowLocation(ctx context.Context, req *goastorage.ShowLocationPayload) (res *goastorage.Location, err error) {
	return s.ShowLocationHandler(ctx, req)
}

func (s StorageService) LocationPackages(ctx context.Context,
	req *goastorage.LocationPackagesPayload,
) (res goastorage.PackageCollection, err error) {
	return s.LocationPackagesHandler(ctx, req)
}

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

		fakeStorageService := StorageService{}
		fakeStorageService.OAuth2AuthHandler = func(ctx context.Context, token string,
			scheme *security.OAuth2Scheme,
		) (ctx2 context.Context, err error) {
			return ctx, nil
		}
		fakeStorageService.SubmitHandler = func(ctx context.Context,
			req *goastorage.SubmitPayload,
		) (res *goastorage.SubmitResult, err error) {
			return &goastorage.SubmitResult{
				URL: minioTestServer.URL + "/aips/foobar.7z",
			}, nil
		}
		fakeStorageService.UpdateHandler = func(ctx context.Context,
			req *goastorage.UpdatePayload,
		) (err error) {
			return nil
		}

		endpoints := goastorage.NewEndpoints(fakeStorageService)
		storageClient := goastorage.NewClient(
			endpoints.Submit,
			endpoints.Update,
			endpoints.Download,
			endpoints.Locations,
			endpoints.AddLocation,
			endpoints.Move,
			endpoints.MoveStatus,
			endpoints.Reject,
			endpoints.Show,
			endpoints.ShowLocation,
			endpoints.LocationPackages,
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
		fakeStorageService.OAuth2AuthHandler = func(ctx context.Context, token string,
			scheme *security.OAuth2Scheme,
		) (ctx2 context.Context, err error) {
			return ctx, nil
		}
		fakeStorageService.SubmitHandler = func(ctx context.Context,
			req *goastorage.SubmitPayload,
		) (res *goastorage.SubmitResult, err error) {
			return &goastorage.SubmitResult{
				URL: minioTestServer.URL + "/aips/foobar.7z",
			}, nil
		}
		fakeStorageService.UpdateHandler = func(ctx context.Context,
			req *goastorage.UpdatePayload,
		) (err error) {
			return errors.New("update failed")
		}

		endpoints := goastorage.NewEndpoints(fakeStorageService)
		storageClient := goastorage.NewClient(
			endpoints.Submit,
			endpoints.Update,
			endpoints.Download,
			endpoints.Locations,
			endpoints.AddLocation,
			endpoints.Move,
			endpoints.MoveStatus,
			endpoints.Reject,
			endpoints.Show,
			endpoints.ShowLocation,
			endpoints.LocationPackages,
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
