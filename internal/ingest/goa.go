package ingest

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"
	"goa.design/goa/v3/security"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/sipsource"
)

var (
	ErrBulkStatusUnavailable error = errors.New("bulk status unavailable")
	ErrForbidden             error = goaingest.Forbidden("Forbidden")
	ErrUnauthorized          error = goaingest.Unauthorized("Unauthorized")
	ErrInternalError         error = goaingest.MakeInternalError(errors.New("internal error"))
)

func (svc *ingestImpl) JWTAuth(
	ctx context.Context,
	token string,
	scheme *security.JWTScheme,
) (context.Context, error) {
	claims, err := svc.tokenVerifier.Verify(ctx, token)
	if err != nil {
		if !errors.Is(err, auth.ErrUnauthorized) {
			svc.logger.V(1).Info("failed to verify token", "err", err)
		}
		return ctx, ErrUnauthorized
	}

	if !claims.CheckAttributes(scheme.RequiredScopes) {
		return ctx, ErrForbidden
	}

	ctx = auth.WithUserClaims(ctx, claims)

	return ctx, nil
}

// AddSip ingests a new SIP from a SIP source.
func (svc *ingestImpl) AddSip(ctx context.Context, payload *goaingest.AddSipPayload) (*goaingest.AddSipResult, error) {
	if payload == nil {
		return nil, goaingest.MakeNotValid(errors.New("missing payload"))
	}

	sourceID, err := uuid.Parse(payload.SourceID)
	if err != nil {
		return nil, goaingest.MakeNotValid(errors.New("invalid SourceID"))
	}

	if payload.Key == "" {
		return nil, goaingest.MakeNotValid(errors.New("empty Key"))
	}

	claims, err := checkClaims(ctx)
	if err != nil {
		return nil, goaingest.MakeNotValid(err)
	}

	// Add a new SIP to the persistence layer.
	s := &datatypes.SIP{
		UUID:   uuid.Must(uuid.NewRandomFromReader(svc.rander)),
		Name:   payload.Key,
		Status: enums.SIPStatusQueued,
	}

	// If claims is nil, it means authentication is not enabled.
	if claims != nil {
		s.Uploader = &datatypes.User{
			UUID:    uuid.Must(uuid.NewRandomFromReader(svc.rander)),
			Email:   claims.Email,
			Name:    claims.Name,
			OIDCIss: claims.Iss,
			OIDCSub: claims.Sub,
		}
	}

	if err := svc.perSvc.CreateSIP(ctx, s); err != nil {
		svc.logger.Error(err, "add SIP")
		return nil, ErrInternalError
	}

	// Initialize the processing workflow.
	req := ProcessingWorkflowRequest{
		SIPUUID:         s.UUID,
		SIPSourceID:     sourceID,
		SIPName:         s.Name,
		Type:            enums.WorkflowTypeCreateAip,
		Key:             payload.Key,
		RetentionPeriod: svc.sipSource.RetentionPeriod(),
	}
	if err := InitProcessingWorkflow(ctx, svc.tc, svc.taskQueue, &req); err != nil {
		// Delete SIP from persistence.
		err = errors.Join(err, svc.perSvc.DeleteSIP(ctx, s.UUID))
		svc.logger.Error(err, "add SIP")
		return nil, ErrInternalError
	}

	PublishEvent(ctx, svc.evsvc, sipToCreatedEvent(s))
	svc.auditLogger.Log(ctx, sipIngestAuditEvent(s))

	svc.logger.V(1).Info(
		"Add SIP: started processing workflow from SIP source.",
		"object_key", payload.Key,
		"source_id", sourceID,
		"sip_id", s.UUID,
	)

	return &goaingest.AddSipResult{UUID: s.UUID.String()}, nil
}

// List all SIPs. It implements goaingest.Service.
func (svc *ingestImpl) ListSips(ctx context.Context, payload *goaingest.ListSipsPayload) (*goaingest.SIPs, error) {
	if payload == nil {
		payload = &goaingest.ListSipsPayload{}
	}

	pf, err := listSipsPayloadToSIPFilter(payload)
	if err != nil {
		return nil, err
	}

	r, pg, err := svc.perSvc.ListSIPs(ctx, pf)
	if err != nil {
		return nil, goaingest.MakeInternalError(err)
	}

	items := make([]*goaingest.SIP, len(r))
	for i, sip := range r {
		items[i] = sip.Goa()
	}

	res := &goaingest.SIPs{
		Items: items,
		Page:  pg.Goa(),
	}

	return res, nil
}

// Show SIP by ID. It implements goaingest.Service.
func (svc *ingestImpl) ShowSip(
	ctx context.Context,
	payload *goaingest.ShowSipPayload,
) (*goaingest.SIP, error) {
	sipUUID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return nil, goaingest.MakeNotValid(errors.New("invalid UUID"))
	}

	s, err := svc.perSvc.ReadSIP(ctx, sipUUID)
	if err == persistence.ErrNotFound {
		return nil, &goaingest.SIPNotFound{UUID: payload.UUID, Message: "SIP not found"}
	} else if err != nil {
		return nil, goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return s.Goa(), nil
}

func (svc *ingestImpl) ListSipWorkflows(
	ctx context.Context,
	payload *goaingest.ListSipWorkflowsPayload,
) (*goaingest.SIPWorkflows, error) {
	sipUUID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return nil, goaingest.MakeNotValid(errors.New("invalid UUID"))
	}

	if _, err := svc.perSvc.ReadSIP(ctx, sipUUID); err == persistence.ErrNotFound {
		return nil, &goaingest.SIPNotFound{UUID: payload.UUID, Message: "SIP not found"}
	} else if err != nil {
		return nil, goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	workflows, err := svc.perSvc.ListWorkflowsBySIP(ctx, sipUUID)
	if err != nil {
		return nil, goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	result := &goaingest.SIPWorkflows{
		Workflows: make([]*goaingest.SIPWorkflow, len(workflows)),
	}
	for i, w := range workflows {
		result.Workflows[i] = workflowToGoa(w)
	}

	return result, nil
}

func (svc *ingestImpl) ConfirmSip(ctx context.Context, payload *goaingest.ConfirmSipPayload) error {
	goaworkflows, err := svc.ListSipWorkflows(ctx, &goaingest.ListSipWorkflowsPayload{UUID: payload.UUID})
	if err != nil {
		return err
	}
	if goaworkflows == nil || len(goaworkflows.Workflows) == 0 || len(goaworkflows.Workflows) > 1 {
		return goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	signal := ReviewPerformedSignal{
		Accepted:   true,
		LocationID: &payload.LocationUUID,
	}
	err = svc.tc.SignalWorkflow(ctx, goaworkflows.Workflows[0].TemporalID, "", ReviewPerformedSignalName, signal)
	if err != nil {
		return goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

func (svc *ingestImpl) RejectSip(ctx context.Context, payload *goaingest.RejectSipPayload) error {
	goaworkflows, err := svc.ListSipWorkflows(ctx, &goaingest.ListSipWorkflowsPayload{UUID: payload.UUID})
	if err != nil {
		return err
	}
	if goaworkflows == nil || len(goaworkflows.Workflows) == 0 || len(goaworkflows.Workflows) > 1 {
		return goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	signal := ReviewPerformedSignal{
		Accepted: false,
	}
	err = svc.tc.SignalWorkflow(ctx, goaworkflows.Workflows[0].TemporalID, "", ReviewPerformedSignalName, signal)
	if err != nil {
		return goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

// List all SIPs. It implements goaingest.Service.
func (svc *ingestImpl) ListUsers(ctx context.Context, payload *goaingest.ListUsersPayload) (*goaingest.Users, error) {
	if payload == nil {
		payload = &goaingest.ListUsersPayload{}
	}

	pf, err := listUsersPayloadToUserFilter(payload)
	if err != nil {
		return nil, err
	}

	r, pg, err := svc.perSvc.ListUsers(ctx, pf)
	if err != nil {
		return nil, goaingest.MakeInternalError(err)
	}

	items := make([]*goaingest.User, len(r))
	for i, user := range r {
		items[i] = user.Goa()
	}

	res := &goaingest.Users{
		Items: items,
		Page:  pg.Goa(),
	}

	return res, nil
}

func (w *ingestImpl) ListSipSourceObjects(
	ctx context.Context,
	payload *goaingest.ListSipSourceObjectsPayload,
) (*goaingest.SIPSourceObjects, error) {
	// TODO: Use the payload.UUID to select a SIP source when we add support for
	// multiple SIP sources.
	if payload == nil {
		payload = &goaingest.ListSipSourceObjectsPayload{}
	}

	var cursor []byte
	if payload.Cursor != nil {
		cursor = []byte(*payload.Cursor)
	}

	page, err := w.sipSource.ListObjects(
		ctx,
		sipsource.ListOptions{
			Token: cursor,
			Limit: ref.DerefZero(payload.Limit),
			Sort:  sipsource.SortByModTime().Desc(),
		},
	)
	if err != nil {
		if errors.Is(err, sipsource.ErrInvalidSource) {
			return nil, goaingest.MakeNotFound(errors.New("SIP Source not found"))
		}
		if errors.Is(err, sipsource.ErrInvalidToken) {
			return nil, goaingest.MakeNotValid(errors.New("invalid cursor"))
		}

		w.logger.Error(err, "Listing SIP source objects")
		return nil, goaingest.MakeInternalError(errors.New("internal error"))
	}

	if page == nil {
		return &goaingest.SIPSourceObjects{}, nil
	}

	res := &goaingest.SIPSourceObjects{
		Objects: sipSourceObjectsToGoa(page.Objects),
		Limit:   page.Limit,
	}

	if page.NextToken != nil {
		res.Next = ref.New(string(page.NextToken))
	}

	return res, nil
}
