package storage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"entgo.io/ent/examples/o2o2types/ent"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	temporalapi_enums "go.temporal.io/api/enums/v1"
	temporalsdk_client "go.temporal.io/sdk/client"
	"gocloud.dev/blob"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

var SubmitURLExpirationTime = 15 * time.Minute

type Service interface {
	// Used in the Goa API.
	Submit(context.Context, *goastorage.SubmitPayload) (res *goastorage.SubmitResult, err error)
	Update(context.Context, *goastorage.UpdatePayload) (err error)
	Download(context.Context, *goastorage.DownloadPayload) ([]byte, error)
	Locations(context.Context) (res goastorage.StoredLocationCollection, err error)
	Move(context.Context, *goastorage.MovePayload) (err error)
	MoveStatus(context.Context, *goastorage.MoveStatusPayload) (res *goastorage.MoveStatusResult, err error)
	Reject(context.Context, *goastorage.RejectPayload) (err error)
	Show(context.Context, *goastorage.ShowPayload) (res *goastorage.StoredStoragePackage, err error)
	AddLocation(context.Context, *goastorage.AddLocationPayload) (res *goastorage.AddLocationResult, err error)
	ShowLocation(context.Context, *goastorage.ShowLocationPayload) (res *goastorage.StoredLocation, err error)

	// Used from workflow activities.
	Location(locationID uuid.UUID) (Location, error)
	ReadPackage(ctx context.Context, AIPID string) (*goastorage.StoredStoragePackage, error)
	UpdatePackageStatus(ctx context.Context, status types.PackageStatus, aipID string) error
	UpdatePackageLocationID(ctx context.Context, locationID uuid.UUID, aipID string) error
	Delete(ctx context.Context, AIPID string) (err error)

	// Both.
	PackageReader(ctx context.Context, pkg *goastorage.StoredStoragePackage) (*blob.Reader, error)
}

type serviceImpl struct {
	logger logr.Logger
	config Config

	// Internal processing location.
	internal Location

	// Temporal client.
	tc temporalsdk_client.Client

	// Persistence client.
	storagePersistence persistence.Storage
}

var _ Service = (*serviceImpl)(nil)

func NewService(logger logr.Logger, config Config, storagePersistence persistence.Storage, tc temporalsdk_client.Client) (s *serviceImpl, err error) {
	s = &serviceImpl{
		logger:             logger,
		tc:                 tc,
		config:             config,
		storagePersistence: storagePersistence,
	}

	s.internal, err = NewInternalLocation(config.Internal)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *serviceImpl) Location(locationID uuid.UUID) (Location, error) {
	if locationID == uuid.Nil {
		return s.internal, nil
	}

	// TODO: get location from the database based on name

	return nil, nil
}

func (s *serviceImpl) Submit(ctx context.Context, payload *goastorage.SubmitPayload) (*goastorage.SubmitResult, error) {
	AIPUUID, err := uuid.Parse(payload.AipID)
	if err != nil {
		return nil, goastorage.MakeNotValid(err)
	}

	_, err = InitStorageUploadWorkflow(ctx, s.tc, &StorageUploadWorkflowRequest{AIPID: payload.AipID})
	if err != nil {
		return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	objectKey := uuid.New()
	_, err = s.storagePersistence.CreatePackage(ctx, &goastorage.StoragePackage{
		Name:      payload.Name,
		AipID:     AIPUUID.String(),
		ObjectKey: &objectKey,
	})
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot persist package"))
	}

	bucket := s.internal.Bucket()
	opts := &blob.SignedURLOptions{
		Expiry: SubmitURLExpirationTime,
		Method: http.MethodPut,
	}
	url, err := bucket.SignedURL(ctx, objectKey.String(), opts)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot persist package"))
	}

	result := &goastorage.SubmitResult{
		URL: url,
	}
	return result, nil
}

func (s *serviceImpl) Update(ctx context.Context, payload *goastorage.UpdatePayload) error {
	signal := UploadDoneSignal{}
	workflowID := fmt.Sprintf("%s-%s", StorageUploadWorkflowName, payload.AipID)
	err := s.tc.SignalWorkflow(ctx, workflowID, "", UploadDoneSignalName, signal)
	if err != nil {
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}
	// Uptade the package status to in_review
	err = s.UpdatePackageStatus(ctx, types.StatusInReview, payload.AipID)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("cannot persist package"))
	}

	return nil
}

func (s *serviceImpl) Download(ctx context.Context, payload *goastorage.DownloadPayload) ([]byte, error) {
	// This service method is unused, see the Download function instead which
	// makes use of http.ResponseWriter.
	return []byte{}, nil
}

func (s *serviceImpl) Locations(ctx context.Context) (goastorage.StoredLocationCollection, error) {
	return s.storagePersistence.ListLocations(ctx)
}

func (s *serviceImpl) Move(ctx context.Context, payload *goastorage.MovePayload) error {
	pkg, err := s.ReadPackage(ctx, payload.AipID)
	if errors.Is(err, &ent.NotFoundError{}) || errors.Is(err, &ent.NotSingularError{}) {
		return &goastorage.StoragePackageNotfound{AipID: payload.AipID, Message: "not_found"}
	} else if err != nil {
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	_, err = InitStorageMoveWorkflow(ctx, s.tc, &StorageMoveWorkflowRequest{
		AIPID:      pkg.AipID,
		LocationID: payload.LocationID,
	})
	if err != nil {
		s.logger.Error(err, "error initializing move workflow")
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

func (s *serviceImpl) MoveStatus(ctx context.Context, payload *goastorage.MoveStatusPayload) (*goastorage.MoveStatusResult, error) {
	p, err := s.ReadPackage(ctx, payload.AipID)
	if errors.Is(err, &ent.NotFoundError{}) || errors.Is(err, &ent.NotSingularError{}) {
		return nil, &goastorage.StoragePackageNotfound{AipID: payload.AipID, Message: "not_found"}
	} else if err != nil {
		return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	resp, err := s.tc.DescribeWorkflowExecution(ctx, fmt.Sprintf("%s-%s", StorageMoveWorkflowName, p.AipID), "")
	if err != nil {
		return nil, goastorage.MakeFailedDependency(errors.New("cannot perform operation"))
	}

	var done bool
	switch resp.WorkflowExecutionInfo.Status {
	case
		temporalapi_enums.WORKFLOW_EXECUTION_STATUS_FAILED,
		temporalapi_enums.WORKFLOW_EXECUTION_STATUS_CANCELED,
		temporalapi_enums.WORKFLOW_EXECUTION_STATUS_TERMINATED,
		temporalapi_enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
		return nil, goastorage.MakeFailedDependency(errors.New("cannot perform operation"))
	case temporalapi_enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		done = true
	case temporalapi_enums.WORKFLOW_EXECUTION_STATUS_RUNNING:
		done = false
	}

	return &goastorage.MoveStatusResult{Done: done}, nil
}

func (s *serviceImpl) Reject(ctx context.Context, payload *goastorage.RejectPayload) error {
	return s.UpdatePackageStatus(ctx, types.StatusRejected, payload.AipID)
}

func (s *serviceImpl) Show(ctx context.Context, payload *goastorage.ShowPayload) (*goastorage.StoredStoragePackage, error) {
	pkg, err := s.ReadPackage(ctx, payload.AipID)
	if errors.Is(err, &ent.NotFoundError{}) || errors.Is(err, &ent.NotSingularError{}) {
		return nil, &goastorage.StoragePackageNotfound{AipID: payload.AipID, Message: "not_found"}
	} else if err != nil {
		return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return pkg, nil
}

func (s *serviceImpl) ReadPackage(ctx context.Context, AIPID string) (*goastorage.StoredStoragePackage, error) {
	AIPUUID, err := uuid.Parse(AIPID)
	if err != nil {
		return nil, err
	}

	return s.storagePersistence.ReadPackage(ctx, AIPUUID)
}

func (s *serviceImpl) UpdatePackageStatus(ctx context.Context, status types.PackageStatus, AIPID string) error {
	AIPUUID, err := uuid.Parse(AIPID)
	if err != nil {
		return err
	}

	return s.storagePersistence.UpdatePackageStatus(ctx, status, AIPUUID)
}

func (s *serviceImpl) UpdatePackageLocationID(ctx context.Context, locationID uuid.UUID, AIPID string) error {
	AIPUUID, err := uuid.Parse(AIPID)
	if err != nil {
		return err
	}

	return s.storagePersistence.UpdatePackageLocationID(ctx, locationID, AIPUUID)
}

// packageBucket returns the bucket and the key of the given package.
func (s *serviceImpl) packageBucket(p *goastorage.StoredStoragePackage) (*blob.Bucket, string, error) {
	// Package is still in the internal processing bucket.
	if p.LocationID == nil || *p.LocationID == uuid.Nil {
		return s.internal.Bucket(), p.ObjectKey.String(), nil
	}

	location, err := s.Location(*p.LocationID)
	if err != nil {
		return nil, "", err
	}
	return location.Bucket(), p.AipID, nil
}

func (s *serviceImpl) Delete(ctx context.Context, AIPID string) error {
	pkg, err := s.ReadPackage(ctx, AIPID)
	if err != nil {
		return err
	}

	bucket, key, err := s.packageBucket(pkg)
	if err != nil {
		return err
	}

	return bucket.Delete(ctx, key)
}

func (s *serviceImpl) PackageReader(ctx context.Context, pkg *goastorage.StoredStoragePackage) (*blob.Reader, error) {
	bucket, key, err := s.packageBucket(pkg)
	if err != nil {
		return nil, err
	}

	reader, err := bucket.NewReader(ctx, key, nil)
	if err != nil {
		return nil, err
	}

	return reader, nil
}

func (s *serviceImpl) AddLocation(ctx context.Context, payload *goastorage.AddLocationPayload) (res *goastorage.AddLocationResult, err error) {
	source := types.NewLocationSource(payload.Source)
	purpose := types.NewLocationPurpose(payload.Purpose)
	UUID := uuid.New()

	var config types.LocationConfig
	switch c := payload.Config.(type) {
	case *goastorage.S3Config:
		config.Value = c.ConvertToS3Config()
	default:
		return nil, fmt.Errorf("unsupported config type: %T", c)
	}

	if !config.Value.Valid() {
		return nil, goastorage.MakeNotValid(errors.New("invalid configuration"))
	}

	_, err = s.storagePersistence.CreateLocation(ctx, &goastorage.Location{
		Name:        payload.Name,
		Description: payload.Description,
		Source:      source.String(),
		Purpose:     purpose.String(),
		UUID:        &UUID,
	}, &config)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot persist location"))
	}

	return &goastorage.AddLocationResult{UUID: UUID.String()}, nil
}

func (s *serviceImpl) ReadLocation(ctx context.Context, UUID uuid.UUID) (*goastorage.StoredLocation, error) {
	return s.storagePersistence.ReadLocation(ctx, UUID)
}

func (s *serviceImpl) ShowLocation(ctx context.Context, payload *goastorage.ShowLocationPayload) (*goastorage.StoredLocation, error) {
	locationID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return nil, &goastorage.StorageLocationNotfound{UUID: locationID, Message: "not_found"}
	}

	l, err := s.ReadLocation(ctx, locationID)
	if errors.Is(err, &ent.NotFoundError{}) || errors.Is(err, &ent.NotSingularError{}) {
		return nil, &goastorage.StorageLocationNotfound{UUID: locationID, Message: "not_found"}
	} else if err != nil {
		return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return l, nil
}
