package package_

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

type (
	PreservationTaskStatus uint
)

const (
	StatusUnspecified PreservationTaskStatus = iota
	StatusComplete
	StatusProcessing
	StatusFailed
)

func NewPreservationTaskStatus(status string) PreservationTaskStatus {
	var s PreservationTaskStatus

	switch strings.ToLower(status) {
	case "processing":
		s = StatusProcessing
	case "complete":
		s = StatusComplete
	case "failed":
		s = StatusFailed
	default:
		s = StatusUnspecified
	}

	return s
}

func (p PreservationTaskStatus) String() string {
	switch p {
	case StatusProcessing:
		return "processing"
	case StatusComplete:
		return "complete"
	case StatusFailed:
		return "failed"
	}
	return "unspecified"
}

func (p PreservationTaskStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *PreservationTaskStatus) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*p = NewPreservationTaskStatus(s)

	return nil
}

// PreservationAction represents a preservation action in the preservation_action table.
type PreservationAction struct {
	ID          uint         `db:"id"`
	Name        string       `db:"name"`
	WorkflowID  string       `db:"workflow_id"`
	StartedAt   sql.NullTime `db:"started_at"`
	CompletedAt sql.NullTime `db:"completed_at"`
	PackageID   uint         `db:"package_id"`
}

// PreservationTask represents a preservation action task in the preservation_task table.
type PreservationTask struct {
	ID                   uint                   `db:"id"`
	TaskID               string                 `db:"task_id"`
	Name                 string                 `db:"name"`
	Status               PreservationTaskStatus `db:"status"`
	StartedAt            sql.NullTime           `db:"started_at"`
	CompletedAt          sql.NullTime           `db:"completed_at"`
	PreservationActionID uint                   `db:"preservation_action_id"`
}

func (svc *packageImpl) CreatePreservationAction(ctx context.Context, pa *PreservationAction) error {
	query := `INSERT INTO preservation_action (name, workflow_id, started_at, completed_at, package_id) VALUES ((?), (?), (?), (?), (?))`
	args := []interface{}{
		pa.Name,
		pa.WorkflowID,
		pa.StartedAt,
		pa.CompletedAt,
		pa.PackageID,
	}

	query = svc.db.Rebind(query)
	res, err := svc.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error inserting preservation action: %w", err)
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		return fmt.Errorf("error retrieving insert ID: %w", err)
	}

	pa.ID = uint(id)

	// publishEvent(ctx, svc.events, EventTypePackageUpdated, pa.PackageID)

	return nil
}

func (svc *packageImpl) CreatePreservationTask(ctx context.Context, pt *PreservationTask) error {
	query := `INSERT INTO preservation_task (task_id, name, status, started_at, completed_at, preservation_action_id) VALUES ((?), (?), (?), (?), (?), (?))`
	args := []interface{}{
		pt.TaskID,
		pt.Name,
		pt.Status,
		pt.StartedAt,
		pt.CompletedAt,
		pt.PreservationActionID,
	}

	query = svc.db.Rebind(query)
	res, err := svc.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error inserting preservation task: %w", err)
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		return fmt.Errorf("error retrieving insert ID: %w", err)
	}

	pt.ID = uint(id)

	return nil
}
