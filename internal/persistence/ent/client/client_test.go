package entclient_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	entclient "github.com/artefactual-sdps/enduro/internal/persistence/ent/client"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db/enttest"
)

func setUpClient(t *testing.T, logger logr.Logger) (*db.Client, persistence.Service) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_fk=1", t.Name())
	entc := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { entc.Close() })

	c := entclient.New(logger, entc)

	return entc, c
}

func TestNew(t *testing.T) {
	t.Run("Returns a working ent DB client", func(t *testing.T) {
		t.Parallel()

		entc, _ := setUpClient(t, logr.Discard())
		runID := uuid.New()
		aipID := uuid.New()

		p, err := entc.Pkg.Create().
			SetName("testing 1-2-3").
			SetWorkflowID("12345").
			SetRunID(runID).
			SetAipID(aipID).
			SetStatus(int8(enums.NewPackageStatus("in progress"))).
			Save(context.Background())

		assert.NilError(t, err)
		assert.Equal(t, p.Name, "testing 1-2-3")
		assert.Equal(t, p.WorkflowID, "12345")
		assert.Equal(t, p.RunID, runID)
		assert.Equal(t, p.AipID, aipID)
		assert.Equal(t, p.Status, int8(enums.PackageStatusInProgress))
	})
}
