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
	"github.com/artefactual-sdps/enduro/internal/auditlog"
	"github.com/artefactual-sdps/enduro/internal/event"
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
	ReadLocation(ctx context.Context, locationID uuid.UUID) (*goastorage.Location, error)
	CreateWorkflow(context.Context, *types.Workflow) error
	UpdateWorkflow(context.Context, int, persistence.WorkflowUpdater) (*types.Workflow, error)
	CreateTask(context.Context, *types.Task) error
	UpdateTask(context.Context, int, persistence.TaskUpdater) (*types.Task, error)
	CreateDeletionRequest(context.Context, *types.DeletionRequest) error
	UpdateDeletionRequest(context.Context, int, persistence.DeletionRequestUpdater) (*types.DeletionRequest, error)
	ReadAipPendingDeletionRequest(ctx context.Context, aipID uuid.UUID) (*types.DeletionRequest, error)

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

	// Storage event service.
	evsvc event.Service[*goastorage.StorageEvent]

	// Token verifier.
	tokenVerifier auth.TokenVerifier

	// Ticket provider.
	ticketProvider auth.TicketProvider

	// Random number generator
	rander io.Reader

	// Audit event logger.
	auditLogger *auditlog.Logger
}

var _ Service = (*serviceImpl)(nil)

var (
	ErrUnauthorized  error = goastorage.Unauthorized("Unauthorized")
	ErrForbidden     error = goastorage.Forbidden("Forbidden")
	ErrInternalError error = goastorage.MakeInternalError(errors.New("internal error"))
)

func NewService(
	logger logr.Logger,
	config Config,
	storagePersistence persistence.Storage,
	tc temporalsdk_client.Client,
	evsvc event.Service[*goastorage.StorageEvent],
	tokenVerifier auth.TokenVerifier,
	ticketProvider auth.TicketProvider,
	rander io.Reader,
	auditLogger *auditlog.Logger,
) (s *serviceImpl, err error) {
	s = &serviceImpl{
		logger:             logger,
		tc:                 tc,
		config:             config,
		storagePersistence: storagePersistence,
		evsvc:              evsvc,
		tokenVerifier:      tokenVerifier,
		ticketProvider:     ticketProvider,
		rander:             rander,
		auditLogger:        auditLogger,
	}

	if s.rander == nil {
		s.rander = rand.Reader
	}
	if s.auditLogger == nil {
		s.auditLogger = auditlog.Discard()
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

	aip, err := s.storagePersistence.CreateAIP(ctx, &goastorage.AIP{
		Name:      payload.Name,
		UUID:      aipID,
		ObjectKey: objectKey,
	})
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot create AIP"))
	}

	PublishEvent(ctx, s.evsvc, &goastorage.AIPCreatedEvent{
		UUID: aipID,
		Item: aip,
	})

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
		Name:         payload.Name,
		UUID:         aipID,
		Status:       payload.Status,
		ObjectKey:    objKey,
		LocationUUID: payload.LocationUUID,
	}

	aip, err := s.storagePersistence.CreateAIP(ctx, p)
	if err != nil {
		return nil, err
	}

	PublishEvent(ctx, s.evsvc, &goastorage.AIPCreatedEvent{
		UUID: aipID,
		Item: aip,
	})

	return aip, nil
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

	// Update AIP status to pending.
	err = s.UpdateAipStatus(ctx, aipID, enums.AIPStatusPending)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("cannot update AIP status"))
	}

	return nil
}

func (s *serviceImpl) ListLocations(
	ctx context.Context,
	payload *goastorage.ListLocationsPayload,
) (goastorage.LocationCollection, error) {
	return s.storagePersistence.ListLocations(ctx)
}

func (s *serviceImpl) ListAips(
	ctx context.Context,
	payload *goastorage.ListAipsPayload,
) (*goastorage.AIPs, error) {
	return s.storagePersistence.ListAIPs(ctx, payload)
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
		LocationID: payload.LocationUUID,
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

	return s.UpdateAipStatus(ctx, aipID, enums.AIPStatusDeleted)
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
	err := s.storagePersistence.UpdateAIPStatus(ctx, aipID, status)
	if err != nil {
		return err
	}

	PublishEvent(ctx, s.evsvc, &goastorage.AIPStatusUpdatedEvent{
		UUID:   aipID,
		Status: status.String(),
	})

	return nil
}

func (s *serviceImpl) UpdateAipLocationID(ctx context.Context, aipID, locationID uuid.UUID) error {
	err := s.storagePersistence.UpdateAIPLocationID(ctx, aipID, locationID)
	if err != nil {
		return err
	}

	PublishEvent(ctx, s.evsvc, &goastorage.AIPLocationUpdatedEvent{
		UUID:         aipID,
		LocationUUID: locationID,
	})

	return nil
}

// aipLocation returns the bucket and the key of the given AIP.
func (s *serviceImpl) aipLocation(ctx context.Context, a *goastorage.AIP) (Location, string, error) {
	// AIP is still in the internal processing bucket.
	if a.LocationUUID == nil || *a.LocationUUID == uuid.Nil {
		return s.internal, a.ObjectKey.String(), nil
	}

	location, err := s.Location(ctx, *a.LocationUUID)
	if err != nil {
		return nil, "", err
	}
	return location, a.UUID.String(), nil
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

func (s *serviceImpl) ListAipWorkflows(
	ctx context.Context,
	payload *goastorage.ListAipWorkflowsPayload,
) (*goastorage.AIPWorkflows, error) {
	f := persistence.WorkflowFilter{}

	aipUUID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("UUID: invalid value"))
	}
	f.AIPUUID = &aipUUID

	if payload.Status != nil {
		s, err := enums.ParseWorkflowStatus(*payload.Status)
		if err != nil {
			return nil, goastorage.MakeNotValid(errors.New("status: invalid value"))
		}
		f.Status = &s
	}

	if payload.Type != nil {
		t, err := enums.ParseWorkflowType(*payload.Type)
		if err != nil {
			return nil, goastorage.MakeNotValid(errors.New("type: invalid value"))
		}
		f.Type = &t
	}

	workflows, err := s.storagePersistence.ListWorkflows(ctx, &f)
	if err != nil {
		return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return &goastorage.AIPWorkflows{Workflows: workflows}, nil
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

	location, err := s.storagePersistence.CreateLocation(ctx, &goastorage.Location{
		Name:        payload.Name,
		Description: payload.Description,
		Source:      source.String(),
		Purpose:     purpose.String(),
		UUID:        UUID,
	}, &config)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("cannot persist location"))
	}

	PublishEvent(ctx, s.evsvc, &goastorage.LocationCreatedEvent{
		UUID: UUID,
		Item: location,
	})

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

func (svc *serviceImpl) CreateWorkflow(ctx context.Context, w *types.Workflow) error {
	err := svc.storagePersistence.CreateWorkflow(ctx, w)
	if err != nil {
		return err
	}

	PublishEvent(ctx, svc.evsvc, &goastorage.AIPWorkflowCreatedEvent{
		UUID: w.UUID,
		Item: svc.workflowToGoa(w),
	})

	return nil
}

func (svc *serviceImpl) UpdateWorkflow(
	ctx context.Context,
	id int,
	upd persistence.WorkflowUpdater,
) (*types.Workflow, error) {
	workflow, err := svc.storagePersistence.UpdateWorkflow(ctx, id, upd)
	if err != nil {
		return nil, err
	}

	PublishEvent(ctx, svc.evsvc, &goastorage.AIPWorkflowUpdatedEvent{
		UUID: workflow.UUID,
		Item: svc.workflowToGoa(workflow),
	})

	return workflow, nil
}

func (svc *serviceImpl) CreateTask(ctx context.Context, t *types.Task) error {
	err := svc.storagePersistence.CreateTask(ctx, t)
	if err != nil {
		return err
	}

	PublishEvent(ctx, svc.evsvc, &goastorage.AIPTaskCreatedEvent{
		UUID: t.UUID,
		Item: svc.taskToGoa(t),
	})

	return nil
}

func (svc *serviceImpl) UpdateTask(
	ctx context.Context,
	id int,
	upd persistence.TaskUpdater,
) (*types.Task, error) {
	task, err := svc.storagePersistence.UpdateTask(ctx, id, upd)
	if err != nil {
		return nil, err
	}

	PublishEvent(ctx, svc.evsvc, &goastorage.AIPTaskUpdatedEvent{
		UUID: task.UUID,
		Item: svc.taskToGoa(task),
	})

	return task, nil
}

func (svc *serviceImpl) CreateDeletionRequest(ctx context.Context, dr *types.DeletionRequest) error {
	if err := svc.storagePersistence.CreateDeletionRequest(ctx, dr); err != nil {
		return err
	}

	PublishEvent(ctx, svc.evsvc, &goastorage.AIPDeletionRequestCreatedEvent{
		UUID: dr.UUID,
		Item: svc.deletionRequestToGoa(dr),
	})
	svc.auditLogger.Log(ctx, deletionRequestAuditEvent(dr))

	return nil
}

func (svc *serviceImpl) UpdateDeletionRequest(
	ctx context.Context,
	id int,
	upd persistence.DeletionRequestUpdater,
) (*types.DeletionRequest, error) {
	dr, err := svc.storagePersistence.UpdateDeletionRequest(ctx, id, upd)
	if err != nil {
		return nil, err
	}
	svc.auditLogger.Log(ctx, deletionRequestAuditEvent(dr))

	return dr, nil
}

func (svc *serviceImpl) ReadAipPendingDeletionRequest(
	ctx context.Context,
	aipID uuid.UUID,
) (*types.DeletionRequest, error) {
	return svc.storagePersistence.ReadAipPendingDeletionRequest(ctx, aipID)
}
