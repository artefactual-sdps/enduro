package client_test

import (
	"fmt"
	"testing"
	"time"

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

var (
	sipUUID = uuid.New()
	wUUID   = uuid.New()
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
	t *testing.T,
	entc *db.Client,
	name string,
	status enums.SIPStatus,
) (*db.SIP, error) {
	aipID := uuid.MustParse("30223842-0650-4f79-80bd-7bf43b810656")

	return entc.SIP.Create().
		SetUUID(sipUUID).
		SetName(name).
		SetAipID(aipID).
		SetStatus(status).
		Save(t.Context())
}

func createUser(
	t *testing.T,
	entc *db.Client,
	uuid uuid.UUID,
) (*db.User, error) {
	return entc.User.Create().
		SetUUID(uuid).
		SetEmail("nobody@example.com").
		SetName("Test User").
		SetCreatedAt(time.Now()).
		SetOidcIss("https://example.com/oidc").
		SetOidcSub("1234567890").
		Save(t.Context())
}

func createWorkflow(
	t *testing.T,
	entc *db.Client,
	sipID int,
	status enums.WorkflowStatus,
) (*db.Workflow, error) {
	return entc.Workflow.Create().
		SetUUID(wUUID).
		SetTemporalID("12345").
		SetType(enums.WorkflowTypeCreateAip).
		SetStatus(int8(status)). // #nosec G115 -- constrained value.
		SetSipID(sipID).
		Save(t.Context())
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("Returns a working ent DB client", func(t *testing.T) {
		t.Parallel()

		entc, _ := setUpClient(t, logr.Discard())
		s, err := createSIP(
			t,
			entc,
			"testing 1-2-3",
			enums.SIPStatusProcessing,
		)
		assert.NilError(t, err)

		assert.Equal(t, s.Name, "testing 1-2-3")
		assert.Equal(t, s.AipID, uuid.MustParse("30223842-0650-4f79-80bd-7bf43b810656"))
		assert.Equal(t, s.Status, enums.SIPStatusProcessing)
	})
}
