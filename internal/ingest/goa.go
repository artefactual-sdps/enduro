package ingest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"
	"goa.design/goa/v3/security"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/sipsource"
)

// GoaWrapper returns a ingestImpl wrapper that implements
// goaingest.Service. It can handle types that are specific to the Goa API.
type goaWrapper struct {
	*ingestImpl
}

var _ goaingest.Service = (*goaWrapper)(nil)

var (
	ErrBulkStatusUnavailable error = errors.New("bulk status unavailable")
	ErrForbidden             error = goaingest.Forbidden("Forbidden")
	ErrUnauthorized          error = goaingest.Unauthorized("Unauthorized")
)

func (w *goaWrapper) JWTAuth(
	ctx context.Context,
	token string,
	scheme *security.JWTScheme,
) (context.Context, error) {
	claims, err := w.tokenVerifier.Verify(ctx, token)
	if err != nil {
		if !errors.Is(err, auth.ErrUnauthorized) {
			w.logger.V(1).Info("failed to verify token", "err", err)
		}
		return ctx, ErrUnauthorized
	}

	if !claims.CheckAttributes(scheme.RequiredScopes) {
		return ctx, ErrForbidden
	}

	ctx = auth.WithUserClaims(ctx, claims)

	return ctx, nil
}

// List all SIPs. It implements goaingest.Service.
func (w *goaWrapper) ListSips(ctx context.Context, payload *goaingest.ListSipsPayload) (*goaingest.SIPs, error) {
	if payload == nil {
		payload = &goaingest.ListSipsPayload{}
	}

	pf, err := listSipsPayloadToSIPFilter(payload)
	if err != nil {
		return nil, err
	}

	r, pg, err := w.perSvc.ListSIPs(ctx, pf)
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
func (w *goaWrapper) ShowSip(
	ctx context.Context,
	payload *goaingest.ShowSipPayload,
) (*goaingest.SIP, error) {
	sipUUID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return nil, goaingest.MakeNotValid(errors.New("invalid UUID"))
	}

	s, err := w.perSvc.ReadSIP(ctx, sipUUID)
	if err == persistence.ErrNotFound {
		return nil, &goaingest.SIPNotFound{UUID: payload.UUID, Message: "SIP not found"}
	} else if err != nil {
		return nil, goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return s.Goa(), nil
}

func (w *goaWrapper) ListSipWorkflows(
	ctx context.Context,
	payload *goaingest.ListSipWorkflowsPayload,
) (*goaingest.SIPWorkflows, error) {
	sipUUID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return nil, goaingest.MakeNotValid(errors.New("invalid UUID"))
	}

	s, err := w.perSvc.ReadSIP(ctx, sipUUID)
	if err == sql.ErrNoRows {
		return nil, &goaingest.SIPNotFound{UUID: payload.UUID, Message: "SIP not found"}
	} else if err != nil {
		return nil, goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	query := "SELECT id, uuid, temporal_id, type, status, CONVERT_TZ(started_at, @@session.time_zone, '+00:00') AS started_at, CONVERT_TZ(completed_at, @@session.time_zone, '+00:00') AS completed_at FROM workflow WHERE sip_id = ? ORDER BY started_at DESC"
	args := []any{s.ID}

	rows, err := w.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying the database: %w", err)
	}
	defer rows.Close()

	workflows := []*goaingest.SIPWorkflow{}
	for rows.Next() {
		workflow := datatypes.Workflow{}
		if err := rows.StructScan(&workflow); err != nil {
			return nil, fmt.Errorf("error scanning database result: %w", err)
		}
		workflow.SIPUUID = s.UUID
		goaworkflow := workflowToGoa(&workflow)

		ptQuery := "SELECT id, uuid, name, status, CONVERT_TZ(started_at, @@session.time_zone, '+00:00') AS started_at, CONVERT_TZ(completed_at, @@session.time_zone, '+00:00') AS completed_at, note FROM task WHERE workflow_id = ?"
		ptQueryArgs := []any{workflow.ID}

		ptRows, err := w.db.QueryxContext(ctx, ptQuery, ptQueryArgs...)
		if err != nil {
			return nil, fmt.Errorf("error querying the database: %w", err)
		}
		defer ptRows.Close()

		tasks := []*goaingest.SIPTask{}
		for ptRows.Next() {
			task := datatypes.Task{}
			if err := ptRows.StructScan(&task); err != nil {
				return nil, fmt.Errorf("error scanning database result: %w", err)
			}
			task.WorkflowUUID = workflow.UUID
			tasks = append(tasks, taskToGoa(&task))
		}

		goaworkflow.Tasks = tasks
		workflows = append(workflows, goaworkflow)
	}

	result := &goaingest.SIPWorkflows{
		Workflows: workflows,
	}

	return result, nil
}

func (w *goaWrapper) ConfirmSip(ctx context.Context, payload *goaingest.ConfirmSipPayload) error {
	goaworkflows, err := w.ListSipWorkflows(ctx, &goaingest.ListSipWorkflowsPayload{UUID: payload.UUID})
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
	err = w.tc.SignalWorkflow(ctx, goaworkflows.Workflows[0].TemporalID, "", ReviewPerformedSignalName, signal)
	if err != nil {
		return goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

func (w *goaWrapper) RejectSip(ctx context.Context, payload *goaingest.RejectSipPayload) error {
	goaworkflows, err := w.ListSipWorkflows(ctx, &goaingest.ListSipWorkflowsPayload{UUID: payload.UUID})
	if err != nil {
		return err
	}
	if goaworkflows == nil || len(goaworkflows.Workflows) == 0 || len(goaworkflows.Workflows) > 1 {
		return goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	signal := ReviewPerformedSignal{
		Accepted: false,
	}
	err = w.tc.SignalWorkflow(ctx, goaworkflows.Workflows[0].TemporalID, "", ReviewPerformedSignalName, signal)
	if err != nil {
		return goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

// List all SIPs. It implements goaingest.Service.
func (w *goaWrapper) ListUsers(ctx context.Context, payload *goaingest.ListUsersPayload) (*goaingest.Users, error) {
	if payload == nil {
		payload = &goaingest.ListUsersPayload{}
	}

	pf, err := listUsersPayloadToUserFilter(payload)
	if err != nil {
		return nil, err
	}

	r, pg, err := w.perSvc.ListUsers(ctx, pf)
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

func (w *goaWrapper) ListSipSourceObjects(
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

	page, err := w.sipSource.ListObjects(ctx, cursor, ref.DerefZero(payload.Limit))
	if err != nil {
		if errors.Is(err, sipsource.ErrMissingBucket) {
			return nil, goaingest.MakeNotFound(errors.New("SIP source not found"))
		}

		w.logger.Error(err, "Listing SIP source objects")
		return nil, goaingest.MakeInternalError(errors.New("internal error"))
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
