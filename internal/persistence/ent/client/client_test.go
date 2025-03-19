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

func createSIP(
	entc *db.Client,
	name string,
	status enums.SIPStatus,
) (*db.SIP, error) {
	aipID := uuid.MustParse("30223842-0650-4f79-80bd-7bf43b810656")

	return entc.SIP.Create().
		SetName(name).
		SetAipID(aipID).
		SetStatus(int8(status)). // #nosec G115 -- constrained value.
		Save(context.Background())
}

func createWorkflow(
	entc *db.Client,
	sipID int,
	status enums.WorkflowStatus,
) (*db.Workflow, error) {
	return entc.Workflow.Create().
		SetTemporalID("12345").
		SetType(int8(enums.WorkflowTypeCreateAip)).
		SetStatus(int8(status)). // #nosec G115 -- constrained value.
		SetSipID(sipID).
		Save(context.Background())
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("Returns a working ent DB client", func(t *testing.T) {
		t.Parallel()

		entc, _ := setUpClient(t, logr.Discard())
		s, err := createSIP(
			entc,
			"testing 1-2-3",
			enums.SIPStatusInProgress,
		)
		assert.NilError(t, err)

		assert.Equal(t, s.Name, "testing 1-2-3")
		assert.Equal(t, s.AipID, uuid.MustParse("30223842-0650-4f79-80bd-7bf43b810656"))
		assert.Equal(t, s.Status, int8(enums.SIPStatusInProgress))
	})
}
