package package_

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
	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
)

var ErrBulkStatusUnavailable = errors.New("bulk status unavailable")

// GoaWrapper returns a packageImpl wrapper that implements
// goapackage.Service. It can handle types that are specific to the Goa API.
type goaWrapper struct {
	*packageImpl
}

var _ goapackage.Service = (*goaWrapper)(nil)

var (
	ErrUnauthorized error = goapackage.Unauthorized("Unauthorized")
	ErrForbidden    error = goapackage.Forbidden("Forbidden")
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
	payload *goapackage.MonitorRequestPayload,
) (*goapackage.MonitorRequestResult, error) {
	res := &goapackage.MonitorRequestResult{}

	ticket, err := w.ticketProvider.Request(ctx)
	if err != nil {
		w.logger.Error(err, "failed to request ticket")
		return nil, goapackage.MakeNotAvailable(errors.New("cannot perform operation"))
	}
	// Do not set cookie unless a ticket is provided.
	if ticket != "" {
		res.Ticket = &ticket
	}

	return res, nil
}

// Monitor package activity. It implements goapackage.Service.
func (w *goaWrapper) Monitor(
	ctx context.Context,
	payload *goapackage.MonitorPayload,
	stream goapackage.MonitorServerStream,
) error {
	defer stream.Close()

	// Verify the ticket.
	if err := w.ticketProvider.Check(ctx, payload.Ticket); err != nil {
		w.logger.V(1).Info("failed to check ticket", "err", err)
		return goapackage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	// Subscribe to the event service.
	sub, err := w.evsvc.Subscribe(ctx)
	if err != nil {
		return err
	}
	defer sub.Close()

	// Say hello to be nice.
	event := &goapackage.MonitorPingEvent{Message: ref.New("Hello")}
	if err := stream.Send(&goapackage.MonitorEvent{Event: event}); err != nil {
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
			event := &goapackage.MonitorPingEvent{Message: ref.New("Ping")}
			if err := stream.Send(&goapackage.MonitorEvent{Event: event}); err != nil {
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

// List all stored packages. It implements goapackage.Service.
func (w *goaWrapper) List(ctx context.Context, payload *goapackage.ListPayload) (*goapackage.EnduroPackages, error) {
	if payload == nil {
		payload = &goapackage.ListPayload{}
	}

	pf, err := listPayloadToPackageFilter(payload)
	if err != nil {
		return nil, err
	}

	r, pg, err := w.perSvc.ListPackages(ctx, pf)
	if err != nil {
		return nil, goapackage.MakeInternalError(err)
	}

	items := make([]*goapackage.EnduroStoredPackage, len(r))
	for i, pkg := range r {
		items[i] = pkg.Goa()
	}

	res := &goapackage.EnduroPackages{
		Items: items,
		Page:  pg.Goa(),
	}

	return res, nil
}

// Show package by ID. It implements goapackage.Service.
func (w *goaWrapper) Show(
	ctx context.Context,
	payload *goapackage.ShowPayload,
) (*goapackage.EnduroStoredPackage, error) {
	c, err := w.read(ctx, payload.ID)
	if err == sql.ErrNoRows {
		return nil, &goapackage.PackageNotFound{ID: payload.ID, Message: "package not found"}
	} else if err != nil {
		return nil, goapackage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return c.Goa(), nil
}

func (w *goaWrapper) PreservationActions(
	ctx context.Context,
	payload *goapackage.PreservationActionsPayload,
) (*goapackage.EnduroPackagePreservationActions, error) {
	goapkg, err := w.Show(ctx, &goapackage.ShowPayload{ID: payload.ID})
	if err != nil {
		return nil, err
	}

	query := "SELECT id, workflow_id, type, status, CONVERT_TZ(started_at, @@session.time_zone, '+00:00') AS started_at, CONVERT_TZ(completed_at, @@session.time_zone, '+00:00') AS completed_at FROM preservation_action WHERE package_id = ? ORDER BY started_at DESC"
	args := []interface{}{goapkg.ID}

	rows, err := w.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying the database: %w", err)
	}
	defer rows.Close()

	preservation_actions := []*goapackage.EnduroPackagePreservationAction{}
	for rows.Next() {
		pa := datatypes.PreservationAction{}
		if err := rows.StructScan(&pa); err != nil {
			return nil, fmt.Errorf("error scanning database result: %w", err)
		}
		goapa := preservationActionToGoa(&pa)

		ptQuery := "SELECT id, task_id, name, status, CONVERT_TZ(started_at, @@session.time_zone, '+00:00') AS started_at, CONVERT_TZ(completed_at, @@session.time_zone, '+00:00') AS completed_at, note FROM preservation_task WHERE preservation_action_id = ?"
		ptQueryArgs := []interface{}{pa.ID}

		ptRows, err := w.db.QueryxContext(ctx, ptQuery, ptQueryArgs...)
		if err != nil {
			return nil, fmt.Errorf("error querying the database: %w", err)
		}
		defer ptRows.Close()

		preservation_tasks := []*goapackage.EnduroPackagePreservationTask{}
		for ptRows.Next() {
			pt := datatypes.PreservationTask{}
			if err := ptRows.StructScan(&pt); err != nil {
				return nil, fmt.Errorf("error scanning database result: %w", err)
			}
			preservation_tasks = append(preservation_tasks, preservationTaskToGoa(&pt))
		}

		goapa.Tasks = preservation_tasks
		preservation_actions = append(preservation_actions, goapa)
	}

	result := &goapackage.EnduroPackagePreservationActions{
		Actions: preservation_actions,
	}

	return result, nil
}

func (w *goaWrapper) Confirm(ctx context.Context, payload *goapackage.ConfirmPayload) error {
	goapkg, err := w.Show(ctx, &goapackage.ShowPayload{ID: payload.ID})
	if err != nil {
		return err
	}

	signal := ReviewPerformedSignal{
		Accepted:   true,
		LocationID: &payload.LocationID,
	}
	err = w.tc.SignalWorkflow(ctx, *goapkg.WorkflowID, "", ReviewPerformedSignalName, signal)
	if err != nil {
		return goapackage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

func (w *goaWrapper) Reject(ctx context.Context, payload *goapackage.RejectPayload) error {
	goapkg, err := w.Show(ctx, &goapackage.ShowPayload{ID: payload.ID})
	if err != nil {
		return err
	}

	signal := ReviewPerformedSignal{
		Accepted: false,
	}
	err = w.tc.SignalWorkflow(ctx, *goapkg.WorkflowID, "", ReviewPerformedSignalName, signal)
	if err != nil {
		return goapackage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

func (w *goaWrapper) Move(ctx context.Context, payload *goapackage.MovePayload) error {
	goapkg, err := w.Show(ctx, &goapackage.ShowPayload{ID: payload.ID})
	if err != nil {
		return err
	}
	if payload.ID > math.MaxInt {
		return goapackage.MakeNotValid(errors.New("invalid ID"))
	}

	_, err = InitMoveWorkflow(ctx, w.tc, &MoveWorkflowRequest{
		ID:         int(payload.ID), // #nosec G115 -- range validated.
		AIPID:      *goapkg.AipID,
		LocationID: payload.LocationID,
		TaskQueue:  w.taskQueue,
	})
	if err != nil {
		w.logger.Error(err, "error initializing move workflow")
		return goapackage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

func (w *goaWrapper) MoveStatus(
	ctx context.Context,
	payload *goapackage.MoveStatusPayload,
) (*goapackage.MoveStatusResult, error) {
	goapkg, err := w.Show(ctx, &goapackage.ShowPayload{ID: payload.ID})
	if err != nil {
		return nil, goapackage.MakeFailedDependency(errors.New("cannot perform operation"))
	}
	if goapkg.AipID == nil {
		return nil, goapackage.MakeFailedDependency(errors.New("cannot perform operation"))
	}

	workflowID := fmt.Sprintf("%s-%s", MoveWorkflowName, *goapkg.AipID)
	resp, err := w.tc.DescribeWorkflowExecution(ctx, workflowID, "")
	if err != nil {
		return nil, goapackage.MakeFailedDependency(errors.New("cannot perform operation"))
	}

	var done bool
	switch resp.WorkflowExecutionInfo.Status {
	case
		temporalapi_enums.WORKFLOW_EXECUTION_STATUS_FAILED,
		temporalapi_enums.WORKFLOW_EXECUTION_STATUS_CANCELED,
		temporalapi_enums.WORKFLOW_EXECUTION_STATUS_TERMINATED,
		temporalapi_enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
		return nil, goapackage.MakeFailedDependency(errors.New("cannot perform operation"))
	case temporalapi_enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		done = true
	case temporalapi_enums.WORKFLOW_EXECUTION_STATUS_RUNNING:
		done = false
	}

	return &goapackage.MoveStatusResult{Done: done}, nil
}
