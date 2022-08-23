package storage_test

import (
	"context"
	"errors"
	"io"
	"net/url"
	"testing"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	temporalsdk_api_enums "go.temporal.io/api/enums/v1"
	temporalsdk_client "go.temporal.io/sdk/client"
	temporalsdk_mocks "go.temporal.io/sdk/mocks"
	goa "goa.design/goa/v3/pkg"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
	"gocloud.dev/blob/memblob"
	"gotest.tools/v3/assert"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/ref"
	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/fake"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

type setUpAttrs struct {
	logger         *logr.Logger
	config         *storage.Config
	persistence    *persistence.Storage
	temporalClient *temporalsdk_client.Client

	persistenceMock    *fake.MockStorage
	temporalClientMock *temporalsdk_mocks.Client
}

func setUpService(t *testing.T, attrs *setUpAttrs) storage.Service {
	t.Helper()

	psMock := fake.NewMockStorage(gomock.NewController(t))
	var ps persistence.Storage = psMock

	tcMock := &temporalsdk_mocks.Client{}
	var tc temporalsdk_client.Client = tcMock

	params := setUpAttrs{
		logger: ref.New(logr.Discard()),
		config: ref.New(storage.Config{
			Internal: storage.LocationConfig{
				Name:   "",
				Bucket: "internal",
				Region: "eu-west-2",
			},
		}),
		persistence:        &ps,
		persistenceMock:    psMock,
		temporalClient:     &tc,
		temporalClientMock: tcMock,
	}
	if attrs.logger != nil {
		params.logger = attrs.logger
	}
	if attrs.config != nil {
		params.config = attrs.config
	}
	if attrs.persistence != nil {
		params.persistence = attrs.persistence
	}
	if attrs.temporalClient != nil {
		params.temporalClient = attrs.temporalClient
	}

	*attrs = params

	s, err := storage.NewService(
		*params.logger,
		*params.config,
		*params.persistence,
		*params.temporalClient,
	)
	assert.NilError(t, err)

	return s
}

type fLocation struct {
	b *blob.Bucket
}

func (l *fLocation) UUID() uuid.UUID {
	return uuid.Nil
}

func (l *fLocation) Bucket() *blob.Bucket {
	return l.b
}

func (l *fLocation) Close() error {
	return nil
}

func fakeInternalLocation(t *testing.T, b *blob.Bucket) {
	t.Helper()

	t.Cleanup(func() { b.Close() })

	if b == nil {
		b = memblob.OpenBucket(nil)
	}

	storage.InternalLocationFactory = func(*storage.LocationConfig) (storage.Location, error) {
		return &fLocation{b}, nil
	}
}

func fakeLocationWithBucket(t *testing.T, b *blob.Bucket) *blob.Bucket {
	t.Helper()

	t.Cleanup(func() { b.Close() })

	if b == nil {
		b = memblob.OpenBucket(nil)
	}

	storage.LocationFactory = func(*goastorage.StoredLocation) (storage.Location, error) {
		return &fLocation{b}, nil
	}

	return b
}

func fakeLocationWithContents(t *testing.T, svc storage.Service, locationID uuid.UUID, objectKey, contents string) {
	t.Helper()

	b := memblob.OpenBucket(nil)
	t.Cleanup(func() { b.Close() })
	b.WriteAll(context.Background(), objectKey, []byte(contents), nil)

	storage.LocationFactory = func(location *goastorage.StoredLocation) (storage.Location, error) {
		return &fLocation{b}, nil
	}
}

func TestNewService(t *testing.T) {
	t.Parallel()

	_, err := storage.NewService(
		logr.Discard(),
		storage.Config{},
		nil,
		nil,
	)

	assert.ErrorContains(t, err, "s3blob.OpenBucket: bucketName is required")
}

func TestServiceSubmit(t *testing.T) {
	t.Parallel()

	t.Run("Returns not_valid if AIPID is invalid", func(t *testing.T) {
		t.Parallel()

		AIPID := "12345"
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		ret, err := svc.Submit(context.Background(), &goastorage.SubmitPayload{
			AipID: AIPID,
		})
		assert.Assert(t, ret == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "invalid UUID length: 5")
	})

	t.Run("Returns not_available if workflow cannot be executed", func(t *testing.T) {
		t.Parallel()

		AIPID := "5ab42bc3-acc2-420b-bbd0-76efdef94828"
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		attrs.temporalClientMock.
			On(
				"ExecuteWorkflow",
				mock.AnythingOfType("*context.timerCtx"),
				temporalsdk_client.StartWorkflowOptions{
					ID:                    "storage-upload-workflow-" + AIPID,
					TaskQueue:             "global",
					WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
				},
				"storage-upload-workflow",
				&storage.StorageUploadWorkflowRequest{AIPID: AIPID},
			).
			Return(
				nil,
				errors.New("something went wrong"),
			).
			Times(1)

		ret, err := svc.Submit(context.Background(), &goastorage.SubmitPayload{
			AipID: AIPID,
		})
		assert.Assert(t, ret == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_available")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns not_valid if package cannot be persisted", func(t *testing.T) {
		t.Parallel()

		AIPID := "5ab42bc3-acc2-420b-bbd0-76efdef94828"
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"ExecuteWorkflow",
				mock.AnythingOfType("*context.timerCtx"),
				temporalsdk_client.StartWorkflowOptions{
					ID:                    "storage-upload-workflow-" + AIPID,
					TaskQueue:             "global",
					WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
				},
				"storage-upload-workflow",
				&storage.StorageUploadWorkflowRequest{AIPID: AIPID},
			).
			Return(
				&temporalsdk_mocks.WorkflowRun{},
				nil,
			).
			Times(1)

		attrs.persistenceMock.
			EXPECT().
			CreatePackage(
				gomock.AssignableToTypeOf(ctx),
				gomock.Any(),
			).
			Return(
				nil,
				errors.New("database server error"),
			).
			Times(1)

		ret, err := svc.Submit(ctx, &goastorage.SubmitPayload{
			Name:  "package",
			AipID: AIPID,
		})
		assert.Assert(t, ret == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot persist package")
	})

	t.Run("Returns not_valid if signed URL cannot be generated", func(t *testing.T) {
		t.Parallel()

		AIPID := "5ab42bc3-acc2-420b-bbd0-76efdef94828"
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"ExecuteWorkflow",
				mock.AnythingOfType("*context.timerCtx"),
				temporalsdk_client.StartWorkflowOptions{
					ID:                    "storage-upload-workflow-" + AIPID,
					TaskQueue:             "global",
					WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
				},
				"storage-upload-workflow",
				&storage.StorageUploadWorkflowRequest{AIPID: AIPID},
			).
			Return(
				&temporalsdk_mocks.WorkflowRun{},
				nil,
			).
			Times(1)

		attrs.persistenceMock.
			EXPECT().
			CreatePackage(
				gomock.AssignableToTypeOf(ctx),
				gomock.Any(),
			).
			Return(
				&goastorage.StoredStoragePackage{},
				nil,
			).
			Times(1)

		fakeInternalLocation(t, nil)

		ret, err := svc.Submit(ctx, &goastorage.SubmitPayload{
			Name:  "package",
			AipID: AIPID,
		})
		assert.Assert(t, ret == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot persist package")
	})

	t.Run("Returns signed URL", func(t *testing.T) {
		t.Parallel()

		AIPID := "5ab42bc3-acc2-420b-bbd0-76efdef94828"
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"ExecuteWorkflow",
				mock.AnythingOfType("*context.timerCtx"),
				temporalsdk_client.StartWorkflowOptions{
					ID:                    "storage-upload-workflow-" + AIPID,
					TaskQueue:             "global",
					WorkflowIDReusePolicy: temporalsdk_api_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
				},
				"storage-upload-workflow",
				&storage.StorageUploadWorkflowRequest{AIPID: AIPID},
			).
			Return(
				&temporalsdk_mocks.WorkflowRun{},
				nil,
			).
			Times(1)

		attrs.persistenceMock.
			EXPECT().
			CreatePackage(
				gomock.AssignableToTypeOf(ctx),
				gomock.Any(),
			).
			Return(
				&goastorage.StoredStoragePackage{},
				nil,
			).
			Times(1)

		// Fake internal location, using fileblob because it can generate signed URLs.
		furl, err := url.Parse("file:///tmp/dir")
		assert.NilError(t, err)
		b, err := fileblob.OpenBucket("/tmp", &fileblob.Options{URLSigner: fileblob.NewURLSignerHMAC(furl, []byte("1234"))})
		assert.NilError(t, err)
		fakeInternalLocation(t, b)

		ret, err := svc.Submit(ctx, &goastorage.SubmitPayload{
			Name:  "package",
			AipID: AIPID,
		})
		assert.Equal(t, ret.URL[0:15], "file:///tmp/dir")
		assert.NilError(t, err)
	})
}

func TestServiceLocation(t *testing.T) {
	t.Parallel()

	svc := setUpService(t, &setUpAttrs{})

	testCases := map[string]struct {
		UUID uuid.UUID
		err  error
	}{
		"Returns internal location": {
			uuid.Nil,
			nil,
		},
		"Returns location": {
			uuid.MustParse("50110114-55ac-4567-b74f-9def601c6293"),
			nil,
		},
		"Returns error when location cannot be found": {
			uuid.MustParse("d8ea8946-dc82-4f4e-8c2d-8d3861f3297d"),
			errors.New("error loading location: unknown location d8ea8946-dc82-4f4e-8c2d-8d3861f3297d"),
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			loc, err := svc.Location(context.Background(), tc.UUID)

			if tc.err == nil {
				assert.NilError(t, err)
				assert.Equal(t, loc.UUID(), tc.UUID)
			} else {
				assert.Error(t, err, tc.err.Error())
			}
		})
	}
}

func TestServiceDownload(t *testing.T) {
	t.Parallel()

	svc := setUpService(t, &setUpAttrs{})

	blob, err := svc.Download(context.Background(), &goastorage.DownloadPayload{})
	assert.NilError(t, err)
	assert.DeepEqual(t, blob, []byte{}) // Not implemented.
}

func TestServiceList(t *testing.T) {
	t.Parallel()

	t.Run("Returns defined locations", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		storedLocations := goastorage.StoredLocationCollection{
			{
				ID:          1,
				Name:        "perma-aips-1",
				Description: ref.New("One"),
				Source:      "minio",
				Purpose:     "aip_store",
				UUID:        ref.New(uuid.New()),
			},
			{
				ID:          2,
				Name:        "perma-aips-2",
				Description: ref.New("Two"),
				Source:      "minio",
				Purpose:     "aip_store",
				UUID:        ref.New(uuid.New()),
			},
		}

		attrs.persistenceMock.
			EXPECT().
			ListLocations(ctx).
			Return(storedLocations, nil).
			Times(1)

		res, err := svc.Locations(ctx)
		assert.NilError(t, err)
		assert.DeepEqual(t, res, storedLocations)
	})
}

func TestReject(t *testing.T) {
	t.Run("Fails when passing an invalid UUID", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()
		AIPID := "7c09fa45-cdac-4874-90af-56dc86a6e73c"

		attrs.persistenceMock.
			EXPECT().
			UpdatePackageStatus(
				ctx,
				types.StatusRejected,
				uuid.MustParse(AIPID),
			).
			Return(nil).
			Times(1)

		err := svc.Reject(ctx, &goastorage.RejectPayload{AipID: AIPID})
		assert.NilError(t, err)
	})
}

func TestServiceReadPackage(t *testing.T) {
	t.Parallel()

	t.Run("Fails when passing an invalid UUID", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()

		pkg, err := svc.ReadPackage(ctx, "<invalid-uuid>")
		assert.Error(t, err, "invalid UUID length: 14")
		assert.Assert(t, pkg == nil)
	})
}

func TestServiceUpdatePackageStatus(t *testing.T) {
	t.Parallel()

	t.Run("Fails when passing an invalid UUID", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()

		err := svc.UpdatePackageStatus(ctx, types.StatusStored, "<invalid-uuid>")
		assert.Error(t, err, "invalid UUID length: 14")
	})

	t.Run("Returns if persistence failed", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()
		AIPID := "7c09fa45-cdac-4874-90af-56dc86a6e73c"

		attrs.persistenceMock.
			EXPECT().
			UpdatePackageStatus(
				ctx,
				types.StatusStored,
				uuid.MustParse(AIPID),
			).
			Return(errors.New("something is wrong")).
			Times(1)

		err := svc.UpdatePackageStatus(ctx, types.StatusStored, AIPID)
		assert.Error(t, err, "something is wrong")
	})
}

func TestServiceUpdatePackageLocationID(t *testing.T) {
	t.Parallel()

	t.Run("Fails when passing an invalid UUID", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()

		err := svc.UpdatePackageLocationID(ctx, uuid.Nil, "<invalid-uuid>")
		assert.Error(t, err, "invalid UUID length: 14")
	})

	t.Run("Returns if persistence failed", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()
		AIPID := "7c09fa45-cdac-4874-90af-56dc86a6e73c"
		locationID := uuid.MustParse("7484e911-7fc3-40c2-acb4-91e552d05380")

		attrs.persistenceMock.
			EXPECT().
			UpdatePackageLocationID(
				ctx,
				locationID,
				uuid.MustParse(AIPID),
			).
			Return(errors.New("something is wrong")).
			Times(1)

		err := svc.UpdatePackageLocationID(ctx, locationID, AIPID)
		assert.Error(t, err, "something is wrong")
	})
}

func TestServiceDelete(t *testing.T) {
	t.Parallel()

	t.Run("From internal location", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()
		AIPID := "7c09fa45-cdac-4874-90af-56dc86a6e73c"

		attrs.persistenceMock.
			EXPECT().
			ReadPackage(
				ctx,
				uuid.MustParse(AIPID),
			).
			Return(
				&goastorage.StoredStoragePackage{
					ID:         1,
					AipID:      AIPID,
					ObjectKey:  uuid.MustParse(AIPID),
					LocationID: &uuid.Nil,
				},
				nil,
			).
			Times(1)

		fakeLocationWithContents(t, svc, uuid.Nil, AIPID, "foobar")

		err := svc.Delete(ctx, AIPID)
		assert.NilError(t, err)
	})

	t.Run("From perma location", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()
		AIPID := "76a654ad-dccc-4dd3-a398-e84cd9f96415"
		locationID := uuid.MustParse("7484e911-7fc3-40c2-acb4-91e552d05380")

		attrs.persistenceMock.
			EXPECT().
			ReadPackage(
				ctx,
				uuid.MustParse(AIPID),
			).
			Return(
				&goastorage.StoredStoragePackage{
					ID:         1,
					AipID:      AIPID,
					ObjectKey:  uuid.MustParse(AIPID),
					LocationID: &locationID,
				},
				nil,
			).
			Times(1)

		fakeLocationWithContents(t, svc, locationID, AIPID, "foobar")

		err := svc.Delete(ctx, AIPID)
		assert.NilError(t, err)
	})

	// t.Run("Fails if object does not exist", func(t *testing.T) {
	// 	t.Parallel()

	// 	attrs := setUpAttrs{}
	// 	svc := setUpService(t, &attrs)
	// 	ctx := context.Background()
	// 	AIPID := "76a654ad-dccc-4dd3-a398-e84cd9f96415"
	// 	locationID := uuid.MustParse("7484e911-7fc3-40c2-acb4-91e552d05380")

	// 	attrs.persistenceMock.
	// 		EXPECT().
	// 		ReadPackage(
	// 			ctx,
	// 			uuid.MustParse(AIPID),
	// 		).
	// 		Return(
	// 			&goastorage.StoredStoragePackage{
	// 				ID:         1,
	// 				AipID:      AIPID,
	// 				ObjectKey:  uuid.MustParse(AIPID),
	// 				LocationID: &locationID,
	// 			},
	// 			nil,
	// 		).
	// 		Times(1)

	// 	// Fake empty location.
	// 	l, err := svc.Location(ctx, locationID)
	// 	assert.NilError(t, err)
	// 	mb := memblob.OpenBucket(nil)
	// 	l.SetBucket(mb)

	// 	err = svc.Delete(ctx, AIPID)
	// 	assert.Error(t, err, "blob (key \"76a654ad-dccc-4dd3-a398-e84cd9f96415\") (code=NotFound): blob not found")
	// })

	t.Run("Fails if location is not available", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()
		AIPID := "76a654ad-dccc-4dd3-a398-e84cd9f96415"
		locationID := uuid.MustParse("7484e911-7fc3-40c2-acb4-91e552d05380")

		attrs.persistenceMock.
			EXPECT().
			ReadPackage(
				ctx,
				uuid.MustParse(AIPID),
			).
			Return(
				&goastorage.StoredStoragePackage{
					ID:         1,
					AipID:      AIPID,
					ObjectKey:  uuid.MustParse(AIPID),
					LocationID: &locationID,
				},
				nil,
			).
			Times(1)

		err := svc.Delete(ctx, AIPID)
		assert.Error(t, err, "error loading location: unknown location 7484e911-7fc3-40c2-acb4-91e552d05380")
	})
}

func TestPackageReader(t *testing.T) {
	t.Parallel()

	t.Run("Provides a valid reader", func(t *testing.T) {
		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()
		AIPID := "7c09fa45-cdac-4874-90af-56dc86a6e73c"
		locationID := uuid.MustParse("7484e911-7fc3-40c2-acb4-91e552d05380")

		fakeLocationWithContents(t, svc, locationID, AIPID, "contents")

		attrs.persistenceMock.
			EXPECT().
			ReadLocation(
				ctx,
				locationID,
			).
			Return(
				&goastorage.StoredLocation{
					UUID: &locationID,
					Config: &goastorage.S3Config{
						Bucket: "perma-aips-1",
						Region: "planet-earth",
					},
				},
				nil,
			).
			Times(1)

		reader, err := svc.PackageReader(ctx, &goastorage.StoredStoragePackage{
			ID:         1,
			AipID:      AIPID,
			ObjectKey:  uuid.MustParse(AIPID),
			LocationID: &locationID,
		})
		assert.NilError(t, err)

		blob, err := io.ReadAll(reader)
		assert.NilError(t, err)
		assert.Equal(t, string(blob), "contents")
	})

	t.Run("Fails if the location is not available", func(t *testing.T) {
		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()
		AIPID := "7c09fa45-cdac-4874-90af-56dc86a6e73c"
		locationID := uuid.MustParse("7484e911-7fc3-40c2-acb4-91e552d05380")

		attrs.persistenceMock.
			EXPECT().
			ReadLocation(
				ctx,
				locationID,
			).
			Return(
				nil,
				errors.New("unknown location 7484e911-7fc3-40c2-acb4-91e552d05380"),
			).
			Times(1)

		_, err := svc.PackageReader(ctx, &goastorage.StoredStoragePackage{
			ID:         1,
			AipID:      AIPID,
			ObjectKey:  uuid.MustParse(AIPID),
			LocationID: &locationID,
		})
		assert.Error(t, err, "error loading location: unknown location 7484e911-7fc3-40c2-acb4-91e552d05380")
	})

	t.Run("Fails if the reader cannot be created", func(t *testing.T) {
		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()
		AIPID := "7c09fa45-cdac-4874-90af-56dc86a6e73c"
		locationID := uuid.MustParse("7484e911-7fc3-40c2-acb4-91e552d05380")

		mb := fakeLocationWithBucket(t, nil)

		attrs.persistenceMock.
			EXPECT().
			ReadLocation(
				ctx,
				locationID,
			).
			Return(
				&goastorage.StoredLocation{
					UUID: &locationID,
					Config: &goastorage.S3Config{
						Bucket: "perma-aips-1",
						Region: "planet-earth",
					},
				},
				nil,
			).
			Times(1)

		mb.Close() // Close the bucket beforehand to force the error.

		_, err := svc.PackageReader(ctx, &goastorage.StoredStoragePackage{
			ID:         1,
			AipID:      AIPID,
			ObjectKey:  uuid.MustParse(AIPID),
			LocationID: &locationID,
		})
		assert.Error(t, err, "blob: Bucket has been closed (code=FailedPrecondition)")
	})
}
