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
	JWTAuthHandler          func(ctx context.Context, token string, scheme *security.JWTScheme) (ctx2 context.Context, err error)
	SubmitAipHandler        func(ctx context.Context, req *goastorage.SubmitAipPayload) (res *goastorage.SubmitAIPResult, err error)
	CreateAipHandler        func(ctx context.Context, req *goastorage.CreateAipPayload) (res *goastorage.AIP, err error)
	UpdateAipHandler        func(ctx context.Context, req *goastorage.UpdateAipPayload) (err error)
	DownloadAipHandler      func(ctx context.Context, req *goastorage.DownloadAipPayload) (res []byte, err error)
	ListLocationsHandler    func(ctx context.Context, req *goastorage.ListLocationsPayload) (res goastorage.LocationCollection, err error)
	MoveAipHandler          func(ctx context.Context, req *goastorage.MoveAipPayload) (err error)
	MoveAipStatusHandler    func(ctx context.Context, req *goastorage.MoveAipStatusPayload) (res *goastorage.MoveStatusResult, err error)
	RejectAipHandler        func(ctx context.Context, req *goastorage.RejectAipPayload) (err error)
	ShowAipHandler          func(ctx context.Context, req *goastorage.ShowAipPayload) (res *goastorage.AIP, err error)
	CreateLocationHandler   func(ctx context.Context, req *goastorage.CreateLocationPayload) (res *goastorage.CreateLocationResult, err error)
	ShowLocationHandler     func(ctx context.Context, req *goastorage.ShowLocationPayload) (res *goastorage.Location, err error)
	ListLocationAipsHandler func(ctx context.Context, req *goastorage.ListLocationAipsPayload) (res goastorage.AIPCollection, err error)
}

func (s StorageService) JWTAuth(
	ctx context.Context,
	token string,
	scheme *security.JWTScheme,
) (ctx2 context.Context, err error) {
	return s.JWTAuthHandler(ctx, token, scheme)
}

func (s StorageService) SubmitAip(
	ctx context.Context,
	req *goastorage.SubmitAipPayload,
) (res *goastorage.SubmitAIPResult, err error) {
	return s.SubmitAipHandler(ctx, req)
}

func (s StorageService) CreateAip(
	ctx context.Context,
	req *goastorage.CreateAipPayload,
) (res *goastorage.AIP, err error) {
	return s.CreateAipHandler(ctx, req)
}

func (s StorageService) UpdateAip(ctx context.Context, req *goastorage.UpdateAipPayload) (err error) {
	return s.UpdateAipHandler(ctx, req)
}

func (s StorageService) DownloadAip(ctx context.Context, req *goastorage.DownloadAipPayload) (res []byte, err error) {
	return s.DownloadAipHandler(ctx, req)
}

func (s StorageService) ListLocations(
	ctx context.Context,
	req *goastorage.ListLocationsPayload,
) (res goastorage.LocationCollection, err error) {
	return s.ListLocationsHandler(ctx, req)
}

func (s StorageService) MoveAip(ctx context.Context, req *goastorage.MoveAipPayload) (err error) {
	return s.MoveAipHandler(ctx, req)
}

func (s StorageService) MoveAipStatus(
	ctx context.Context,
	req *goastorage.MoveAipStatusPayload,
) (res *goastorage.MoveStatusResult, err error) {
	return s.MoveAipStatusHandler(ctx, req)
}

func (s StorageService) RejectAip(ctx context.Context, req *goastorage.RejectAipPayload) (err error) {
	return s.RejectAipHandler(ctx, req)
}

func (s StorageService) ShowAip(ctx context.Context, req *goastorage.ShowAipPayload) (res *goastorage.AIP, err error) {
	return s.ShowAipHandler(ctx, req)
}

func (s StorageService) CreateLocation(
	ctx context.Context,
	req *goastorage.CreateLocationPayload,
) (res *goastorage.CreateLocationResult, err error) {
	return s.CreateLocationHandler(ctx, req)
}

func (s StorageService) ShowLocation(
	ctx context.Context,
	req *goastorage.ShowLocationPayload,
) (res *goastorage.Location, err error) {
	return s.ShowLocationHandler(ctx, req)
}

func (s StorageService) ListLocationAips(
	ctx context.Context,
	req *goastorage.ListLocationAipsPayload,
) (res goastorage.AIPCollection, err error) {
	return s.ListLocationAipsHandler(ctx, req)
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
		fakeStorageService.JWTAuthHandler = func(ctx context.Context, token string, scheme *security.JWTScheme) (ctx2 context.Context, err error) {
			return ctx, nil
		}
		fakeStorageService.SubmitAipHandler = func(ctx context.Context, req *goastorage.SubmitAipPayload) (res *goastorage.SubmitAIPResult, err error) {
			return &goastorage.SubmitAIPResult{
				URL: minioTestServer.URL + "/aips/foobar.7z",
			}, nil
		}
		fakeStorageService.UpdateAipHandler = func(ctx context.Context, req *goastorage.UpdateAipPayload) (err error) {
			return nil
		}

		endpoints := goastorage.NewEndpoints(fakeStorageService)
		storageClient := goastorage.NewClient(
			endpoints.CreateAip,
			endpoints.SubmitAip,
			endpoints.UpdateAip,
			endpoints.DownloadAip,
			endpoints.MoveAip,
			endpoints.MoveAipStatus,
			endpoints.RejectAip,
			endpoints.ShowAip,
			endpoints.ListLocations,
			endpoints.CreateLocation,
			endpoints.ShowLocation,
			endpoints.ListLocationAips,
		)

		tmpDir := fs.NewDir(t, "", fs.WithFile("aip.7z", "contents-of-the-aip"))
		defer tmpDir.Remove()

		activity := NewUploadActivity(storageClient)

		_, err := activity.Execute(context.Background(), &UploadActivityParams{
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
		fakeStorageService.JWTAuthHandler = func(ctx context.Context, token string, scheme *security.JWTScheme) (ctx2 context.Context, err error) {
			return ctx, nil
		}
		fakeStorageService.SubmitAipHandler = func(ctx context.Context, req *goastorage.SubmitAipPayload) (res *goastorage.SubmitAIPResult, err error) {
			return &goastorage.SubmitAIPResult{
				URL: minioTestServer.URL + "/aips/foobar.7z",
			}, nil
		}
		fakeStorageService.UpdateAipHandler = func(ctx context.Context, req *goastorage.UpdateAipPayload) (err error) {
			return errors.New("update failed")
		}

		endpoints := goastorage.NewEndpoints(fakeStorageService)
		storageClient := goastorage.NewClient(
			endpoints.CreateAip,
			endpoints.SubmitAip,
			endpoints.UpdateAip,
			endpoints.DownloadAip,
			endpoints.MoveAip,
			endpoints.MoveAipStatus,
			endpoints.RejectAip,
			endpoints.ShowAip,
			endpoints.ListLocations,
			endpoints.CreateLocation,
			endpoints.ShowLocation,
			endpoints.ListLocationAips,
		)

		tmpDir := fs.NewDir(t, "", fs.WithFile("aip.7z", "contents-of-the-aip"))
		defer tmpDir.Remove()

		activity := NewUploadActivity(storageClient)

		_, err := activity.Execute(context.Background(), &UploadActivityParams{
			AIPPath: tmpDir.Join("aip.7z"),
			AIPID:   uuid.New().String(),
			Name:    "aip.7z",
		})
		assert.Error(t, err, "update failed")
	})
}
