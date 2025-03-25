package client

import (
	"context"
	"time"

	"go.artefactual.dev/tools/ref"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func locationAsGoa(loc *db.Location) *goastorage.Location {
	l := &goastorage.Location{
		Name:        loc.Name,
		Description: &loc.Description,
		Source:      loc.Source.String(),
		Purpose:     loc.Purpose.String(),
		UUID:        loc.UUID,
		CreatedAt:   loc.CreatedAt.Format(time.RFC3339),
	}

	switch c := loc.Config.Value.(type) {
	case *types.S3Config:
		l.Config = &goastorage.S3Config{
			Bucket:    c.Bucket,
			Region:    c.Region,
			Endpoint:  &c.Endpoint,
			PathStyle: &c.PathStyle,
			Profile:   &c.Profile,
			Key:       &c.Key,
			Secret:    &c.Secret,
			Token:     &c.Token,
		}
	case *types.SFTPConfig:
		l.Config = &goastorage.SFTPConfig{
			Address:   c.Address,
			Username:  c.Username,
			Password:  c.Password,
			Directory: c.Directory,
		}
	case *types.AMSSConfig:
		l.Config = &goastorage.AMSSConfig{
			APIKey:   c.APIKey,
			URL:      c.URL,
			Username: c.Username,
		}
	case *types.URLConfig:
		l.Config = &goastorage.URLConfig{
			URL: c.URL,
		}

	}

	return l
}

func aipAsGoa(ctx context.Context, a *db.AIP) *goastorage.AIP {
	p := &goastorage.AIP{
		Name:      a.Name,
		UUID:      a.AipID,
		Status:    a.Status.String(),
		ObjectKey: a.ObjectKey,
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
	}

	// TODO: should we use UUID as the foreign key?
	l, err := a.QueryLocation().Only(ctx)
	if err == nil {
		p.LocationID = &l.UUID
	}

	return p
}

func workflowAsGoa(dbw *db.Workflow) *goastorage.AIPWorkflow {
	w := &goastorage.AIPWorkflow{
		UUID:       dbw.UUID,
		TemporalID: dbw.TemporalID,
		Type:       dbw.Type.String(),
		Status:     dbw.Status.String(),
	}

	if !dbw.StartedAt.IsZero() {
		w.StartedAt = ref.New(dbw.StartedAt.Format(time.RFC3339))
	}
	if !dbw.CompletedAt.IsZero() {
		w.CompletedAt = ref.New(dbw.CompletedAt.Format(time.RFC3339))
	}

	if len(dbw.Edges.Tasks) > 0 {
		for _, dbt := range dbw.Edges.Tasks {
			if dbt != nil {
				w.Tasks = append(w.Tasks, taskAsGoa(dbt))
			}
		}
	}

	return w
}

func taskAsGoa(dbt *db.Task) *goastorage.AIPTask {
	t := &goastorage.AIPTask{
		UUID:   dbt.UUID,
		Name:   dbt.Name,
		Status: dbt.Status.String(),
	}

	if !dbt.StartedAt.IsZero() {
		t.StartedAt = ref.New(dbt.StartedAt.Format(time.RFC3339))
	}
	if !dbt.CompletedAt.IsZero() {
		t.CompletedAt = ref.New(dbt.CompletedAt.Format(time.RFC3339))
	}
	if dbt.Note != "" {
		t.Note = &dbt.Note
	}

	return t
}
