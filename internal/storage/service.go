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
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

var SubmitURLExpirationTime = 15 * time.Minute

// Service provides an interface for persisting storage data.
type Service interface {
	goastorage.Service

	// Used from workflow activities.
	Location(ctx context.Context, locationID uuid.UUID) (Location, error)
	ReadAip(ctx context.Context, aipID uuid.UUID) (*goastorage.AIP, error)
	UpdateAipStatus(ctx context.Context, aipID uuid.UUID, status enums.AIPStatus) error
	UpdateAipLocationID(ctx context.Context, aipID, locationID uuid.UUID) error
	DeleteAip(ctx context.Context, aipID uuid.UUID) (err error)

	// Both.
	AipReader(ctx context.Context, aip *goastorage.AIP) (*blob.Reader, error)
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

func (s *serviceImpl) SubmitAip(
	ctx context.Context,
	payload *goastorage.SubmitAipPayload,
) (*goastorage.SubmitAIPResult, error) {
	aipID, err := uuid.Parse(payload.UUID)
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

	_, err = s.storagePersistence.CreateAIP(ctx, &goastorage.AIP{
		Name:      payload.Name,
		UUID:      aipID,
		ObjectKey: objectKey,
	})
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot create AIP"))
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

	result := &goastorage.SubmitAIPResult{
		URL: url,
	}
	return result, nil
}

func (s *serviceImpl) CreateAip(ctx context.Context, payload *goastorage.CreateAipPayload) (*goastorage.AIP, error) {
	aipID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("invalid aip_id"))
	}

	objKey, err := uuid.Parse(payload.ObjectKey)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("invalid object_key"))
	}

	p := &goastorage.AIP{
		Name:       payload.Name,
		UUID:       aipID,
		Status:     payload.Status,
		ObjectKey:  objKey,
		LocationID: payload.LocationID,
	}

	return s.storagePersistence.CreateAIP(ctx, p)
}

func (s *serviceImpl) UpdateAip(ctx context.Context, payload *goastorage.UpdateAipPayload) error {
	aipID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("cannot perform operation"))
	}

	signal := UploadDoneSignal{}
	workflowID := fmt.Sprintf("%s-%s", StorageUploadWorkflowName, aipID)
	err = s.tc.SignalWorkflow(ctx, workflowID, "", UploadDoneSignalName, signal)
	if err != nil {
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}
	// Update the AIP status to in_review
	err = s.UpdateAipStatus(ctx, aipID, enums.AIPStatusInReview)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("cannot update AIP status"))
	}

	return nil
}

func (s *serviceImpl) DownloadAip(ctx context.Context, payload *goastorage.DownloadAipPayload) ([]byte, error) {
	// This service method is unused, see the Download function instead which
	// makes use of http.ResponseWriter.
	return []byte{}, nil
}

func (s *serviceImpl) ListLocations(
	ctx context.Context,
	payload *goastorage.ListLocationsPayload,
) (goastorage.LocationCollection, error) {
	return s.storagePersistence.ListLocations(ctx)
}

func (s *serviceImpl) MoveAip(ctx context.Context, payload *goastorage.MoveAipPayload) error {
	aipID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("cannot perform operation"))
	}

	aip, err := s.ReadAip(ctx, aipID)
	if err != nil {
		return err
	}

	_, err = InitStorageMoveWorkflow(ctx, s.tc, &StorageMoveWorkflowRequest{
		AIPID:      aip.UUID,
		LocationID: payload.LocationID,
		TaskQueue:  s.config.TaskQueue,
	})
	if err != nil {
		s.logger.Error(err, "error initializing move workflow")
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

func (s *serviceImpl) MoveAipStatus(
	ctx context.Context,
	payload *goastorage.MoveAipStatusPayload,
) (*goastorage.MoveStatusResult, error) {
	aipID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot perform operation"))
	}

	p, err := s.ReadAip(ctx, aipID)
	if err != nil {
		return nil, err
	}

	resp, err := s.tc.DescribeWorkflowExecution(ctx, fmt.Sprintf("%s-%s", StorageMoveWorkflowName, p.UUID), "")
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

func (s *serviceImpl) RejectAip(ctx context.Context, payload *goastorage.RejectAipPayload) error {
	aipID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("cannot perform operation"))
	}

	return s.UpdateAipStatus(ctx, aipID, enums.AIPStatusRejected)
}

func (s *serviceImpl) ShowAip(ctx context.Context, payload *goastorage.ShowAipPayload) (*goastorage.AIP, error) {
	aipID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot perform operation"))
	}

	return s.ReadAip(ctx, aipID)
}

func (s *serviceImpl) ReadAip(ctx context.Context, aipID uuid.UUID) (*goastorage.AIP, error) {
	return s.storagePersistence.ReadAIP(ctx, aipID)
}

func (s *serviceImpl) UpdateAipStatus(ctx context.Context, aipID uuid.UUID, status enums.AIPStatus) error {
	return s.storagePersistence.UpdateAIPStatus(ctx, aipID, status)
}

func (s *serviceImpl) UpdateAipLocationID(ctx context.Context, aipID, locationID uuid.UUID) error {
	return s.storagePersistence.UpdateAIPLocationID(ctx, aipID, locationID)
}

// aipLocation returns the bucket and the key of the given AIP.
func (s *serviceImpl) aipLocation(ctx context.Context, p *goastorage.AIP) (Location, string, error) {
	// AIP is still in the internal processing bucket.
	if p.LocationID == nil || *p.LocationID == uuid.Nil {
		return s.internal, p.ObjectKey.String(), nil
	}

	location, err := s.Location(ctx, *p.LocationID)
	if err != nil {
		return nil, "", err
	}
	return location, p.UUID.String(), nil
}

func (s *serviceImpl) DeleteAip(ctx context.Context, aipID uuid.UUID) error {
	aip, err := s.ReadAip(ctx, aipID)
	if err != nil {
		return err
	}

	location, key, err := s.aipLocation(ctx, aip)
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

func (s *serviceImpl) AipReader(ctx context.Context, a *goastorage.AIP) (*blob.Reader, error) {
	location, key, err := s.aipLocation(ctx, a)
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

func (s *serviceImpl) CreateLocation(
	ctx context.Context,
	payload *goastorage.CreateLocationPayload,
) (res *goastorage.CreateLocationResult, err error) {
	purpose, err := enums.ParseLocationPurposeWithDefault(payload.Purpose)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("purpose: invalid value"))
	}
	source, err := enums.ParseLocationSourceWithDefault(payload.Source)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("source: invalid value"))
	}

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

	return &goastorage.CreateLocationResult{UUID: UUID.String()}, nil
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

func (s *serviceImpl) ListLocationAips(
	ctx context.Context,
	payload *goastorage.ListLocationAipsPayload,
) (goastorage.AIPCollection, error) {
	locationID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot perform operation"))
	}

	aips, err := s.storagePersistence.LocationAIPs(ctx, locationID)
	if err != nil {
		return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return aips, nil
}
