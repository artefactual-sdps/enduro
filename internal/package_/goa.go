package package_

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	temporalapi_common "go.temporal.io/api/common/v1"
	temporalapi_enums "go.temporal.io/api/enums/v1"
	temporalapi_serviceerror "go.temporal.io/api/serviceerror"
	temporalsdk_client "go.temporal.io/sdk/client"

	goapackage "github.com/artefactual-labs/enduro/internal/api/gen/package_"
	"github.com/artefactual-labs/enduro/internal/temporal"
)

var ErrBulkStatusUnavailable = errors.New("bulk status unavailable")

// GoaWrapper returns a packageImpl wrapper that implements
// goapackage.Service. It can handle types that are specific to the Goa API.
type goaWrapper struct {
	*packageImpl
}

var _ goapackage.Service = (*goaWrapper)(nil)

var patternMatchingCharReplacer = strings.NewReplacer(
	"%", "\\%",
	"_", "\\_",
)

// Monitor package activity. It implements goapackage.Service.
func (w *goaWrapper) Monitor(ctx context.Context, stream goapackage.MonitorServerStream) error {
	defer stream.Close()

	// Subscribe to the event service.
	sub, err := w.events.Subscribe(ctx)
	if err != nil {
		return err
	}
	defer sub.Close()

	// Say hello to be nice.
	if err := stream.Send(&goapackage.EnduroMonitorUpdate{Type: "hello"}); err != nil {
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
			if err := stream.Send(&goapackage.EnduroMonitorUpdate{Type: "ping"}); err != nil {
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
func (w *goaWrapper) List(ctx context.Context, payload *goapackage.ListPayload) (*goapackage.ListResult, error) {
	query := "SELECT id, name, workflow_id, run_id, aip_id, location, status, CONVERT_TZ(created_at, @@session.time_zone, '+00:00') AS created_at, CONVERT_TZ(started_at, @@session.time_zone, '+00:00') AS started_at, CONVERT_TZ(completed_at, @@session.time_zone, '+00:00') AS completed_at FROM package"
	args := []interface{}{}

	// We extract one extra item so we can tell the next cursor.
	const limit = 20
	const limitSQL = "21"

	conds := [][2]string{}

	if payload.Name != nil {
		name := patternMatchingCharReplacer.Replace(*payload.Name) + "%"
		args = append(args, name)
		conds = append(conds, [2]string{"AND", "name LIKE (?)"})
	}
	if payload.AipID != nil {
		args = append(args, payload.AipID)
		conds = append(conds, [2]string{"AND", "aip_id = (?)"})
	}
	if payload.Location != nil {
		args = append(args, payload.Location)
		conds = append(conds, [2]string{"AND", "location = (?)"})
	}
	if payload.Status != nil {
		args = append(args, NewStatus(*payload.Status))
		conds = append(conds, [2]string{"AND", "status = (?)"})
	}
	if payload.EarliestCreatedTime != nil {
		args = append(args, payload.EarliestCreatedTime)
		conds = append(conds, [2]string{"AND", "created_at >= (?)"})
	}
	if payload.LatestCreatedTime != nil {
		args = append(args, payload.LatestCreatedTime)
		conds = append(conds, [2]string{"AND", "created_at <= (?)"})
	}

	if payload.Cursor != nil {
		args = append(args, *payload.Cursor)
		conds = append(conds, [2]string{"AND", "id <= (?)"})
	}

	var where string
	for i, cond := range conds {
		if i == 0 {
			where = " WHERE " + cond[1]
			continue
		}
		where += fmt.Sprintf(" %s %s", cond[0], cond[1])
	}

	query += where + " ORDER BY id DESC LIMIT " + limitSQL

	query = w.db.Rebind(query)
	rows, err := w.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying the database: %w", err)
	}
	defer rows.Close()

	cols := []*goapackage.EnduroStoredPackage{}
	for rows.Next() {
		c := Package{}
		if err := rows.StructScan(&c); err != nil {
			return nil, fmt.Errorf("error scanning database result: %w", err)
		}
		cols = append(cols, c.Goa())
	}

	res := &goapackage.ListResult{
		Items: cols,
	}

	length := len(cols)
	if length > limit {
		last := cols[length-1]               // Capture last item.
		lastID := strconv.Itoa(int(last.ID)) // We also need its ID (cursor).
		res.Items = cols[:len(cols)-1]       // Remove it from the results.
		res.NextCursor = &lastID             // Populate cursor.
	}

	return res, nil
}

// Show package by ID. It implements goapackage.Service.
func (w *goaWrapper) Show(ctx context.Context, payload *goapackage.ShowPayload) (*goapackage.EnduroStoredPackage, error) {
	c, err := w.read(ctx, payload.ID)
	if err == sql.ErrNoRows {
		return nil, &goapackage.PackageNotfound{ID: payload.ID, Message: "not_found"}
	} else if err != nil {
		return nil, err
	}

	return c.Goa(), nil
}

// Delete package by ID. It implements goapackage.Service.
//
// TODO: return error if it's still running?
func (w *goaWrapper) Delete(ctx context.Context, payload *goapackage.DeletePayload) error {
	query := "DELETE FROM package WHERE id = (?)"

	query = w.db.Rebind(query)
	res, err := w.db.ExecContext(ctx, query, payload.ID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return &goapackage.PackageNotfound{ID: payload.ID, Message: "not_found"}
	}

	publishEvent(ctx, w.events, EventTypePackageDeleted, payload.ID)

	return nil
}

// Cancel package processing by ID. It implements goapackage.Service.
func (w *goaWrapper) Cancel(ctx context.Context, payload *goapackage.CancelPayload) error {
	goapkg, err := w.Show(ctx, &goapackage.ShowPayload{ID: payload.ID})
	if err != nil {
		return err
	}

	if err := w.tc.CancelWorkflow(ctx, *goapkg.WorkflowID, *goapkg.RunID); err != nil {
		// TODO: return custom errors
		return err
	}

	publishEvent(ctx, w.events, EventTypePackageUpdated, payload.ID)

	return nil
}

// Retry package processing by ID. It implements goapackage.Service.
//
// TODO: conceptually Temporal workflows should handle retries, i.e. retry could be part of workflow code too (e.g. signals, children, etc).
func (w *goaWrapper) Retry(ctx context.Context, payload *goapackage.RetryPayload) error {
	goapkg, err := w.Show(ctx, &goapackage.ShowPayload{ID: payload.ID})
	if err != nil {
		return err
	}

	execution := &temporalapi_common.WorkflowExecution{
		WorkflowId: *goapkg.WorkflowID,
		RunId:      *goapkg.RunID,
	}

	historyEvent, err := temporal.FirstHistoryEvent(ctx, w.tc, execution)
	if err != nil {
		return fmt.Errorf("error loading history of the previous workflow run: %w", err)
	}
	if historyEvent.GetEventType() != temporalapi_enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED {
		return fmt.Errorf("error loading history of the previous workflow run: initiator state not found")
	}

	input := historyEvent.GetWorkflowExecutionStartedEventAttributes().Input
	if len(input.Payloads) == 0 {
		return errors.New("error loading state of the previous workflow run")
	}
	eventPayload := input.Payloads[0]
	eventAttrs := eventPayload.GetData()

	req := &ProcessingWorkflowRequest{}
	if err := json.Unmarshal(eventAttrs, req); err != nil {
		return fmt.Errorf("error loading state of the previous workflow run: %w", err)
	}

	req.WorkflowID = *goapkg.WorkflowID
	req.PackageID = goapkg.ID
	if err := InitProcessingWorkflow(ctx, w.tc, req); err != nil {
		return fmt.Errorf("error starting the new workflow instance: %w", err)
	}

	publishEvent(ctx, w.events, EventTypePackageUpdated, payload.ID)

	return nil
}

func (w *goaWrapper) Workflow(ctx context.Context, payload *goapackage.WorkflowPayload) (res *goapackage.EnduroPackageWorkflowStatus, err error) {
	goapkg, err := w.Show(ctx, &goapackage.ShowPayload{ID: payload.ID})
	if err != nil {
		return nil, err
	}

	resp := &goapackage.EnduroPackageWorkflowStatus{
		History: []*goapackage.EnduroPackageWorkflowHistory{},
	}

	we, err := w.tc.DescribeWorkflowExecution(ctx, *goapkg.WorkflowID, *goapkg.RunID)
	if err != nil {
		switch err.(type) {
		case *temporalapi_serviceerror.NotFound:
			return nil, &goapackage.PackageNotfound{Message: "not_found"}
		default:
			return nil, fmt.Errorf("error looking up history: %v", err)
		}
	}

	status := "ACTIVE"
	if we.WorkflowExecutionInfo.Status != temporalapi_enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		status = we.WorkflowExecutionInfo.Status.String()
	}
	resp.Status = &status

	iter := w.tc.GetWorkflowHistory(ctx, *goapkg.WorkflowID, *goapkg.RunID, false, temporalapi_enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return nil, fmt.Errorf("error looking up history events: %v", err)
		}

		eventID := uint(event.EventId)
		eventType := event.EventType.String()
		resp.History = append(resp.History, &goapackage.EnduroPackageWorkflowHistory{
			ID:      &eventID,
			Type:    &eventType,
			Details: event,
		})
	}

	return resp, nil
}

func (w *goaWrapper) Bulk(ctx context.Context, payload *goapackage.BulkPayload) (*goapackage.BulkResult, error) {
	if payload.Size == 0 {
		return nil, goapackage.MakeNotValid(errors.New("size is zero"))
	}
	input := BulkWorkflowInput{
		Operation: BulkWorkflowOperation(payload.Operation),
		Status:    NewStatus(payload.Status),
		Size:      payload.Size,
	}

	opts := temporalsdk_client.StartWorkflowOptions{
		ID:                       BulkWorkflowID,
		WorkflowIDReusePolicy:    temporalapi_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
		TaskQueue:                temporal.GlobalTaskQueue,
		WorkflowExecutionTimeout: time.Hour,
	}
	exec, err := w.tc.ExecuteWorkflow(ctx, opts, BulkWorkflowName, input)
	if err != nil {
		switch err := err.(type) {
		case *temporalapi_serviceerror.NotFound:
			return nil, goapackage.MakeNotAvailable(
				fmt.Errorf("error starting bulk - operation is already in progress (workflowID=%s)", BulkWorkflowID),
			)
		default:
			w.logger.Info("error starting bulk", "err", err)
			return nil, fmt.Errorf("error starting bulk")
		}
	}

	return &goapackage.BulkResult{
		WorkflowID: exec.GetID(),
		RunID:      exec.GetRunID(),
	}, nil
}

func (w *goaWrapper) BulkStatus(ctx context.Context) (*goapackage.BulkStatusResult, error) {
	result := &goapackage.BulkStatusResult{}

	resp, err := w.tc.DescribeWorkflowExecution(ctx, BulkWorkflowID, "")
	if err != nil {
		switch err := err.(type) {
		case *temporalapi_serviceerror.NotFound:
			// We've never seen a workflow run before.
			return result, nil
		default:
			w.logger.Info("error retrieving workflow", "err", err)
			return nil, ErrBulkStatusUnavailable
		}
	}

	if resp.WorkflowExecutionInfo == nil {
		w.logger.Info("error retrieving workflow execution details")
		return nil, ErrBulkStatusUnavailable
	}

	result.WorkflowID = &resp.WorkflowExecutionInfo.Execution.WorkflowId
	result.RunID = &resp.WorkflowExecutionInfo.Execution.RunId

	if resp.WorkflowExecutionInfo.StartTime != nil {
		t := resp.WorkflowExecutionInfo.StartTime.Format(time.RFC3339)
		result.StartedAt = &t
	}

	if resp.WorkflowExecutionInfo.CloseTime != nil {
		t := resp.WorkflowExecutionInfo.CloseTime.Format(time.RFC3339)
		result.ClosedAt = &t
	}

	// Workflow is not running!
	if resp.WorkflowExecutionInfo.Status != temporalapi_enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		st := strings.ToLower(resp.WorkflowExecutionInfo.Status.String())
		result.Status = &st

		return result, nil
	}

	result.Running = true

	// We can use the status property to communicate progress from heartbeats.
	length := len(resp.PendingActivities)
	if length > 0 {
		latest := resp.PendingActivities[length-1]
		progress := &BulkProgress{}
		details := latest.HeartbeatDetails.String()
		if err := json.Unmarshal([]byte(details), progress); err == nil {
			status := fmt.Sprintf("Processing package %d (done: %d)", progress.CurrentID, progress.Count)
			result.Status = &status
		}
	}

	return result, nil
}

func (w *goaWrapper) Confirm(ctx context.Context, payload *goapackage.ConfirmPayload) error {
	goapkg, err := w.Show(ctx, &goapackage.ShowPayload{ID: payload.ID})
	if err != nil {
		return err
	}

	signal := ReviewPerformedSignal{
		Accepted: true,
		Location: &payload.Location,
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
	err = w.tc.SignalWorkflow(context.Background(), *goapkg.WorkflowID, "", ReviewPerformedSignalName, signal)
	if err != nil {
		return goapackage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

func (w *goaWrapper) Move(ctx context.Context, payload *goapackage.MovePayload) error {
	if payload.Location == "" {
		return goapackage.MakeNotValid(errors.New("location attribute is empty"))
	}

	goapkg, err := w.Show(ctx, &goapackage.ShowPayload{ID: payload.ID})
	if err != nil {
		return err
	}

	_, err = InitMoveWorkflow(ctx, w.tc, &MoveWorkflowRequest{
		ID:       payload.ID,
		AIPID:    *goapkg.AipID,
		Location: payload.Location,
	})
	if err != nil {
		w.logger.Error(err, "error initializing move workflow")
		return goapackage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

func (w *goaWrapper) MoveStatus(ctx context.Context, payload *goapackage.MoveStatusPayload) (*goapackage.MoveStatusResult, error) {
	goapkg, err := w.Show(ctx, &goapackage.ShowPayload{ID: payload.ID})
	if err != nil {
		return nil, err
	}

	resp, err := w.tc.DescribeWorkflowExecution(ctx, fmt.Sprintf("%s-%s", MoveWorkflowName, *goapkg.AipID), "")
	if err != nil || resp.WorkflowExecutionInfo == nil {
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
