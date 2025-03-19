package ingest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"go.artefactual.dev/tools/ref"
	temporalapi_enums "go.temporal.io/api/enums/v1"
	"goa.design/goa/v3/security"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
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

func (w *goaWrapper) MonitorRequest(
	ctx context.Context,
	payload *goaingest.MonitorRequestPayload,
) (*goaingest.MonitorRequestResult, error) {
	res := &goaingest.MonitorRequestResult{}

	ticket, err := w.ticketProvider.Request(ctx)
	if err != nil {
		w.logger.Error(err, "failed to request ticket")
		return nil, goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}
	// Do not set cookie unless a ticket is provided.
	if ticket != "" {
		res.Ticket = &ticket
	}

	return res, nil
}

// Monitor ingest activity. It implements goaingest.Service.
func (w *goaWrapper) Monitor(
	ctx context.Context,
	payload *goaingest.MonitorPayload,
	stream goaingest.MonitorServerStream,
) error {
	defer stream.Close()

	// Verify the ticket.
	if err := w.ticketProvider.Check(ctx, payload.Ticket); err != nil {
		w.logger.V(1).Info("failed to check ticket", "err", err)
		return goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	// Subscribe to the event service.
	sub, err := w.evsvc.Subscribe(ctx)
	if err != nil {
		return err
	}
	defer sub.Close()

	// Say hello to be nice.
	event := &goaingest.MonitorPingEvent{Message: ref.New("Hello")}
	if err := stream.Send(&goaingest.MonitorEvent{Event: event}); err != nil {
		return err
	}

	// We'll use this ticker to ping the client once in a while to detect stale
	// connections. I'm not entirely sure this is needed, it may depend on the
	// client or the various middlewares.
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			event := &goaingest.MonitorPingEvent{Message: ref.New("Ping")}
			if err := stream.Send(&goaingest.MonitorEvent{Event: event}); err != nil {
				return nil
			}

		case event, ok := <-sub.C():
			if !ok {
				return nil
			}

			if err := stream.Send(event); err != nil {
				return err
			}
		}
	}
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
	c, err := w.read(ctx, payload.ID)
	if err == sql.ErrNoRows {
		return nil, &goaingest.SIPNotFound{ID: payload.ID, Message: "SIP not found"}
	} else if err != nil {
		return nil, goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return c.Goa(), nil
}

func (w *goaWrapper) ListSipWorkflows(
	ctx context.Context,
	payload *goaingest.ListSipWorkflowsPayload,
) (*goaingest.SIPWorkflows, error) {
	goasip, err := w.ShowSip(ctx, &goaingest.ShowSipPayload{ID: payload.ID})
	if err != nil {
		return nil, err
	}

	query := "SELECT id, workflow_id, type, status, CONVERT_TZ(started_at, @@session.time_zone, '+00:00') AS started_at, CONVERT_TZ(completed_at, @@session.time_zone, '+00:00') AS completed_at FROM workflow WHERE sip_id = ? ORDER BY started_at DESC"
	args := []interface{}{goasip.ID}

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
		goaworkflow := workflowToGoa(&workflow)

		ptQuery := "SELECT id, task_id, name, status, CONVERT_TZ(started_at, @@session.time_zone, '+00:00') AS started_at, CONVERT_TZ(completed_at, @@session.time_zone, '+00:00') AS completed_at, note FROM task WHERE workflow_id = ?"
		ptQueryArgs := []interface{}{workflow.ID}

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
	goasip, err := w.ShowSip(ctx, &goaingest.ShowSipPayload{ID: payload.ID})
	if err != nil {
		return err
	}

	signal := ReviewPerformedSignal{
		Accepted:   true,
		LocationID: &payload.LocationID,
	}
	err = w.tc.SignalWorkflow(ctx, *goasip.WorkflowID, "", ReviewPerformedSignalName, signal)
	if err != nil {
		return goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

func (w *goaWrapper) RejectSip(ctx context.Context, payload *goaingest.RejectSipPayload) error {
	goasip, err := w.ShowSip(ctx, &goaingest.ShowSipPayload{ID: payload.ID})
	if err != nil {
		return err
	}

	signal := ReviewPerformedSignal{
		Accepted: false,
	}
	err = w.tc.SignalWorkflow(ctx, *goasip.WorkflowID, "", ReviewPerformedSignalName, signal)
	if err != nil {
		return goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

func (w *goaWrapper) MoveSip(ctx context.Context, payload *goaingest.MoveSipPayload) error {
	goasip, err := w.ShowSip(ctx, &goaingest.ShowSipPayload{ID: payload.ID})
	if err != nil {
		return err
	}
	if payload.ID > math.MaxInt {
		return goaingest.MakeNotValid(errors.New("invalid ID"))
	}

	_, err = InitMoveWorkflow(ctx, w.tc, &MoveWorkflowRequest{
		ID:         int(payload.ID), // #nosec G115 -- range validated.
		AIPID:      *goasip.AipID,
		LocationID: payload.LocationID,
		TaskQueue:  w.taskQueue,
	})
	if err != nil {
		w.logger.Error(err, "error initializing move workflow")
		return goaingest.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

func (w *goaWrapper) MoveSipStatus(
	ctx context.Context,
	payload *goaingest.MoveSipStatusPayload,
) (*goaingest.MoveStatusResult, error) {
	goasip, err := w.ShowSip(ctx, &goaingest.ShowSipPayload{ID: payload.ID})
	if err != nil {
		return nil, goaingest.MakeFailedDependency(errors.New("cannot perform operation"))
	}
	if goasip.AipID == nil {
		return nil, goaingest.MakeFailedDependency(errors.New("cannot perform operation"))
	}

	workflowID := fmt.Sprintf("%s-%s", MoveWorkflowName, *goasip.AipID)
	resp, err := w.tc.DescribeWorkflowExecution(ctx, workflowID, "")
	if err != nil {
		return nil, goaingest.MakeFailedDependency(errors.New("cannot perform operation"))
	}

	var done bool
	switch resp.WorkflowExecutionInfo.Status {
	case
		temporalapi_enums.WORKFLOW_EXECUTION_STATUS_FAILED,
		temporalapi_enums.WORKFLOW_EXECUTION_STATUS_CANCELED,
		temporalapi_enums.WORKFLOW_EXECUTION_STATUS_TERMINATED,
		temporalapi_enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
		return nil, goaingest.MakeFailedDependency(errors.New("cannot perform operation"))
	case temporalapi_enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		done = true
	case temporalapi_enums.WORKFLOW_EXECUTION_STATUS_RUNNING:
		done = false
	}

	return &goaingest.MoveStatusResult{Done: done}, nil
}
