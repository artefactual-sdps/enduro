package entclient

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/package_"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
)

const TimeFormat = time.RFC3339

type client struct {
	logger logr.Logger
	ent    *db.Client
}

var _ persistence.Service = (*client)(nil)

func New(logger logr.Logger, ent *db.Client) persistence.Service {
	return &client{logger: logger, ent: ent}
}

func (c *client) CreatePackage(ctx context.Context, pkg *package_.Package) (*package_.Package, error) {
	// Validate required fields.
	if pkg.Name == "" {
		return nil, newRequiredFieldError("Name")
	}
	if pkg.WorkflowID == "" {
		return nil, newRequiredFieldError("WorkflowID")
	}

	if pkg.RunID == "" {
		return nil, newRequiredFieldError("RunID")
	}
	runID, err := uuid.Parse(pkg.RunID)
	if err != nil {
		return nil, newParseError(err, "RunID")
	}

	if pkg.AIPID == "" {
		return nil, newRequiredFieldError("AIPID")
	}
	aipID, err := uuid.Parse(pkg.AIPID)
	if err != nil {
		return nil, newParseError(err, "AIPID")
	}

	q := c.ent.Pkg.Create().
		SetName(pkg.Name).
		SetWorkflowID(pkg.WorkflowID).
		SetRunID(runID).
		SetAipID(aipID).
		SetStatus(int8(pkg.Status))

	// Add optional fields.
	if pkg.LocationID.Valid {
		q.SetLocationID(pkg.LocationID.UUID)
	}
	if pkg.StartedAt.Valid {
		q.SetStartedAt(pkg.StartedAt.Time)
	}
	if pkg.CompletedAt.Valid {
		q.SetCompletedAt(pkg.CompletedAt.Time)
	}

	// Set CreatedAt and Save package.
	p, err := q.SetCreatedAt(time.Now()).Save(ctx)
	if err != nil {
		return nil, newDBErrorWithDetails(err, "create package")
	}

	return convertPkgToPackage(p), nil
}
