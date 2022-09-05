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
	temporalapi_enums "go.temporal.io/api/enums/v1"
	temporalapi_workflow "go.temporal.io/api/workflow/v1"
	temporalapi_workflowservice "go.temporal.io/api/workflowservice/v1"
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
	logger                  *logr.Logger
	config                  *storage.Config
	persistence             *persistence.Storage
	temporalClient          *temporalsdk_client.Client
	internalLocationFactory storage.InternalLocationFactory
	locationFactory         storage.LocationFactory

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
		persistence:             &ps,
		persistenceMock:         psMock,
		temporalClient:          &tc,
		temporalClientMock:      tcMock,
		internalLocationFactory: storage.DefaultInternalLocationFactory,
		locationFactory:         storage.DefaultLocationFactory,
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
	if attrs.internalLocationFactory != nil {
		params.internalLocationFactory = attrs.internalLocationFactory
	}
	if attrs.locationFactory != nil {
		params.locationFactory = attrs.locationFactory
	}

	*attrs = params

	s, err := storage.NewService(
		*params.logger,
		*params.config,
		*params.persistence,
		*params.temporalClient,
		params.internalLocationFactory,
		params.locationFactory,
	)
	assert.NilError(t, err)

	return s
}

type fakeLocation struct {
	b  *blob.Bucket
	id uuid.UUID
}

func (l *fakeLocation) UUID() uuid.UUID {
	return l.id
}

func (l *fakeLocation) Bucket() *blob.Bucket {
	return l.b
}

func (l *fakeLocation) Close() error {
	return nil
}

func fakeInternalLocationFactory(t *testing.T, b *blob.Bucket) storage.InternalLocationFactory {
	t.Helper()

	t.Cleanup(func() { b.Close() })

	if b == nil {
		b = memblob.OpenBucket(nil)
	}

	return func(config *storage.LocationConfig) (storage.Location, error) {
		return &fakeLocation{
			b:  b,
			id: uuid.Nil,
		}, nil
	}
}

func fakeInternalLocationFactoryWithContents(t *testing.T, b *blob.Bucket, objectKey, contents string) storage.InternalLocationFactory {
	t.Helper()

	if b == nil {
		b = memblob.OpenBucket(nil)
	}
	t.Cleanup(func() { b.Close() })
	b.WriteAll(context.Background(), objectKey, []byte(contents), nil)

	return func(config *storage.LocationConfig) (storage.Location, error) {
		return &fakeLocation{
			b:  b,
			id: uuid.Nil,
		}, nil
	}
}

func fakeLocationFactory(t *testing.T, b *blob.Bucket) storage.LocationFactory {
	t.Helper()

	t.Cleanup(func() { b.Close() })

	if b == nil {
		b = memblob.OpenBucket(nil)
	}

	return func(location *goastorage.Location) (storage.Location, error) {
		return &fakeLocation{
			b:  b,
			id: location.UUID,
		}, nil
	}
}

func fakeLocationFactoryWithContents(t *testing.T, b *blob.Bucket, objectKey, contents string) storage.LocationFactory {
	t.Helper()

	if b == nil {
		b = memblob.OpenBucket(nil)
	}
	t.Cleanup(func() { b.Close() })
	b.WriteAll(context.Background(), objectKey, []byte(contents), nil)

	return func(location *goastorage.Location) (storage.Location, error) {
		return &fakeLocation{
			b:  b,
			id: location.UUID,
		}, nil
	}
}

// io.Reader used as number generator for making UUIDs predictable
type staticRand struct{}

func (f staticRand) Read(buf []byte) (n int, err error) {
	for i := range buf {
		buf[i] = byte(i)
	}
	return len(buf), nil
}

func TestNewService(t *testing.T) {
	t.Parallel()

	_, err := storage.NewService(
		logr.Discard(),
		storage.Config{},
		nil,
		nil,
		storage.DefaultInternalLocationFactory,
		storage.DefaultLocationFactory,
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
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns not_available if workflow cannot be executed", func(t *testing.T) {
		t.Parallel()

		AIPID := uuid.MustParse("5ab42bc3-acc2-420b-bbd0-76efdef94828")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		attrs.temporalClientMock.
			On(
				"ExecuteWorkflow",
				mock.AnythingOfType("*context.timerCtx"),
				temporalsdk_client.StartWorkflowOptions{
					ID:                    "storage-upload-workflow-" + AIPID.String(),
					TaskQueue:             "global",
					WorkflowIDReusePolicy: temporalapi_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
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
			AipID: AIPID.String(),
		})
		assert.Assert(t, ret == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_available")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns not_valid if package cannot be persisted", func(t *testing.T) {
		t.Parallel()

		AIPID := uuid.MustParse("5ab42bc3-acc2-420b-bbd0-76efdef94828")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"ExecuteWorkflow",
				mock.AnythingOfType("*context.timerCtx"),
				temporalsdk_client.StartWorkflowOptions{
					ID:                    "storage-upload-workflow-" + AIPID.String(),
					TaskQueue:             "global",
					WorkflowIDReusePolicy: temporalapi_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
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
			AipID: AIPID.String(),
		})
		assert.Assert(t, ret == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot persist package")
	})

	t.Run("Returns not_valid if signed URL cannot be generated", func(t *testing.T) {
		t.Parallel()

		AIPID := uuid.MustParse("5ab42bc3-acc2-420b-bbd0-76efdef94828")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"ExecuteWorkflow",
				mock.AnythingOfType("*context.timerCtx"),
				temporalsdk_client.StartWorkflowOptions{
					ID:                    "storage-upload-workflow-" + AIPID.String(),
					TaskQueue:             "global",
					WorkflowIDReusePolicy: temporalapi_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
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
				&goastorage.Package{},
				nil,
			).
			Times(1)

		ret, err := svc.Submit(ctx, &goastorage.SubmitPayload{
			Name:  "package",
			AipID: AIPID.String(),
		})
		assert.Assert(t, ret == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot persist package")
	})

	t.Run("Returns signed URL", func(t *testing.T) {
		t.Parallel()

		// Fake internal location, using fileblob because it can generate signed URLs.
		furl, err := url.Parse("file:///tmp/dir")
		assert.NilError(t, err)
		b, err := fileblob.OpenBucket("/tmp", &fileblob.Options{URLSigner: fileblob.NewURLSignerHMAC(furl, []byte("1234"))})
		assert.NilError(t, err)

		AIPID := uuid.MustParse("5ab42bc3-acc2-420b-bbd0-76efdef94828")
		attrs := &setUpAttrs{
			internalLocationFactory: fakeInternalLocationFactory(t, b),
		}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"ExecuteWorkflow",
				mock.AnythingOfType("*context.timerCtx"),
				temporalsdk_client.StartWorkflowOptions{
					ID:                    "storage-upload-workflow-" + AIPID.String(),
					TaskQueue:             "global",
					WorkflowIDReusePolicy: temporalapi_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
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
				&goastorage.Package{},
				nil,
			).
			Times(1)

		ret, err := svc.Submit(ctx, &goastorage.SubmitPayload{
			Name:  "package",
			AipID: AIPID.String(),
		})
		assert.Equal(t, ret.URL[0:15], "file:///tmp/dir")
		assert.NilError(t, err)
	})
}

func TestServiceLocation(t *testing.T) {
	t.Parallel()

	attrs := &setUpAttrs{
		locationFactory: fakeLocationFactory(t, nil),
	}
	ctx := context.Background()
	svc := setUpService(t, attrs)
	locationID := uuid.MustParse("50110114-55ac-4567-b74f-9def601c6293")

	attrs.persistenceMock.
		EXPECT().
		ReadLocation(
			ctx,
			locationID,
		).
		Return(
			&goastorage.Location{
				UUID: locationID,
			},
			nil,
		).Times(1)
	attrs.persistenceMock.
		EXPECT().
		ReadLocation(
			ctx,
			uuid.MustParse("d8ea8946-dc82-4f4e-8c2d-8d3861f3297d"),
		).
		Return(
			nil,
			&goastorage.LocationNotFound{
				UUID:    uuid.MustParse("d8ea8946-dc82-4f4e-8c2d-8d3861f3297d"),
				Message: "location not found",
			},
		).Times(1)

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
			&goastorage.LocationNotFound{
				UUID:    uuid.MustParse("d8ea8946-dc82-4f4e-8c2d-8d3861f3297d"),
				Message: "location not found",
			},
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

		storedLocations := goastorage.LocationCollection{
			{
				Name:        "perma-aips-1",
				Description: ref.New("One"),
				Source:      "minio",
				Purpose:     "aip_store",
				UUID:        uuid.New(),
			},
			{
				Name:        "perma-aips-2",
				Description: ref.New("Two"),
				Source:      "minio",
				Purpose:     "aip_store",
				UUID:        uuid.New(),
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
	t.Parallel()

	t.Run("Returns not_valid if AIPID is invalid", func(t *testing.T) {
		t.Parallel()

		AIPID := "12345"
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		err := svc.Reject(context.Background(), &goastorage.RejectPayload{
			AipID: AIPID,
		})
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Fails when passing an invalid UUID", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()
		AIPID := uuid.MustParse("7c09fa45-cdac-4874-90af-56dc86a6e73c")

		attrs.persistenceMock.
			EXPECT().
			UpdatePackageStatus(
				ctx,
				AIPID,
				types.StatusRejected,
			).
			Return(nil).
			Times(1)

		err := svc.Reject(ctx, &goastorage.RejectPayload{AipID: AIPID.String()})
		assert.NilError(t, err)
	})
}

func TestServiceReadPackage(t *testing.T) {
	t.Parallel()

	attrs := setUpAttrs{}
	svc := setUpService(t, &attrs)
	ctx := context.Background()
	AIPID := uuid.MustParse("76a654ad-dccc-4dd3-a398-e84cd9f96415")

	attrs.persistenceMock.
		EXPECT().
		ReadPackage(
			ctx,
			AIPID,
		).
		Return(
			&goastorage.Package{},
			nil,
		).
		Times(1)

	pkg, err := svc.ReadPackage(ctx, AIPID)
	assert.NilError(t, err)
	assert.DeepEqual(t, pkg, &goastorage.Package{})
}

func TestServiceUpdatePackageStatus(t *testing.T) {
	t.Parallel()

	t.Run("Returns if persistence failed", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()
		AIPID := uuid.MustParse("7c09fa45-cdac-4874-90af-56dc86a6e73c")

		attrs.persistenceMock.
			EXPECT().
			UpdatePackageStatus(
				ctx,
				AIPID,
				types.StatusStored,
			).
			Return(errors.New("something is wrong")).
			Times(1)

		err := svc.UpdatePackageStatus(ctx, AIPID, types.StatusStored)
		assert.Error(t, err, "something is wrong")
	})
}

func TestServiceUpdatePackageLocationID(t *testing.T) {
	t.Parallel()

	t.Run("Returns if persistence failed", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()
		AIPID := uuid.MustParse("7c09fa45-cdac-4874-90af-56dc86a6e73c")
		locationID := uuid.MustParse("7484e911-7fc3-40c2-acb4-91e552d05380")

		attrs.persistenceMock.
			EXPECT().
			UpdatePackageLocationID(
				ctx,
				AIPID,
				locationID,
			).
			Return(errors.New("something is wrong")).
			Times(1)

		err := svc.UpdatePackageLocationID(ctx, AIPID, locationID)
		assert.Error(t, err, "something is wrong")
	})
}

func TestServiceDelete(t *testing.T) {
	t.Parallel()

	t.Run("From internal location", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		AIPID := uuid.MustParse("7c09fa45-cdac-4874-90af-56dc86a6e73c")
		attrs := setUpAttrs{
			internalLocationFactory: fakeInternalLocationFactoryWithContents(t, nil, AIPID.String(), "foobar"),
		}
		svc := setUpService(t, &attrs)

		attrs.persistenceMock.
			EXPECT().
			ReadPackage(
				ctx,
				AIPID,
			).
			Return(
				&goastorage.Package{
					AipID:      AIPID,
					ObjectKey:  AIPID,
					LocationID: &uuid.Nil,
				},
				nil,
			).
			Times(1)

		err := svc.Delete(ctx, AIPID)
		assert.NilError(t, err)
	})

	t.Run("From perma location", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		AIPID := uuid.MustParse("76a654ad-dccc-4dd3-a398-e84cd9f96415")
		locationID := uuid.MustParse("7484e911-7fc3-40c2-acb4-91e552d05380")
		attrs := setUpAttrs{
			locationFactory: fakeLocationFactoryWithContents(t, nil, AIPID.String(), "foobar"),
		}
		svc := setUpService(t, &attrs)

		attrs.persistenceMock.
			EXPECT().
			ReadPackage(
				ctx,
				AIPID,
			).
			Return(
				&goastorage.Package{
					AipID:      AIPID,
					ObjectKey:  AIPID,
					LocationID: &locationID,
				},
				nil,
			).
			Times(1)

		attrs.persistenceMock.
			EXPECT().
			ReadLocation(
				ctx,
				locationID,
			).
			Return(
				&goastorage.Location{
					UUID: locationID,
				},
				nil,
			).
			Times(1)

		err := svc.Delete(ctx, AIPID)
		assert.NilError(t, err)
	})

	t.Run("Fails if object does not exist", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		AIPID := uuid.MustParse("76a654ad-dccc-4dd3-a398-e84cd9f96415")
		locationID := uuid.MustParse("7484e911-7fc3-40c2-acb4-91e552d05380")
		attrs := setUpAttrs{
			locationFactory: fakeLocationFactory(t, nil),
		}
		svc := setUpService(t, &attrs)

		attrs.persistenceMock.
			EXPECT().
			ReadPackage(
				ctx,
				AIPID,
			).
			Return(
				&goastorage.Package{
					AipID:      AIPID,
					ObjectKey:  AIPID,
					LocationID: &locationID,
				},
				nil,
			).
			Times(1)

		attrs.persistenceMock.
			EXPECT().
			ReadLocation(
				ctx,
				locationID,
			).
			Return(
				&goastorage.Location{
					UUID: locationID,
				},
				nil,
			).
			Times(1)

		err := svc.Delete(ctx, AIPID)
		assert.Error(t, err, "blob (key \"76a654ad-dccc-4dd3-a398-e84cd9f96415\") (code=NotFound): blob not found")
	})

	t.Run("Fails if location is not available", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()
		AIPID := uuid.MustParse("76a654ad-dccc-4dd3-a398-e84cd9f96415")
		locationID := uuid.MustParse("7484e911-7fc3-40c2-acb4-91e552d05380")

		attrs.persistenceMock.
			EXPECT().
			ReadPackage(
				ctx,
				AIPID,
			).
			Return(
				&goastorage.Package{
					AipID:      AIPID,
					ObjectKey:  AIPID,
					LocationID: &locationID,
				},
				nil,
			).
			Times(1)

		attrs.persistenceMock.
			EXPECT().
			ReadLocation(
				ctx,
				locationID,
			).
			Return(
				nil,
				&goastorage.LocationNotFound{UUID: locationID, Message: "location not found"},
			).
			Times(1)

		err := svc.Delete(ctx, AIPID)
		assert.ErrorContains(t, err, "location not found")
	})

	t.Run("Fails if package does not exist", func(t *testing.T) {
		t.Parallel()

		AIPID := uuid.MustParse("76a654ad-dccc-4dd3-a398-e84cd9f96415")
		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			ReadPackage(
				ctx,
				AIPID,
			).
			Return(
				nil,
				&goastorage.PackageNotFound{AipID: AIPID, Message: "package not found"},
			).
			Times(1)

		err := svc.Delete(ctx, AIPID)
		assert.ErrorContains(t, err, "package not found")
	})
}

func TestPackageReader(t *testing.T) {
	t.Parallel()

	t.Run("Provides a valid reader", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		AIPID := uuid.MustParse("7c09fa45-cdac-4874-90af-56dc86a6e73c")
		locationID := uuid.MustParse("7484e911-7fc3-40c2-acb4-91e552d05380")
		attrs := setUpAttrs{
			locationFactory: fakeLocationFactoryWithContents(t, nil, AIPID.String(), "contents"),
		}
		svc := setUpService(t, &attrs)

		attrs.persistenceMock.
			EXPECT().
			ReadLocation(
				ctx,
				locationID,
			).
			Return(
				&goastorage.Location{
					UUID: locationID,
					Config: &goastorage.S3Config{
						Bucket: "perma-aips-1",
						Region: "planet-earth",
					},
				},
				nil,
			).
			Times(1)

		reader, err := svc.PackageReader(ctx, &goastorage.Package{
			AipID:      AIPID,
			ObjectKey:  AIPID,
			LocationID: &locationID,
		})
		assert.NilError(t, err)

		blob, err := io.ReadAll(reader)
		assert.NilError(t, err)
		assert.Equal(t, string(blob), "contents")
	})

	t.Run("Fails if the location is not available", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()
		AIPID := uuid.MustParse("7c09fa45-cdac-4874-90af-56dc86a6e73c")
		locationID := uuid.MustParse("7484e911-7fc3-40c2-acb4-91e552d05380")

		attrs.persistenceMock.
			EXPECT().
			ReadLocation(
				ctx,
				locationID,
			).
			Return(
				nil,
				&goastorage.LocationNotFound{UUID: locationID, Message: "location not found"},
			).
			Times(1)

		_, err := svc.PackageReader(ctx, &goastorage.Package{
			AipID:      AIPID,
			ObjectKey:  AIPID,
			LocationID: &locationID,
		})
		assert.ErrorContains(t, err, "location not found")
	})

	t.Run("Fails if the reader cannot be created", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		AIPID := uuid.MustParse("7c09fa45-cdac-4874-90af-56dc86a6e73c")
		locationID := uuid.MustParse("7484e911-7fc3-40c2-acb4-91e552d05380")
		b := memblob.OpenBucket(nil)
		attrs := setUpAttrs{
			locationFactory: fakeLocationFactoryWithContents(t, b, AIPID.String(), "foobar"),
		}
		svc := setUpService(t, &attrs)
		b.Close()

		attrs.persistenceMock.
			EXPECT().
			ReadLocation(
				ctx,
				locationID,
			).
			Return(
				&goastorage.Location{
					UUID: locationID,
					Config: &goastorage.S3Config{
						Bucket: "perma-aips-1",
						Region: "planet-earth",
					},
				},
				nil,
			).
			Times(1)

		_, err := svc.PackageReader(ctx, &goastorage.Package{
			AipID:      AIPID,
			ObjectKey:  AIPID,
			LocationID: &locationID,
		})
		assert.Error(t, err, "blob: Bucket has been closed (code=FailedPrecondition)")
	})
}

func TestServiceUpdate(t *testing.T) {
	t.Parallel()

	t.Run("Returns not_valid if AIPID is invalid", func(t *testing.T) {
		t.Parallel()

		AIPID := "12345"
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		err := svc.Update(context.Background(), &goastorage.UpdatePayload{
			AipID: AIPID,
		})
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns not_available if workflow cannot be signaled", func(t *testing.T) {
		t.Parallel()

		AIPID := uuid.MustParse("5ab42bc3-acc2-420b-bbd0-76efdef94828")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"SignalWorkflow",
				ctx,
				"storage-upload-workflow-"+AIPID.String(),
				"",
				"upload-done-signal",
				storage.UploadDoneSignal{},
			).
			Return(
				errors.New("something went wrong"),
			).
			Times(1)

		err := svc.Update(ctx, &goastorage.UpdatePayload{
			AipID: AIPID.String(),
		})
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_available")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns not_valid if package cannot be updated", func(t *testing.T) {
		t.Parallel()

		AIPID := uuid.MustParse("5ab42bc3-acc2-420b-bbd0-76efdef94828")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"SignalWorkflow",
				ctx,
				"storage-upload-workflow-"+AIPID.String(),
				"",
				"upload-done-signal",
				storage.UploadDoneSignal{},
			).
			Return(
				nil,
			).
			Times(1)

		attrs.persistenceMock.
			EXPECT().
			UpdatePackageStatus(
				gomock.AssignableToTypeOf(ctx),
				gomock.Any(),
				types.StatusInReview,
			).
			Return(
				errors.New("unexpected error"),
			).
			Times(1)

		err := svc.Update(ctx, &goastorage.UpdatePayload{
			AipID: AIPID.String(),
		})
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot persist package")
	})

	t.Run("Returns no error if package is updated", func(t *testing.T) {
		t.Parallel()

		AIPID := uuid.MustParse("5ab42bc3-acc2-420b-bbd0-76efdef94828")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"SignalWorkflow",
				ctx,
				"storage-upload-workflow-"+AIPID.String(),
				"",
				"upload-done-signal",
				storage.UploadDoneSignal{},
			).
			Return(
				nil,
			).
			Times(1)

		attrs.persistenceMock.
			EXPECT().
			UpdatePackageStatus(
				gomock.AssignableToTypeOf(ctx),
				gomock.Any(),
				types.StatusInReview,
			).
			Return(
				nil,
			).
			Times(1)

		err := svc.Update(ctx, &goastorage.UpdatePayload{
			AipID: AIPID.String(),
		})
		assert.NilError(t, err)
	})
}

func TestServiceMove(t *testing.T) {
	t.Parallel()

	t.Run("Returns not_valid if AIPID is invalid", func(t *testing.T) {
		t.Parallel()

		AIPID := "12345"
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		err := svc.Move(context.Background(), &goastorage.MovePayload{
			AipID: AIPID,
		})
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns not found error if package does not exist", func(t *testing.T) {
		t.Parallel()

		AIPID := uuid.MustParse("5ab42bc3-acc2-420b-bbd0-76efdef94828")
		LocationID := uuid.MustParse("4b15a34a-f765-407d-a811-7248d2d2af66")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			ReadPackage(
				gomock.AssignableToTypeOf(ctx),
				AIPID,
			).
			Return(
				nil,
				&goastorage.PackageNotFound{AipID: AIPID, Message: "package not found"},
			).
			Times(1)

		err := svc.Move(ctx, &goastorage.MovePayload{
			AipID:      AIPID.String(),
			LocationID: LocationID,
		})
		assert.ErrorContains(t, err, "package not found")
	})

	t.Run("Returns not_available if workflow cannot be executed", func(t *testing.T) {
		t.Parallel()

		AIPID := uuid.MustParse("5ab42bc3-acc2-420b-bbd0-76efdef94828")
		LocationID := uuid.MustParse("4b15a34a-f765-407d-a811-7248d2d2af66")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"ExecuteWorkflow",
				mock.AnythingOfType("*context.timerCtx"),
				temporalsdk_client.StartWorkflowOptions{
					ID:                    "storage-move-workflow-" + AIPID.String(),
					TaskQueue:             "global",
					WorkflowIDReusePolicy: temporalapi_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
				},
				"storage-move-workflow",
				&storage.StorageMoveWorkflowRequest{AIPID: AIPID, LocationID: LocationID},
			).
			Return(
				nil,
				errors.New("something went wrong"),
			).
			Times(1)

		attrs.persistenceMock.
			EXPECT().
			ReadPackage(
				gomock.AssignableToTypeOf(ctx),
				AIPID,
			).
			Return(
				&goastorage.Package{AipID: AIPID},
				nil,
			).
			Times(1)

		err := svc.Move(ctx, &goastorage.MovePayload{
			AipID:      AIPID.String(),
			LocationID: LocationID,
		})
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_available")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns no error if package is moved", func(t *testing.T) {
		t.Parallel()

		AIPID := uuid.MustParse("5ab42bc3-acc2-420b-bbd0-76efdef94828")
		LocationID := uuid.MustParse("4b15a34a-f765-407d-a811-7248d2d2af66")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"ExecuteWorkflow",
				mock.AnythingOfType("*context.timerCtx"),
				temporalsdk_client.StartWorkflowOptions{
					ID:                    "storage-move-workflow-" + AIPID.String(),
					TaskQueue:             "global",
					WorkflowIDReusePolicy: temporalapi_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
				},
				"storage-move-workflow",
				&storage.StorageMoveWorkflowRequest{AIPID: AIPID, LocationID: LocationID},
			).
			Return(
				&temporalsdk_mocks.WorkflowRun{},
				nil,
			).
			Times(1)

		attrs.persistenceMock.
			EXPECT().
			ReadPackage(
				gomock.AssignableToTypeOf(ctx),
				AIPID,
			).
			Return(
				&goastorage.Package{AipID: AIPID},
				nil,
			).
			Times(1)

		err := svc.Move(ctx, &goastorage.MovePayload{
			AipID:      AIPID.String(),
			LocationID: LocationID,
		})
		assert.NilError(t, err)
	})
}

func TestServiceMoveStatus(t *testing.T) {
	t.Parallel()

	t.Run("Returns not_valid if AIPID is invalid", func(t *testing.T) {
		t.Parallel()

		AIPID := "12345"
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		res, err := svc.MoveStatus(context.Background(), &goastorage.MoveStatusPayload{
			AipID: AIPID,
		})
		assert.Assert(t, res == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns not found error if package does not exist", func(t *testing.T) {
		t.Parallel()

		AIPID := uuid.MustParse("5ab42bc3-acc2-420b-bbd0-76efdef94828")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			ReadPackage(
				gomock.AssignableToTypeOf(ctx),
				AIPID,
			).
			Return(
				nil,
				&goastorage.PackageNotFound{AipID: AIPID, Message: "package not found"},
			).
			Times(1)

		res, err := svc.MoveStatus(ctx, &goastorage.MoveStatusPayload{
			AipID: AIPID.String(),
		})
		assert.Assert(t, res == nil)
		assert.ErrorContains(t, err, "package not found")
	})

	t.Run("Returns failed_dependency error if workflow execution cannot be found", func(t *testing.T) {
		t.Parallel()

		AIPID := uuid.MustParse("5ab42bc3-acc2-420b-bbd0-76efdef94828")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"DescribeWorkflowExecution",
				ctx,
				"storage-move-workflow-"+AIPID.String(),
				"",
			).
			Return(
				nil,
				errors.New("something went wrong"),
			).
			Times(1)

		attrs.persistenceMock.
			EXPECT().
			ReadPackage(
				gomock.AssignableToTypeOf(ctx),
				AIPID,
			).
			Return(
				&goastorage.Package{AipID: AIPID},
				nil,
			).
			Times(1)

		res, err := svc.MoveStatus(ctx, &goastorage.MoveStatusPayload{
			AipID: AIPID.String(),
		})
		assert.Assert(t, res == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "failed_dependency")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns failed_dependency error if workflow execution failed", func(t *testing.T) {
		t.Parallel()

		AIPID := uuid.MustParse("5ab42bc3-acc2-420b-bbd0-76efdef94828")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"DescribeWorkflowExecution",
				ctx,
				"storage-move-workflow-"+AIPID.String(),
				"",
			).
			Return(
				&temporalapi_workflowservice.DescribeWorkflowExecutionResponse{
					WorkflowExecutionInfo: &temporalapi_workflow.WorkflowExecutionInfo{
						Status: temporalapi_enums.WORKFLOW_EXECUTION_STATUS_FAILED,
					},
				},
				nil,
			).
			Times(1)

		attrs.persistenceMock.
			EXPECT().
			ReadPackage(
				gomock.AssignableToTypeOf(ctx),
				AIPID,
			).
			Return(
				&goastorage.Package{AipID: AIPID},
				nil,
			).
			Times(1)

		res, err := svc.MoveStatus(ctx, &goastorage.MoveStatusPayload{
			AipID: AIPID.String(),
		})
		assert.Assert(t, res == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "failed_dependency")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns move not done if workflow is running", func(t *testing.T) {
		t.Parallel()

		AIPID := uuid.MustParse("5ab42bc3-acc2-420b-bbd0-76efdef94828")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"DescribeWorkflowExecution",
				ctx,
				"storage-move-workflow-"+AIPID.String(),
				"",
			).
			Return(
				&temporalapi_workflowservice.DescribeWorkflowExecutionResponse{
					WorkflowExecutionInfo: &temporalapi_workflow.WorkflowExecutionInfo{
						Status: temporalapi_enums.WORKFLOW_EXECUTION_STATUS_RUNNING,
					},
				},
				nil,
			).
			Times(1)

		attrs.persistenceMock.
			EXPECT().
			ReadPackage(
				gomock.AssignableToTypeOf(ctx),
				AIPID,
			).
			Return(
				&goastorage.Package{AipID: AIPID},
				nil,
			).
			Times(1)

		res, err := svc.MoveStatus(ctx, &goastorage.MoveStatusPayload{
			AipID: AIPID.String(),
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, res, &goastorage.MoveStatusResult{Done: false})
	})

	t.Run("Returns move done if workflow completed", func(t *testing.T) {
		t.Parallel()

		AIPID := uuid.MustParse("5ab42bc3-acc2-420b-bbd0-76efdef94828")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"DescribeWorkflowExecution",
				ctx,
				"storage-move-workflow-"+AIPID.String(),
				"",
			).
			Return(
				&temporalapi_workflowservice.DescribeWorkflowExecutionResponse{
					WorkflowExecutionInfo: &temporalapi_workflow.WorkflowExecutionInfo{
						Status: temporalapi_enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
					},
				},
				nil,
			).
			Times(1)

		attrs.persistenceMock.
			EXPECT().
			ReadPackage(
				gomock.AssignableToTypeOf(ctx),
				AIPID,
			).
			Return(
				&goastorage.Package{AipID: AIPID},
				nil,
			).
			Times(1)

		res, err := svc.MoveStatus(ctx, &goastorage.MoveStatusPayload{
			AipID: AIPID.String(),
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, res, &goastorage.MoveStatusResult{Done: true})
	})
}

func TestServiceAddLocation(t *testing.T) {
	t.Parallel()

	t.Run("Returns error if unsupported configuration type", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		res, err := svc.AddLocation(ctx, &goastorage.AddLocationPayload{
			Name:    "perma-aips-1",
			Source:  types.LocationSourceMinIO.String(),
			Purpose: types.LocationPurposeAIPStore.String(),
			Config:  nil,
		})
		assert.Assert(t, res == nil)
		assert.ErrorContains(t, err, "unsupported config type")
	})

	t.Run("Returns error if configuration is invalid", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		res, err := svc.AddLocation(ctx, &goastorage.AddLocationPayload{
			Name:    "perma-aips-1",
			Source:  types.LocationSourceMinIO.String(),
			Purpose: types.LocationPurposeAIPStore.String(),
			Config:  &goastorage.S3Config{},
		})
		assert.Assert(t, res == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "invalid configuration")
	})

	t.Run("Returns not_valid if cannot persist location", func(t *testing.T) {
		t.Cleanup(func() { uuid.SetRand(nil) })

		uuid.SetRand(staticRand{})
		locationID := uuid.MustParse("00010203-0405-4607-8809-0a0b0c0d0e0f")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			CreateLocation(
				gomock.AssignableToTypeOf(ctx),
				&goastorage.Location{
					Name:    "perma-aips-1",
					Source:  types.LocationSourceMinIO.String(),
					Purpose: types.LocationPurposeAIPStore.String(),
					UUID:    locationID,
				},
				&types.LocationConfig{
					Value: &types.S3Config{
						Bucket: "perma-aips-1",
						Region: "planet-earth",
					},
				},
			).
			Return(
				nil,
				errors.New("unexpected error"),
			).
			Times(1)

		res, err := svc.AddLocation(ctx, &goastorage.AddLocationPayload{
			Name:    "perma-aips-1",
			Source:  types.LocationSourceMinIO.String(),
			Purpose: types.LocationPurposeAIPStore.String(),
			Config: &goastorage.S3Config{
				Bucket: "perma-aips-1",
				Region: "planet-earth",
			},
		})
		assert.Assert(t, res == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot persist location")
	})

	t.Run("Returns result with location UUID", func(t *testing.T) {
		t.Cleanup(func() { uuid.SetRand(nil) })

		uuid.SetRand(staticRand{})
		locationID := uuid.MustParse("00010203-0405-4607-8809-0a0b0c0d0e0f")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			CreateLocation(
				gomock.AssignableToTypeOf(ctx),
				&goastorage.Location{
					Name:    "perma-aips-1",
					Source:  types.LocationSourceMinIO.String(),
					Purpose: types.LocationPurposeAIPStore.String(),
					UUID:    locationID,
				},
				&types.LocationConfig{
					Value: &types.S3Config{
						Bucket: "perma-aips-1",
						Region: "planet-earth",
					},
				},
			).
			Return(
				&goastorage.Location{},
				nil,
			).
			Times(1)

		res, err := svc.AddLocation(ctx, &goastorage.AddLocationPayload{
			Name:    "perma-aips-1",
			Source:  types.LocationSourceMinIO.String(),
			Purpose: types.LocationPurposeAIPStore.String(),
			Config: &goastorage.S3Config{
				Bucket: "perma-aips-1",
				Region: "planet-earth",
			},
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, res, &goastorage.AddLocationResult{UUID: locationID.String()})
	})
}

func TestServiceShowLocation(t *testing.T) {
	t.Parallel()

	t.Run("Returns not_valid if cannot parse location UUID", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		res, err := svc.ShowLocation(ctx, &goastorage.ShowLocationPayload{
			UUID: "hello world",
		})
		assert.Assert(t, res == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns stored location", func(t *testing.T) {
		t.Parallel()

		locationID := uuid.MustParse("c145d0b3-5ad6-4fa2-b8ec-7b66bc353241")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			ReadLocation(
				ctx,
				locationID,
			).
			Return(
				&goastorage.Location{
					UUID: locationID,
				},
				nil,
			).Times(1)

		res, err := svc.ShowLocation(ctx, &goastorage.ShowLocationPayload{
			UUID: locationID.String(),
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, res, &goastorage.Location{UUID: locationID})
	})
}

func TestServiceLocationPackages(t *testing.T) {
	t.Parallel()

	t.Run("Returns not_valid if cannot parse location UUID", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		res, err := svc.LocationPackages(ctx, &goastorage.LocationPackagesPayload{
			UUID: "hello world",
		})
		assert.Assert(t, res == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns not_available if packages cannot be read", func(t *testing.T) {
		t.Parallel()

		locationID := uuid.MustParse("c145d0b3-5ad6-4fa2-b8ec-7b66bc353241")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			LocationPackages(
				ctx,
				locationID,
			).
			Return(
				nil,
				errors.New("unexpected error"),
			).Times(1)

		res, err := svc.LocationPackages(ctx, &goastorage.LocationPackagesPayload{
			UUID: locationID.String(),
		})
		assert.Assert(t, res == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_available")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns stored packages", func(t *testing.T) {
		t.Parallel()

		locationID := uuid.MustParse("c145d0b3-5ad6-4fa2-b8ec-7b66bc353241")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			LocationPackages(
				ctx,
				locationID,
			).
			Return(
				goastorage.PackageCollection{
					{
						Name:       "Package",
						AipID:      uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e"),
						Status:     "stored",
						ObjectKey:  uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a"),
						LocationID: ref.New(locationID),
						CreatedAt:  "2013-02-03T19:54:00Z",
					},
				},
				nil,
			).Times(1)

		res, err := svc.LocationPackages(ctx, &goastorage.LocationPackagesPayload{
			UUID: locationID.String(),
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, res, goastorage.PackageCollection{
			{
				Name:       "Package",
				AipID:      uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e"),
				Status:     "stored",
				ObjectKey:  uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a"),
				LocationID: ref.New(locationID),
				CreatedAt:  "2013-02-03T19:54:00Z",
			},
		})
	})
}

func TestServiceShow(t *testing.T) {
	t.Parallel()

	t.Run("Returns not_valid if AIPID is invalid", func(t *testing.T) {
		t.Parallel()

		AIPID := "12345"
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		res, err := svc.Show(context.Background(), &goastorage.ShowPayload{
			AipID: AIPID,
		})
		assert.Assert(t, res == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns stored package", func(t *testing.T) {
		t.Parallel()

		AIPID := uuid.MustParse("9a8f43de-9e1c-4313-aaaa-c694ebe0d45f")
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			ReadPackage(
				ctx,
				AIPID,
			).
			Return(
				&goastorage.Package{
					AipID:      AIPID,
					ObjectKey:  AIPID,
					LocationID: &uuid.Nil,
				},
				nil,
			).
			Times(1)

		res, err := svc.Show(ctx, &goastorage.ShowPayload{
			AipID: AIPID.String(),
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, res, &goastorage.Package{
			AipID:      AIPID,
			ObjectKey:  AIPID,
			LocationID: &uuid.Nil,
		})
	})
}
