package storage_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"go.artefactual.dev/tools/mockutil"
	"go.artefactual.dev/tools/ref"
	temporalapi_enums "go.temporal.io/api/enums/v1"
	temporalapi_workflow "go.temporal.io/api/workflow/v1"
	temporalapi_workflowservice "go.temporal.io/api/workflowservice/v1"
	temporalsdk_client "go.temporal.io/sdk/client"
	temporalsdk_mocks "go.temporal.io/sdk/mocks"
	"go.uber.org/mock/gomock"
	goa "goa.design/goa/v3/pkg"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/fake"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

var (
	aipID      = uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e")
	locationID = uuid.MustParse("a06a155c-9cf0-4416-a2b6-e90e58ef3186")
	objectKey  = uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a")
	uuid0      = uuid.MustParse("52fdfc07-2182-454f-963f-5f0f9a621d72")
)

type setUpAttrs struct {
	logger         *logr.Logger
	config         *storage.Config
	persistence    *persistence.Storage
	temporalClient *temporalsdk_client.Client

	persistenceMock    *fake.MockStorage
	temporalClientMock *temporalsdk_mocks.Client
	tokenVerifier      auth.TokenVerifier
	ticketProvider     auth.TicketProvider
}

func setUpService(t *testing.T, attrs *setUpAttrs) storage.Service {
	t.Helper()

	psMock := fake.NewMockStorage(gomock.NewController(t))
	var ps persistence.Storage = psMock

	tcMock := &temporalsdk_mocks.Client{}
	var tc temporalsdk_client.Client = tcMock

	td := tfs.NewDir(t, "enduro-service-test")

	params := setUpAttrs{
		logger: ref.New(logr.Discard()),
		config: &storage.Config{
			TaskQueue: "global",
			Internal: storage.LocationConfig{
				URL: "file://" + td.Path(),
			},
		},
		persistence:        &ps,
		persistenceMock:    psMock,
		temporalClient:     &tc,
		temporalClientMock: tcMock,
		tokenVerifier:      &auth.OIDCTokenVerifier{},
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
	if attrs.tokenVerifier != nil {
		params.tokenVerifier = attrs.tokenVerifier
	}
	if attrs.ticketProvider != nil {
		params.ticketProvider = attrs.ticketProvider
	}

	*attrs = params

	s, err := storage.NewService(
		*params.logger,
		*params.config,
		*params.persistence,
		*params.temporalClient,
		event.NewStorageEventServiceNop(),
		params.tokenVerifier,
		params.ticketProvider,
		rand.New(rand.NewSource(1)), // #nosec: G404
	)
	assert.NilError(t, err)

	return s
}

// writeTestBlob writes a test blob with the given key to the bucket at urlstr.
func writeTestBlob(ctx context.Context, t *testing.T, urlstr, key string) {
	t.Helper()

	b, err := blob.OpenBucket(ctx, urlstr)
	assert.NilError(t, err)
	defer b.Close()
	err = b.WriteAll(ctx, key, []byte("Testing 1-2-3!"), nil)
	assert.NilError(t, err)
}

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("Errors on invalid configuration", func(t *testing.T) {
		t.Parallel()

		_, err := storage.NewService(
			logr.Discard(),
			storage.Config{},
			nil,
			nil,
			event.NewStorageEventServiceNop(),
			&auth.OIDCTokenVerifier{},
			nil,
			nil,
		)

		assert.ErrorContains(t, err, "invalid configuration")
	})
}

func TestServiceSubmit(t *testing.T) {
	t.Parallel()

	t.Run("Returns not_valid if AIP ID is invalid", func(t *testing.T) {
		t.Parallel()

		aipID := "12345"
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		ret, err := svc.SubmitAip(context.Background(), &goastorage.SubmitAipPayload{
			UUID: aipID,
		})
		assert.Assert(t, ret == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns not_available if workflow cannot be executed", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		attrs.temporalClientMock.
			On(
				"ExecuteWorkflow",
				mock.AnythingOfType("*context.timerCtx"),
				temporalsdk_client.StartWorkflowOptions{
					ID:                    "storage-upload-workflow-" + aipID.String(),
					TaskQueue:             "global",
					WorkflowIDReusePolicy: temporalapi_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
				},
				"storage-upload-workflow",
				&storage.StorageUploadWorkflowRequest{
					AIPID:     aipID,
					TaskQueue: "global",
				},
			).
			Return(
				nil,
				errors.New("something went wrong"),
			)

		ret, err := svc.SubmitAip(context.Background(), &goastorage.SubmitAipPayload{
			UUID: aipID.String(),
		})
		assert.Assert(t, ret == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_available")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns not_valid if AIP cannot be persisted", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()
		attrs.temporalClientMock.
			On(
				"ExecuteWorkflow",
				mock.AnythingOfType("*context.timerCtx"),
				temporalsdk_client.StartWorkflowOptions{
					ID:                    "storage-upload-workflow-" + aipID.String(),
					TaskQueue:             "global",
					WorkflowIDReusePolicy: temporalapi_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
				},
				"storage-upload-workflow",
				&storage.StorageUploadWorkflowRequest{
					AIPID:     aipID,
					TaskQueue: "global",
				},
			).
			Return(
				&temporalsdk_mocks.WorkflowRun{},
				nil,
			)

		attrs.persistenceMock.
			EXPECT().
			CreateAIP(
				gomock.AssignableToTypeOf(ctx),
				gomock.Any(),
			).
			Return(
				nil,
				errors.New("database server error"),
			)

		ret, err := svc.SubmitAip(ctx, &goastorage.SubmitAipPayload{
			Name: "AIP",
			UUID: aipID.String(),
		})
		assert.Assert(t, ret == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot create AIP")
	})

	t.Run("Returns not_valid if signed URL cannot be generated", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"ExecuteWorkflow",
				mock.AnythingOfType("*context.timerCtx"),
				temporalsdk_client.StartWorkflowOptions{
					ID:                    "storage-upload-workflow-" + aipID.String(),
					TaskQueue:             "global",
					WorkflowIDReusePolicy: temporalapi_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
				},
				"storage-upload-workflow",
				&storage.StorageUploadWorkflowRequest{
					AIPID:     aipID,
					TaskQueue: "global",
				},
			).
			Return(
				&temporalsdk_mocks.WorkflowRun{},
				nil,
			)

		attrs.persistenceMock.
			EXPECT().
			CreateAIP(
				gomock.AssignableToTypeOf(ctx),
				gomock.Any(),
			).
			Return(
				&goastorage.AIP{},
				nil,
			)

		ret, err := svc.SubmitAip(ctx, &goastorage.SubmitAipPayload{
			Name: "AIP",
			UUID: aipID.String(),
		})
		assert.Assert(t, ret == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot sign URL")
	})

	t.Run("Returns signed URL", func(t *testing.T) {
		t.Parallel()

		td := tfs.NewDir(t, "enduro-service-test")

		attrs := setUpAttrs{
			config: &storage.Config{
				TaskQueue: "global",
				Internal: storage.LocationConfig{
					URL: fmt.Sprintf(
						"file://%s?base_url=file://tmp/dir&secret_key_path=fake/signing.key",
						td.Path(),
					),
				},
			},
		}
		svc := setUpService(t, &attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"ExecuteWorkflow",
				mock.AnythingOfType("*context.timerCtx"),
				temporalsdk_client.StartWorkflowOptions{
					ID:                    "storage-upload-workflow-" + aipID.String(),
					TaskQueue:             "global",
					WorkflowIDReusePolicy: temporalapi_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
				},
				"storage-upload-workflow",
				&storage.StorageUploadWorkflowRequest{
					AIPID:     aipID,
					TaskQueue: "global",
				},
			).
			Return(
				&temporalsdk_mocks.WorkflowRun{},
				nil,
			)

		attrs.persistenceMock.
			EXPECT().
			CreateAIP(
				gomock.AssignableToTypeOf(ctx),
				gomock.Any(),
			).
			Return(
				&goastorage.AIP{},
				nil,
			)

		ret, err := svc.SubmitAip(ctx, &goastorage.SubmitAipPayload{
			Name: "AIP",
			UUID: aipID.String(),
		})
		assert.NilError(t, err)
		assert.Equal(t, ret.URL[0:15], "file://tmp/dir?")
	})
}

func TestServiceCreate(t *testing.T) {
	t.Parallel()

	t.Run("Creates a new AIP", func(t *testing.T) {
		t.Parallel()

		name := "AIP 1"
		status := "stored"
		created := time.Date(2024, 5, 3, 14, 55, 2, 22, time.UTC)

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)

		attrs.persistenceMock.
			EXPECT().
			CreateAIP(
				mockutil.Context(),
				&goastorage.AIP{
					Name:         name,
					UUID:         aipID,
					Status:       status,
					ObjectKey:    objectKey,
					LocationUUID: &locationID,
				},
			).
			Return(
				&goastorage.AIP{
					Name:         name,
					UUID:         aipID,
					Status:       status,
					ObjectKey:    objectKey,
					LocationUUID: &locationID,
					CreatedAt:    created.Format(time.DateTime),
				},
				nil,
			)

		got, err := svc.CreateAip(context.Background(), &goastorage.CreateAipPayload{
			UUID:         aipID.String(),
			Name:         name,
			ObjectKey:    objectKey.String(),
			Status:       status,
			LocationUUID: &locationID,
		})

		assert.NilError(t, err)
		assert.DeepEqual(t, got, &goastorage.AIP{
			UUID:         aipID,
			Name:         name,
			ObjectKey:    objectKey,
			Status:       status,
			LocationUUID: &locationID,
			CreatedAt:    created.Format(time.DateTime),
		})
	})

	t.Run("Returns not_valid if AIP ID is invalid", func(t *testing.T) {
		t.Parallel()

		aipID := "12345"
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		ret, err := svc.CreateAip(context.Background(), &goastorage.CreateAipPayload{
			UUID: aipID,
		})
		assert.Assert(t, ret == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.Error(t, err, "invalid aip_id")
	})

	t.Run("Returns not_valid if ObjectKey is invalid", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		ret, err := svc.CreateAip(context.Background(), &goastorage.CreateAipPayload{
			UUID:      "f5fddd8c-570b-48d3-8426-78c03f24fa78",
			ObjectKey: "12345",
		})
		assert.Assert(t, ret == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.Error(t, err, "invalid object_key")
	})
}

func TestServiceLocation(t *testing.T) {
	t.Parallel()

	var attrs setUpAttrs
	ctx := context.Background()
	svc := setUpService(t, &attrs)
	locID2 := uuid.MustParse("d8ea8946-dc82-4f4e-8c2d-8d3861f3297d")

	attrs.persistenceMock.
		EXPECT().
		ReadLocation(
			ctx,
			locationID,
		).
		Return(
			&goastorage.Location{
				UUID: locationID,
				Config: &goastorage.URLConfig{
					URL: "mem://",
				},
			},
			nil,
		)
	attrs.persistenceMock.
		EXPECT().
		ReadLocation(ctx, locID2).
		Return(
			nil,
			&goastorage.LocationNotFound{
				UUID:    locID2,
				Message: "location not found",
			},
		)

	testCases := map[string]struct {
		UUID uuid.UUID
		err  error
	}{
		"Returns internal location": {
			uuid.Nil,
			nil,
		},
		"Returns location": {
			locationID,
			nil,
		},
		"Returns error when location cannot be found": {
			locID2,
			&goastorage.LocationNotFound{
				UUID:    locID2,
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
			Return(storedLocations, nil)

		res, err := svc.ListLocations(ctx, &goastorage.ListLocationsPayload{})
		assert.NilError(t, err)
		assert.DeepEqual(t, res, storedLocations)
	})
}

func TestServiceListAips(t *testing.T) {
	t.Parallel()

	t.Run("Returns defined AIPs", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		payload := &goastorage.ListAipsPayload{
			Limit: ref.New(20),
		}
		aips := &goastorage.AIPs{
			Items: goastorage.AIPCollection{
				{
					Name:      "Test AIP 1",
					UUID:      aipID,
					ObjectKey: objectKey,
					Status:    "stored",
					CreatedAt: time.Now().Format(time.RFC3339),
				},
			},
			Page: &goastorage.EnduroPage{
				Limit:  20,
				Offset: 0,
				Total:  1,
			},
		}

		attrs.persistenceMock.
			EXPECT().
			ListAIPs(ctx, payload).
			Return(aips, nil)

		res, err := svc.ListAips(ctx, payload)
		assert.NilError(t, err)
		assert.DeepEqual(t, res, aips)
	})

	t.Run("Returns an error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		mockErr := errors.New("test error")

		attrs.persistenceMock.
			EXPECT().
			ListAIPs(ctx, nil).
			Return(nil, mockErr)

		_, err := svc.ListAips(ctx, nil)
		assert.ErrorIs(t, err, mockErr)
	})
}

func TestReject(t *testing.T) {
	t.Parallel()

	t.Run("Returns not_valid if AIP ID is invalid", func(t *testing.T) {
		t.Parallel()

		aipID := "12345"
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		err := svc.RejectAip(context.Background(), &goastorage.RejectAipPayload{
			UUID: aipID,
		})
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Rejects the AIP", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			UpdateAIPStatus(
				ctx,
				aipID,
				enums.AIPStatusDeleted,
			).
			Return(nil)

		err := svc.RejectAip(ctx, &goastorage.RejectAipPayload{UUID: aipID.String()})
		assert.NilError(t, err)
	})
}

func TestServiceReadAip(t *testing.T) {
	t.Parallel()

	attrs := setUpAttrs{}
	svc := setUpService(t, &attrs)
	ctx := context.Background()

	attrs.persistenceMock.
		EXPECT().
		ReadAIP(
			ctx,
			aipID,
		).
		Return(
			&goastorage.AIP{},
			nil,
		)

	aip, err := svc.ReadAip(ctx, aipID)
	assert.NilError(t, err)
	assert.DeepEqual(t, aip, &goastorage.AIP{})
}

func TestServiceUpdateAipStatus(t *testing.T) {
	t.Parallel()

	t.Run("Returns if persistence failed", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			UpdateAIPStatus(
				ctx,
				aipID,
				enums.AIPStatusStored,
			).
			Return(errors.New("something is wrong"))

		err := svc.UpdateAipStatus(ctx, aipID, enums.AIPStatusStored)
		assert.Error(t, err, "something is wrong")
	})
}

func TestServiceUpdateAipLocationUUID(t *testing.T) {
	t.Parallel()

	t.Run("Returns if persistence failed", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			UpdateAIPLocationID(
				ctx,
				aipID,
				locationID,
			).
			Return(errors.New("something is wrong"))

		err := svc.UpdateAipLocationID(ctx, aipID, locationID)
		assert.Error(t, err, "something is wrong")
	})
}

func TestServiceDelete(t *testing.T) {
	t.Parallel()

	t.Run("From internal location", func(t *testing.T) {
		t.Parallel()

		var attrs setUpAttrs
		svc := setUpService(t, &attrs)
		ctx := context.Background()

		// Write a test blob to the internal bucket.
		writeTestBlob(ctx, t, "file://"+attrs.config.Internal.URL, aipID.String())

		attrs.persistenceMock.
			EXPECT().
			ReadAIP(
				ctx,
				aipID,
			).
			Return(
				&goastorage.AIP{
					UUID:         aipID,
					ObjectKey:    aipID,
					LocationUUID: &uuid.Nil,
				},
				nil,
			)

		err := svc.DeleteAip(ctx, aipID)
		assert.NilError(t, err)
	})

	t.Run("From perma location", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		td := tfs.NewDir(t, "enduro-service-test")

		var attrs setUpAttrs
		svc := setUpService(t, &attrs)

		// Write a test blob to the perma location.
		writeTestBlob(ctx, t, "file://"+td.Path(), aipID.String())

		attrs.persistenceMock.
			EXPECT().
			ReadAIP(
				ctx,
				aipID,
			).
			Return(
				&goastorage.AIP{
					UUID:         aipID,
					ObjectKey:    aipID,
					LocationUUID: &locationID,
				},
				nil,
			)

		attrs.persistenceMock.
			EXPECT().
			ReadLocation(
				ctx,
				locationID,
			).
			Return(
				&goastorage.Location{
					UUID: locationID,
					Config: &goastorage.URLConfig{
						URL: "file://" + td.Path(),
					},
				},
				nil,
			)

		err := svc.DeleteAip(ctx, aipID)
		assert.NilError(t, err)
	})

	t.Run("Fails if object does not exist", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		td := tfs.NewDir(t, "enduro-service-test")

		var attrs setUpAttrs
		svc := setUpService(t, &attrs)

		attrs.persistenceMock.
			EXPECT().
			ReadAIP(
				ctx,
				aipID,
			).
			Return(
				&goastorage.AIP{
					UUID:         aipID,
					ObjectKey:    objectKey,
					LocationUUID: &locationID,
				},
				nil,
			)

		attrs.persistenceMock.
			EXPECT().
			ReadLocation(
				ctx,
				locationID,
			).
			Return(
				&goastorage.Location{
					UUID: locationID,
					Config: &goastorage.URLConfig{
						URL: "file://" + td.Path(),
					},
				},
				nil,
			)

		err := svc.DeleteAip(ctx, aipID)
		assert.Error(t, err, fmt.Sprintf(
			"blob (key %q) (code=NotFound): remove %s/%s: no such file or directory",
			aipID.String(), td.Path(), aipID.String(),
		))
	})

	t.Run("Fails if location is not available", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			ReadAIP(
				ctx,
				aipID,
			).
			Return(
				&goastorage.AIP{
					UUID:         aipID,
					ObjectKey:    objectKey,
					LocationUUID: &locationID,
				},
				nil,
			)

		attrs.persistenceMock.
			EXPECT().
			ReadLocation(
				ctx,
				locationID,
			).
			Return(
				nil,
				&goastorage.LocationNotFound{UUID: locationID, Message: "location not found"},
			)

		err := svc.DeleteAip(ctx, aipID)
		assert.ErrorContains(t, err, "location not found")
	})

	t.Run("Fails if AIP does not exist", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			ReadAIP(
				ctx,
				aipID,
			).
			Return(
				nil,
				&goastorage.AIPNotFound{UUID: aipID, Message: "AIP not found"},
			)

		err := svc.DeleteAip(ctx, aipID)
		assert.ErrorContains(t, err, "AIP not found")
	})
}

func TestAipReader(t *testing.T) {
	t.Parallel()

	t.Run("Provides a valid reader", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		td := tfs.NewDir(t, "enduro-service-test")

		var attrs setUpAttrs
		svc := setUpService(t, &attrs)

		// Write a test blob to the bucket.
		writeTestBlob(ctx, t, "file://"+td.Path(), aipID.String())

		attrs.persistenceMock.
			EXPECT().
			ReadLocation(
				ctx,
				locationID,
			).
			Return(
				&goastorage.Location{
					UUID: locationID,
					Config: &goastorage.URLConfig{
						URL: "file://" + td.Path(),
					},
				},
				nil,
			)

		reader, err := svc.AipReader(ctx, &goastorage.AIP{
			UUID:         aipID,
			ObjectKey:    aipID,
			LocationUUID: &locationID,
		})
		assert.NilError(t, err)
		defer reader.Close()

		blob, err := io.ReadAll(reader)
		assert.NilError(t, err)
		assert.Equal(t, string(blob), "Testing 1-2-3!")
	})

	t.Run("Fails if the location is not available", func(t *testing.T) {
		t.Parallel()

		attrs := setUpAttrs{}
		svc := setUpService(t, &attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			ReadLocation(
				ctx,
				locationID,
			).
			Return(
				nil,
				&goastorage.LocationNotFound{UUID: locationID, Message: "location not found"},
			)

		_, err := svc.AipReader(ctx, &goastorage.AIP{
			UUID:         aipID,
			ObjectKey:    aipID,
			LocationUUID: &locationID,
		})
		assert.ErrorContains(t, err, "location not found")
	})

	t.Run("Fails if the reader cannot be created", func(t *testing.T) {
		t.Parallel()

		var attrs setUpAttrs
		svc := setUpService(t, &attrs)
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
					Config: &goastorage.URLConfig{
						URL: "mem://",
					},
				},
				nil,
			)

		_, err := svc.AipReader(ctx, &goastorage.AIP{
			UUID:         aipID,
			ObjectKey:    aipID,
			LocationUUID: &locationID,
		})
		assert.Error(t, err, fmt.Sprintf(
			"blob (key %q) (code=NotFound): blob not found",
			aipID.String(),
		))
	})
}

func TestListAipWorkflows(t *testing.T) {
	t.Parallel()

	workflows := goastorage.AIPWorkflowCollection{
		{
			UUID:       uuid.New(),
			TemporalID: "temporal-id-1",
			Type:       enums.WorkflowTypeMoveAip.String(),
			Status:     enums.WorkflowStatusDone.String(),
		},
		{
			UUID:       uuid.New(),
			TemporalID: "temporal-id-2",
			Type:       enums.WorkflowTypeDeleteAip.String(),
			Status:     enums.WorkflowStatusInProgress.String(),
		},
	}

	type test struct {
		name    string
		payload *goastorage.ListAipWorkflowsPayload
		mock    func(context.Context, *fake.MockStorage)
		want    *goastorage.AIPWorkflows
		wantErr string
	}

	for _, tt := range []test{
		{
			name: "Lists AIP workflows",
			payload: &goastorage.ListAipWorkflowsPayload{
				UUID: aipID.String(),
			},
			mock: func(ctx context.Context, s *fake.MockStorage) {
				s.EXPECT().
					ListWorkflows(ctx, &persistence.WorkflowFilter{AIPUUID: &aipID}).
					Return(workflows, nil)
			},
			want: &goastorage.AIPWorkflows{Workflows: workflows},
		},
		{
			name: "Filter AIP workflows by status",
			payload: &goastorage.ListAipWorkflowsPayload{
				UUID:   aipID.String(),
				Status: ref.New(enums.WorkflowStatusInProgress.String()),
			},
			mock: func(ctx context.Context, s *fake.MockStorage) {
				s.EXPECT().
					ListWorkflows(ctx, &persistence.WorkflowFilter{
						AIPUUID: &aipID,
						Status:  ref.New(enums.WorkflowStatusInProgress),
					}).
					Return(workflows[1:], nil)
			},
			want: &goastorage.AIPWorkflows{Workflows: workflows[1:]},
		},
		{
			name: "Filter AIP workflows by type",
			payload: &goastorage.ListAipWorkflowsPayload{
				UUID: aipID.String(),
				Type: ref.New(enums.WorkflowTypeMoveAip.String()),
			},
			mock: func(ctx context.Context, s *fake.MockStorage) {
				s.EXPECT().
					ListWorkflows(ctx, &persistence.WorkflowFilter{
						AIPUUID: &aipID,
						Type:    ref.New(enums.WorkflowTypeMoveAip),
					}).
					Return(workflows[:1], nil)
			},
			want: &goastorage.AIPWorkflows{Workflows: workflows[:1]},
		},
		{
			name: "Fails on invalid AIP UUID",
			payload: &goastorage.ListAipWorkflowsPayload{
				UUID: "invalid-uuid",
			},
			wantErr: "UUID: invalid value",
		},
		{
			name: "Fails on invalid workflow status",
			payload: &goastorage.ListAipWorkflowsPayload{
				UUID:   aipID.String(),
				Status: ref.New("bad status"),
			},
			wantErr: "status: invalid value",
		},
		{
			name: "Fails on invalid workflow type",
			payload: &goastorage.ListAipWorkflowsPayload{
				UUID: aipID.String(),
				Type: ref.New("bad type"),
			},
			wantErr: "type: invalid value",
		},
		{
			name: "Fails on persistence error",
			payload: &goastorage.ListAipWorkflowsPayload{
				UUID: aipID.String(),
			},
			mock: func(ctx context.Context, s *fake.MockStorage) {
				s.EXPECT().
					ListWorkflows(ctx, &persistence.WorkflowFilter{AIPUUID: &aipID}).
					Return(nil, errors.New("persistence error"))
			},
			wantErr: "cannot perform operation",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			attrs := &setUpAttrs{}
			svc := setUpService(t, attrs)

			if tt.mock != nil {
				tt.mock(ctx, attrs.persistenceMock)
			}

			re, err := svc.ListAipWorkflows(ctx, tt.payload)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, re, tt.want)
		})
	}
}

func TestServiceUpdate(t *testing.T) {
	t.Parallel()

	t.Run("Returns not_valid if AIP ID is invalid", func(t *testing.T) {
		t.Parallel()

		aipID := "12345"
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		err := svc.UpdateAip(context.Background(), &goastorage.UpdateAipPayload{
			UUID: aipID,
		})
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns not_available if workflow cannot be signaled", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"SignalWorkflow",
				ctx,
				"storage-upload-workflow-"+aipID.String(),
				"",
				"upload-done-signal",
				storage.UploadDoneSignal{},
			).
			Return(
				errors.New("something went wrong"),
			)

		err := svc.UpdateAip(ctx, &goastorage.UpdateAipPayload{
			UUID: aipID.String(),
		})
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_available")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns not_valid if AIP cannot be updated", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"SignalWorkflow",
				ctx,
				"storage-upload-workflow-"+aipID.String(),
				"",
				"upload-done-signal",
				storage.UploadDoneSignal{},
			).
			Return(
				nil,
			)

		attrs.persistenceMock.
			EXPECT().
			UpdateAIPStatus(
				gomock.AssignableToTypeOf(ctx),
				gomock.Any(),
				enums.AIPStatusPending,
			).
			Return(
				errors.New("unexpected error"),
			)

		err := svc.UpdateAip(ctx, &goastorage.UpdateAipPayload{
			UUID: aipID.String(),
		})
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot update AIP status")
	})

	t.Run("Returns no error if AIP is updated", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"SignalWorkflow",
				ctx,
				"storage-upload-workflow-"+aipID.String(),
				"",
				"upload-done-signal",
				storage.UploadDoneSignal{},
			).
			Return(
				nil,
			)

		attrs.persistenceMock.
			EXPECT().
			UpdateAIPStatus(
				gomock.AssignableToTypeOf(ctx),
				gomock.Any(),
				enums.AIPStatusPending,
			).
			Return(
				nil,
			)

		err := svc.UpdateAip(ctx, &goastorage.UpdateAipPayload{
			UUID: aipID.String(),
		})
		assert.NilError(t, err)
	})
}

func TestServiceMove(t *testing.T) {
	t.Parallel()

	t.Run("Returns not_valid if AIP ID is invalid", func(t *testing.T) {
		t.Parallel()

		aipID := "12345"
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		err := svc.MoveAip(context.Background(), &goastorage.MoveAipPayload{
			UUID: aipID,
		})
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns not found error if AIP does not exist", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			ReadAIP(
				gomock.AssignableToTypeOf(ctx),
				aipID,
			).
			Return(
				nil,
				&goastorage.AIPNotFound{UUID: aipID, Message: "AIP not found"},
			)

		err := svc.MoveAip(ctx, &goastorage.MoveAipPayload{
			UUID:         aipID.String(),
			LocationUUID: locationID,
		})
		assert.ErrorContains(t, err, "AIP not found")
	})

	t.Run("Returns not_available if workflow cannot be executed", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"ExecuteWorkflow",
				mock.AnythingOfType("*context.timerCtx"),
				temporalsdk_client.StartWorkflowOptions{
					ID:                    "storage-move-workflow-" + aipID.String(),
					TaskQueue:             "global",
					WorkflowIDReusePolicy: temporalapi_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
				},
				"storage-move-workflow",
				&storage.StorageMoveWorkflowRequest{AIPID: aipID, LocationID: locationID, TaskQueue: "global"},
			).
			Return(
				nil,
				errors.New("something went wrong"),
			)

		attrs.persistenceMock.
			EXPECT().
			ReadAIP(
				gomock.AssignableToTypeOf(ctx),
				aipID,
			).
			Return(
				&goastorage.AIP{UUID: aipID},
				nil,
			)

		err := svc.MoveAip(ctx, &goastorage.MoveAipPayload{
			UUID:         aipID.String(),
			LocationUUID: locationID,
		})
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_available")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns no error if AIP is moved", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"ExecuteWorkflow",
				mock.AnythingOfType("*context.timerCtx"),
				temporalsdk_client.StartWorkflowOptions{
					ID:                    "storage-move-workflow-" + aipID.String(),
					TaskQueue:             "global",
					WorkflowIDReusePolicy: temporalapi_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
				},
				"storage-move-workflow",
				&storage.StorageMoveWorkflowRequest{AIPID: aipID, LocationID: locationID, TaskQueue: "global"},
			).
			Return(
				&temporalsdk_mocks.WorkflowRun{},
				nil,
			)

		attrs.persistenceMock.
			EXPECT().
			ReadAIP(
				gomock.AssignableToTypeOf(ctx),
				aipID,
			).
			Return(
				&goastorage.AIP{UUID: aipID},
				nil,
			)

		err := svc.MoveAip(ctx, &goastorage.MoveAipPayload{
			UUID:         aipID.String(),
			LocationUUID: locationID,
		})
		assert.NilError(t, err)
	})
}

func TestServiceMoveStatus(t *testing.T) {
	t.Parallel()

	t.Run("Returns not_valid if AIP ID is invalid", func(t *testing.T) {
		t.Parallel()

		aipID := "12345"
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		res, err := svc.MoveAipStatus(context.Background(), &goastorage.MoveAipStatusPayload{
			UUID: aipID,
		})
		assert.Assert(t, res == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns not found error if AIP does not exist", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			ReadAIP(
				gomock.AssignableToTypeOf(ctx),
				aipID,
			).
			Return(
				nil,
				&goastorage.AIPNotFound{UUID: aipID, Message: "AIP not found"},
			)

		res, err := svc.MoveAipStatus(ctx, &goastorage.MoveAipStatusPayload{
			UUID: aipID.String(),
		})
		assert.Assert(t, res == nil)
		assert.ErrorContains(t, err, "AIP not found")
	})

	t.Run("Returns failed_dependency error if workflow execution cannot be found", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()
		attrs.config.TaskQueue = "global"

		attrs.temporalClientMock.
			On(
				"DescribeWorkflowExecution",
				ctx,
				"storage-move-workflow-"+aipID.String(),
				"",
			).
			Return(
				nil,
				errors.New("something went wrong"),
			)

		attrs.persistenceMock.
			EXPECT().
			ReadAIP(
				gomock.AssignableToTypeOf(ctx),
				aipID,
			).
			Return(
				&goastorage.AIP{UUID: aipID},
				nil,
			)

		res, err := svc.MoveAipStatus(ctx, &goastorage.MoveAipStatusPayload{
			UUID: aipID.String(),
		})
		assert.Assert(t, res == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "failed_dependency")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns failed_dependency error if workflow execution failed", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()
		attrs.config.TaskQueue = "global"

		attrs.temporalClientMock.
			On(
				"DescribeWorkflowExecution",
				ctx,
				"storage-move-workflow-"+aipID.String(),
				"",
			).
			Return(
				&temporalapi_workflowservice.DescribeWorkflowExecutionResponse{
					WorkflowExecutionInfo: &temporalapi_workflow.WorkflowExecutionInfo{
						Status: temporalapi_enums.WORKFLOW_EXECUTION_STATUS_FAILED,
					},
				},
				nil,
			)

		attrs.persistenceMock.
			EXPECT().
			ReadAIP(
				gomock.AssignableToTypeOf(ctx),
				aipID,
			).
			Return(
				&goastorage.AIP{UUID: aipID},
				nil,
			)

		res, err := svc.MoveAipStatus(ctx, &goastorage.MoveAipStatusPayload{
			UUID: aipID.String(),
		})
		assert.Assert(t, res == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "failed_dependency")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns move not done if workflow is running", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"DescribeWorkflowExecution",
				ctx,
				"storage-move-workflow-"+aipID.String(),
				"",
			).
			Return(
				&temporalapi_workflowservice.DescribeWorkflowExecutionResponse{
					WorkflowExecutionInfo: &temporalapi_workflow.WorkflowExecutionInfo{
						Status: temporalapi_enums.WORKFLOW_EXECUTION_STATUS_RUNNING,
					},
				},
				nil,
			)

		attrs.persistenceMock.
			EXPECT().
			ReadAIP(
				gomock.AssignableToTypeOf(ctx),
				aipID,
			).
			Return(
				&goastorage.AIP{UUID: aipID},
				nil,
			)

		res, err := svc.MoveAipStatus(ctx, &goastorage.MoveAipStatusPayload{
			UUID: aipID.String(),
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, res, &goastorage.MoveStatusResult{Done: false})
	})

	t.Run("Returns move done if workflow completed", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.temporalClientMock.
			On(
				"DescribeWorkflowExecution",
				ctx,
				"storage-move-workflow-"+aipID.String(),
				"",
			).
			Return(
				&temporalapi_workflowservice.DescribeWorkflowExecutionResponse{
					WorkflowExecutionInfo: &temporalapi_workflow.WorkflowExecutionInfo{
						Status: temporalapi_enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
					},
				},
				nil,
			)

		attrs.persistenceMock.
			EXPECT().
			ReadAIP(
				gomock.AssignableToTypeOf(ctx),
				aipID,
			).
			Return(
				&goastorage.AIP{UUID: aipID},
				nil,
			)

		res, err := svc.MoveAipStatus(ctx, &goastorage.MoveAipStatusPayload{
			UUID: aipID.String(),
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

		res, err := svc.CreateLocation(ctx, &goastorage.CreateLocationPayload{
			Name:    "perma-aips-1",
			Source:  enums.LocationSourceMinio.String(),
			Purpose: enums.LocationPurposeAipStore.String(),
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

		res, err := svc.CreateLocation(ctx, &goastorage.CreateLocationPayload{
			Name:    "perma-aips-1",
			Source:  enums.LocationSourceMinio.String(),
			Purpose: enums.LocationPurposeAipStore.String(),
			Config:  &goastorage.S3Config{},
		})
		assert.Assert(t, res == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "invalid configuration")
	})

	t.Run("Returns not_valid if cannot persist location", func(t *testing.T) {
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			CreateLocation(
				gomock.AssignableToTypeOf(ctx),
				&goastorage.Location{
					Name:    "perma-aips-1",
					Source:  enums.LocationSourceMinio.String(),
					Purpose: enums.LocationPurposeAipStore.String(),
					UUID:    uuid0,
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
			)

		res, err := svc.CreateLocation(ctx, &goastorage.CreateLocationPayload{
			Name:    "perma-aips-1",
			Source:  enums.LocationSourceMinio.String(),
			Purpose: enums.LocationPurposeAipStore.String(),
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
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			CreateLocation(
				gomock.AssignableToTypeOf(ctx),
				&goastorage.Location{
					Name:    "perma-aips-1",
					Source:  enums.LocationSourceMinio.String(),
					Purpose: enums.LocationPurposeAipStore.String(),
					UUID:    uuid0,
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
			)

		res, err := svc.CreateLocation(ctx, &goastorage.CreateLocationPayload{
			Name:    "perma-aips-1",
			Source:  enums.LocationSourceMinio.String(),
			Purpose: enums.LocationPurposeAipStore.String(),
			Config: &goastorage.S3Config{
				Bucket: "perma-aips-1",
				Region: "planet-earth",
			},
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, res, &goastorage.CreateLocationResult{UUID: uuid0.String()})
	})

	t.Run("Returns location with URL config", func(t *testing.T) {
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			CreateLocation(
				gomock.AssignableToTypeOf(ctx),
				&goastorage.Location{
					Name:    "perma-aips-1",
					Source:  enums.LocationSourceMinio.String(),
					Purpose: enums.LocationPurposeAipStore.String(),
					UUID:    uuid0,
				},
				&types.LocationConfig{
					Value: &types.URLConfig{
						URL: "mem://",
					},
				},
			).
			Return(
				&goastorage.Location{},
				nil,
			)

		res, err := svc.CreateLocation(ctx, &goastorage.CreateLocationPayload{
			Name:    "perma-aips-1",
			Source:  enums.LocationSourceMinio.String(),
			Purpose: enums.LocationPurposeAipStore.String(),
			Config: &goastorage.URLConfig{
				URL: "mem://",
			},
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, res, &goastorage.CreateLocationResult{UUID: uuid0.String()})
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
			)

		res, err := svc.ShowLocation(ctx, &goastorage.ShowLocationPayload{
			UUID: locationID.String(),
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, res, &goastorage.Location{UUID: locationID})
	})
}

func TestServiceListLocationAips(t *testing.T) {
	t.Parallel()

	t.Run("Returns not_valid if cannot parse location UUID", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		res, err := svc.ListLocationAips(ctx, &goastorage.ListLocationAipsPayload{
			UUID: "hello world",
		})
		assert.Assert(t, res == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns not_available if AIPs cannot be read", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			LocationAIPs(
				ctx,
				locationID,
			).
			Return(
				nil,
				errors.New("unexpected error"),
			)

		res, err := svc.ListLocationAips(ctx, &goastorage.ListLocationAipsPayload{
			UUID: locationID.String(),
		})
		assert.Assert(t, res == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_available")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns stored AIPs", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			LocationAIPs(
				ctx,
				locationID,
			).
			Return(
				goastorage.AIPCollection{
					{
						Name:         "AIP",
						UUID:         aipID,
						Status:       "stored",
						ObjectKey:    objectKey,
						LocationUUID: &locationID,
						CreatedAt:    "2013-02-03T19:54:00Z",
					},
				},
				nil,
			)

		res, err := svc.ListLocationAips(ctx, &goastorage.ListLocationAipsPayload{
			UUID: locationID.String(),
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, res, goastorage.AIPCollection{
			{
				Name:         "AIP",
				UUID:         aipID,
				Status:       "stored",
				ObjectKey:    objectKey,
				LocationUUID: &locationID,
				CreatedAt:    "2013-02-03T19:54:00Z",
			},
		})
	})
}

func TestServiceShow(t *testing.T) {
	t.Parallel()

	t.Run("Returns not_valid if AIP ID is invalid", func(t *testing.T) {
		t.Parallel()

		aipID := "12345"
		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)

		res, err := svc.ShowAip(context.Background(), &goastorage.ShowAipPayload{
			UUID: aipID,
		})
		assert.Assert(t, res == nil)
		assert.Equal(t, err.(*goa.ServiceError).Name, "not_valid")
		assert.ErrorContains(t, err, "cannot perform operation")
	})

	t.Run("Returns stored AIP", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		attrs.persistenceMock.
			EXPECT().
			ReadAIP(
				ctx,
				aipID,
			).
			Return(
				&goastorage.AIP{
					UUID:         aipID,
					ObjectKey:    objectKey,
					LocationUUID: &uuid.Nil,
				},
				nil,
			)

		res, err := svc.ShowAip(ctx, &goastorage.ShowAipPayload{
			UUID: aipID.String(),
		})
		assert.NilError(t, err)
		assert.DeepEqual(t, res, &goastorage.AIP{
			UUID:         aipID,
			ObjectKey:    objectKey,
			LocationUUID: &uuid.Nil,
		})
	})
}

func TestCreateWorkflow(t *testing.T) {
	t.Parallel()

	t.Run("Creates a Workflow", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		dbID := 1
		workflow := &types.Workflow{
			UUID:       uuid.New(),
			TemporalID: "temporal-id",
			Type:       enums.WorkflowTypeMoveAip,
			Status:     enums.WorkflowStatusInProgress,
		}

		attrs.persistenceMock.
			EXPECT().
			CreateWorkflow(ctx, workflow).
			DoAndReturn(func(ctx context.Context, w *types.Workflow) error {
				w.DBID = dbID
				return nil
			})

		err := svc.CreateWorkflow(ctx, workflow)
		assert.NilError(t, err)
		assert.Equal(t, workflow.DBID, dbID)
	})

	t.Run("Returns a persistence error", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		workflow := &types.Workflow{
			UUID:       uuid.New(),
			TemporalID: "temporal-id",
			Type:       enums.WorkflowTypeMoveAip,
			Status:     enums.WorkflowStatusInProgress,
		}
		perErr := errors.New("persistence error")

		attrs.persistenceMock.
			EXPECT().
			CreateWorkflow(ctx, workflow).
			Return(perErr)

		err := svc.CreateWorkflow(ctx, workflow)
		assert.ErrorIs(t, err, perErr)
	})
}

func TestUpdateWorkflow(t *testing.T) {
	t.Parallel()

	t.Run("Updates a Workflow", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		workflowID := 1
		workflow := &types.Workflow{
			DBID:       workflowID,
			UUID:       uuid.New(),
			TemporalID: "temporal-id",
			Type:       enums.WorkflowTypeMoveAip,
			Status:     enums.WorkflowStatusInProgress,
		}
		updater := func(w *types.Workflow) (*types.Workflow, error) { return w, nil }

		attrs.persistenceMock.
			EXPECT().
			UpdateWorkflow(
				ctx,
				workflowID,
				mockutil.Func(
					"should update workflow",
					func(updater persistence.WorkflowUpdater) error {
						_, err := updater(workflow)
						return err
					},
				),
			).
			Return(workflow, nil)

		re, err := svc.UpdateWorkflow(ctx, workflowID, updater)
		assert.NilError(t, err)
		assert.DeepEqual(t, re, workflow)
	})
}

func TestCreateTask(t *testing.T) {
	t.Parallel()

	t.Run("Creates a Task", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		dbID := 1
		task := &types.Task{
			UUID:   uuid.New(),
			Name:   "task",
			Status: enums.TaskStatusInProgress,
		}

		attrs.persistenceMock.
			EXPECT().
			CreateTask(ctx, task).
			DoAndReturn(func(ctx context.Context, t *types.Task) error {
				t.DBID = dbID
				return nil
			})

		err := svc.CreateTask(ctx, task)
		assert.NilError(t, err)
		assert.Equal(t, task.DBID, dbID)
	})

	t.Run("Returns a persistence error", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		task := &types.Task{
			UUID:   uuid.New(),
			Name:   "task",
			Status: enums.TaskStatusInProgress,
		}
		perErr := errors.New("persistence error")

		attrs.persistenceMock.
			EXPECT().
			CreateTask(ctx, task).
			Return(perErr)

		err := svc.CreateTask(ctx, task)
		assert.ErrorIs(t, err, perErr)
	})
}

func TestUpdateTask(t *testing.T) {
	t.Parallel()

	t.Run("Updates a Task", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		taskID := 1
		task := &types.Task{
			DBID:   taskID,
			UUID:   uuid.New(),
			Name:   "task",
			Status: enums.TaskStatusInProgress,
		}
		updater := func(t *types.Task) (*types.Task, error) { return t, nil }

		attrs.persistenceMock.
			EXPECT().
			UpdateTask(
				ctx,
				taskID,
				mockutil.Func(
					"should update task",
					func(updater persistence.TaskUpdater) error {
						_, err := updater(task)
						return err
					},
				),
			).
			Return(task, nil)

		re, err := svc.UpdateTask(ctx, taskID, updater)
		assert.NilError(t, err)
		assert.DeepEqual(t, re, task)
	})
}

func TestCreateDeletionRequest(t *testing.T) {
	t.Parallel()

	t.Run("Creates a DeletionRequest", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		dbID := 1
		dr := &types.DeletionRequest{
			UUID:        uuid.New(),
			AIPUUID:     uuid.New(),
			Reason:      "Reason",
			Status:      enums.DeletionRequestStatusPending,
			RequestedAt: time.Now(),
		}

		attrs.persistenceMock.
			EXPECT().
			CreateDeletionRequest(ctx, dr).
			DoAndReturn(func(ctx context.Context, dr *types.DeletionRequest) error {
				dr.DBID = dbID
				return nil
			})

		err := svc.CreateDeletionRequest(ctx, dr)
		assert.NilError(t, err)
		assert.Equal(t, dr.DBID, dbID)
	})

	t.Run("Returns a persistence error", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		dr := &types.DeletionRequest{
			UUID:        uuid.New(),
			AIPUUID:     uuid.New(),
			Reason:      "Reason",
			Status:      enums.DeletionRequestStatusPending,
			RequestedAt: time.Now(),
		}
		perErr := errors.New("persistence error")

		attrs.persistenceMock.
			EXPECT().
			CreateDeletionRequest(ctx, dr).
			Return(perErr)

		err := svc.CreateDeletionRequest(ctx, dr)
		assert.ErrorIs(t, err, perErr)
	})
}

func TestUpdateDeletionRequest(t *testing.T) {
	t.Parallel()

	t.Run("Updates a DeletionRequest", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		drID := 1
		dr := &types.DeletionRequest{
			DBID:        drID,
			UUID:        uuid.New(),
			AIPUUID:     uuid.New(),
			Reason:      "Updated reason",
			Status:      enums.DeletionRequestStatusApproved,
			RequestedAt: time.Now(),
			ReviewedAt:  time.Now(),
			Reviewer:    "reviewer@example.com",
			ReviewerIss: "issuer",
			ReviewerSub: "subject",
		}
		updater := func(dr *types.DeletionRequest) (*types.DeletionRequest, error) { return dr, nil }

		attrs.persistenceMock.
			EXPECT().
			UpdateDeletionRequest(
				ctx,
				drID,
				mockutil.Func(
					"should update deletion request",
					func(updater persistence.DeletionRequestUpdater) error {
						_, err := updater(dr)
						return err
					},
				),
			).
			Return(dr, nil)

		re, err := svc.UpdateDeletionRequest(ctx, drID, updater)
		assert.NilError(t, err)
		assert.DeepEqual(t, re, dr)
	})
}

func TestReadAipPendingDeletionRequest(t *testing.T) {
	t.Parallel()

	t.Run("Returns pending DeletionRequest", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		aipUUID := uuid.New()
		dr := &types.DeletionRequest{
			UUID:        uuid.New(),
			AIPUUID:     aipUUID,
			Reason:      "Pending deletion request",
			Status:      enums.DeletionRequestStatusPending,
			RequestedAt: time.Now(),
		}

		attrs.persistenceMock.
			EXPECT().
			ReadAipPendingDeletionRequest(ctx, aipUUID).
			Return(dr, nil)

		re, err := svc.ReadAipPendingDeletionRequest(ctx, aipUUID)
		assert.NilError(t, err)
		assert.DeepEqual(t, re, dr)
	})

	t.Run("Returns a persistence error", func(t *testing.T) {
		t.Parallel()

		attrs := &setUpAttrs{}
		svc := setUpService(t, attrs)
		ctx := context.Background()

		aipUUID := uuid.New()
		perErr := errors.New("persistence error")

		attrs.persistenceMock.
			EXPECT().
			ReadAipPendingDeletionRequest(ctx, aipUUID).
			Return(nil, perErr)

		_, err := svc.ReadAipPendingDeletionRequest(ctx, aipUUID)
		assert.ErrorIs(t, err, perErr)
	})
}
