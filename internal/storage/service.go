package storage

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	temporalapi_enums "go.temporal.io/api/enums/v1"
	temporalsdk_client "go.temporal.io/sdk/client"
	"goa.design/goa/v3/security"
	"gocloud.dev/blob"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

var SubmitURLExpirationTime = 15 * time.Minute

// Service provides an interface for persisting storage data.
type Service interface {
	goastorage.Service

	// Used from workflow activities.
	Location(ctx context.Context, locationID uuid.UUID) (Location, error)
	ReadPackage(ctx context.Context, aipID uuid.UUID) (*goastorage.Package, error)
	UpdatePackageStatus(ctx context.Context, aipID uuid.UUID, status types.AIPStatus) error
	UpdatePackageLocationID(ctx context.Context, aipID, locationID uuid.UUID) error
	Delete(ctx context.Context, aipID uuid.UUID) (err error)

	// Both.
	PackageReader(ctx context.Context, pkg *goastorage.Package) (*blob.Reader, error)
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

	// Token verifier.
	tokenVerifier auth.TokenVerifier

	// Random number generator
	rander io.Reader
}

var _ Service = (*serviceImpl)(nil)

var (
	ErrUnauthorized error = goastorage.Unauthorized("Unauthorized")
	ErrForbidden    error = goastorage.Forbidden("Forbidden")
)

func NewService(
	logger logr.Logger,
	config Config,
	storagePersistence persistence.Storage,
	tc temporalsdk_client.Client,
	tokenVerifier auth.TokenVerifier,
	rander io.Reader,
) (s *serviceImpl, err error) {
	s = &serviceImpl{
		logger:             logger,
		tc:                 tc,
		config:             config,
		storagePersistence: storagePersistence,
		tokenVerifier:      tokenVerifier,
		rander:             rander,
	}

	if s.rander == nil {
		s.rander = rand.Reader
	}

	l, err := NewInternalLocation(&config.Internal)
	if err != nil {
		return nil, err
	}
	s.internal = l

	return s, nil
}

func (s *serviceImpl) JWTAuth(
	ctx context.Context,
	token string,
	scheme *security.JWTScheme,
) (context.Context, error) {
	claims, err := s.tokenVerifier.Verify(ctx, token)
	if err != nil {
		if !errors.Is(err, auth.ErrUnauthorized) {
			s.logger.V(1).Info("failed to verify token", "err", err)
		}
		return ctx, ErrUnauthorized
	}

	if !claims.CheckAttributes(scheme.RequiredScopes) {
		return ctx, ErrForbidden
	}

	ctx = auth.WithUserClaims(ctx, claims)

	return ctx, nil
}

func (s *serviceImpl) Location(ctx context.Context, locationID uuid.UUID) (Location, error) {
	if locationID == uuid.Nil {
		return s.internal, nil
	}

	goaLoc, err := s.ReadLocation(ctx, locationID)
	if err != nil {
		return nil, err
	}

	return NewLocation(goaLoc)
}

func (s *serviceImpl) Submit(ctx context.Context, payload *goastorage.SubmitPayload) (*goastorage.SubmitResult, error) {
	aipID, err := uuid.Parse(payload.AipID)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot perform operation"))
	}

	_, err = InitStorageUploadWorkflow(ctx, s.tc, &StorageUploadWorkflowRequest{
		AIPID:     aipID,
		TaskQueue: s.config.TaskQueue,
	})
	if err != nil {
		s.logger.Error(err, "storage service: InitStorageUploadWorkflow")
		return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	objectKey := uuid.Must(uuid.NewRandomFromReader(s.rander))

	_, err = s.storagePersistence.CreateAIP(ctx, &goastorage.Package{
		Name:      payload.Name,
		AipID:     aipID,
		ObjectKey: objectKey,
	})
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot create package"))
	}

	bucket, err := s.internal.OpenBucket(ctx)
	if err != nil {
		return nil, err
	}
	defer bucket.Close()

	opts := &blob.SignedURLOptions{
		Expiry: SubmitURLExpirationTime,
		Method: http.MethodPut,
	}
	url, err := bucket.SignedURL(ctx, objectKey.String(), opts)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot sign URL"))
	}

	result := &goastorage.SubmitResult{
		URL: url,
	}
	return result, nil
}

func (s *serviceImpl) Create(ctx context.Context, payload *goastorage.CreatePayload) (*goastorage.Package, error) {
	aipID, err := uuid.Parse(payload.AipID)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("invalid aip_id"))
	}

	objKey, err := uuid.Parse(payload.ObjectKey)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("invalid object_key"))
	}

	p := &goastorage.Package{
		Name:       payload.Name,
		AipID:      aipID,
		Status:     payload.Status,
		ObjectKey:  objKey,
		LocationID: payload.LocationID,
	}

	return s.storagePersistence.CreateAIP(ctx, p)
}

func (s *serviceImpl) Update(ctx context.Context, payload *goastorage.UpdatePayload) error {
	aipID, err := uuid.Parse(payload.AipID)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("cannot perform operation"))
	}

	signal := UploadDoneSignal{}
	workflowID := fmt.Sprintf("%s-%s", StorageUploadWorkflowName, aipID)
	err = s.tc.SignalWorkflow(ctx, workflowID, "", UploadDoneSignalName, signal)
	if err != nil {
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}
	// Update the package status to in_review
	err = s.UpdatePackageStatus(ctx, aipID, types.AIPStatusInReview)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("cannot update package status"))
	}

	return nil
}

func (s *serviceImpl) Download(ctx context.Context, payload *goastorage.DownloadPayload) ([]byte, error) {
	// This service method is unused, see the Download function instead which
	// makes use of http.ResponseWriter.
	return []byte{}, nil
}

func (s *serviceImpl) Locations(
	ctx context.Context,
	payload *goastorage.LocationsPayload,
) (goastorage.LocationCollection, error) {
	return s.storagePersistence.ListLocations(ctx)
}

func (s *serviceImpl) Move(ctx context.Context, payload *goastorage.MovePayload) error {
	aipID, err := uuid.Parse(payload.AipID)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("cannot perform operation"))
	}

	pkg, err := s.ReadPackage(ctx, aipID)
	if err != nil {
		return err
	}

	_, err = InitStorageMoveWorkflow(ctx, s.tc, &StorageMoveWorkflowRequest{
		AIPID:      pkg.AipID,
		LocationID: payload.LocationID,
		TaskQueue:  s.config.TaskQueue,
	})
	if err != nil {
		s.logger.Error(err, "error initializing move workflow")
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

func (s *serviceImpl) MoveStatus(
	ctx context.Context,
	payload *goastorage.MoveStatusPayload,
) (*goastorage.MoveStatusResult, error) {
	aipID, err := uuid.Parse(payload.AipID)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot perform operation"))
	}

	p, err := s.ReadPackage(ctx, aipID)
	if err != nil {
		return nil, err
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
	aipID, err := uuid.Parse(payload.AipID)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("cannot perform operation"))
	}

	return s.UpdatePackageStatus(ctx, aipID, types.AIPStatusRejected)
}

func (s *serviceImpl) Show(ctx context.Context, payload *goastorage.ShowPayload) (*goastorage.Package, error) {
	aipID, err := uuid.Parse(payload.AipID)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot perform operation"))
	}

	return s.ReadPackage(ctx, aipID)
}

func (s *serviceImpl) ReadPackage(ctx context.Context, aipID uuid.UUID) (*goastorage.Package, error) {
	return s.storagePersistence.ReadAIP(ctx, aipID)
}

func (s *serviceImpl) UpdatePackageStatus(ctx context.Context, aipID uuid.UUID, status types.AIPStatus) error {
	return s.storagePersistence.UpdateAIPStatus(ctx, aipID, status)
}

func (s *serviceImpl) UpdatePackageLocationID(ctx context.Context, aipID, locationID uuid.UUID) error {
	return s.storagePersistence.UpdateAIPLocationID(ctx, aipID, locationID)
}

// packageLocation returns the bucket and the key of the given package.
func (s *serviceImpl) packageLocation(ctx context.Context, p *goastorage.Package) (Location, string, error) {
	// Package is still in the internal processing bucket.
	if p.LocationID == nil || *p.LocationID == uuid.Nil {
		return s.internal, p.ObjectKey.String(), nil
	}

	location, err := s.Location(ctx, *p.LocationID)
	if err != nil {
		return nil, "", err
	}
	return location, p.AipID.String(), nil
}

func (s *serviceImpl) Delete(ctx context.Context, aipID uuid.UUID) error {
	pkg, err := s.ReadPackage(ctx, aipID)
	if err != nil {
		return err
	}

	location, key, err := s.packageLocation(ctx, pkg)
	if err != nil {
		return err
	}

	bucket, err := location.OpenBucket(ctx)
	if err != nil {
		return err
	}
	defer bucket.Close()

	return bucket.Delete(ctx, key)
}

func (s *serviceImpl) PackageReader(ctx context.Context, pkg *goastorage.Package) (*blob.Reader, error) {
	location, key, err := s.packageLocation(ctx, pkg)
	if err != nil {
		return nil, err
	}

	bucket, err := location.OpenBucket(ctx)
	if err != nil {
		return nil, err
	}
	defer bucket.Close()

	reader, err := bucket.NewReader(ctx, key, nil)
	if err != nil {
		return nil, err
	}

	return reader, nil
}

func (s *serviceImpl) AddLocation(
	ctx context.Context,
	payload *goastorage.AddLocationPayload,
) (res *goastorage.AddLocationResult, err error) {
	source := types.NewLocationSource(payload.Source)
	purpose := types.NewLocationPurpose(payload.Purpose)
	UUID := uuid.Must(uuid.NewRandomFromReader(s.rander))

	var config types.LocationConfig
	switch c := payload.Config.(type) {
	case *goastorage.URLConfig:
		config.Value = c.ConvertToURLConfig()
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
		UUID:        UUID,
	}, &config)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot persist location"))
	}

	return &goastorage.AddLocationResult{UUID: UUID.String()}, nil
}

// ReadLocation retrieves location data for location UUID from the persistence
// layer.
//
// A `goastorage.LocationNotFound` error is returned if location UUID doesn't
// exist. A `goa.ServiceError` "not_available" error is returned for all other
// errors.
func (s *serviceImpl) ReadLocation(ctx context.Context, UUID uuid.UUID) (*goastorage.Location, error) {
	return s.storagePersistence.ReadLocation(ctx, UUID)
}

func (s *serviceImpl) ShowLocation(
	ctx context.Context,
	payload *goastorage.ShowLocationPayload,
) (*goastorage.Location, error) {
	locationID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot perform operation"))
	}

	return s.ReadLocation(ctx, locationID)
}

func (s *serviceImpl) LocationPackages(
	ctx context.Context,
	payload *goastorage.LocationPackagesPayload,
) (goastorage.PackageCollection, error) {
	locationID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot perform operation"))
	}

	pkgs, err := s.storagePersistence.LocationAIPs(ctx, locationID)
	if err != nil {
		return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return pkgs, nil
}
