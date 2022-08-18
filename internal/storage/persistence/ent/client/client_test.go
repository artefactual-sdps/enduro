package client_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/ref"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/client"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/enttest"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/pkg"
	"github.com/artefactual-sdps/enduro/internal/storage/purpose"
	"github.com/artefactual-sdps/enduro/internal/storage/source"
	"github.com/artefactual-sdps/enduro/internal/storage/status"
)

func setUpClient(t *testing.T) (*db.Client, *client.Client) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_fk=1", t.Name())
	entc := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { entc.Close() })

	c := client.NewClient(entc)

	return entc, c
}

func TestCreatePackage(t *testing.T) {
	t.Parallel()

	entc, c := setUpClient(t)

	pkg, err := c.CreatePackage(
		context.Background(),
		"test_package",
		uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e"),
		uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a"),
	)
	assert.NilError(t, err)

	dbpkg := entc.Pkg.GetX(context.Background(), int(pkg.ID))
	assert.Equal(t, dbpkg.Name, "test_package")
	assert.Equal(t, dbpkg.AipID.String(), "488c64cc-d89b-4916-9131-c94152dfb12e")
	assert.Equal(t, dbpkg.ObjectKey.String(), "e2630293-a714-4787-ab6d-e68254a6fb6a")
}

func TestListPackages(t *testing.T) {
	t.Parallel()

	entc, c := setUpClient(t)

	entc.Pkg.Create().
		SetName("Package").
		SetAipID(uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e")).
		SetObjectKey(uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a")).
		SetStatus(status.StatusStored).
		SaveX(context.Background())
	entc.Pkg.Create().
		SetName("Another Package").
		SetAipID(uuid.MustParse("96e182a0-31ab-4738-a620-1ff1954d9ecb")).
		SetObjectKey(uuid.MustParse("49b0a604-6c81-458c-852a-1afa713f1fd9")).
		SetStatus(status.StatusRejected).
		SaveX(context.Background())

	pkgs, err := c.ListPackages(context.Background())
	assert.NilError(t, err)
	assert.DeepEqual(t, pkgs, []*storage.StoredStoragePackage{
		{
			ID:        1,
			Name:      "Package",
			AipID:     "488c64cc-d89b-4916-9131-c94152dfb12e",
			Status:    "stored",
			ObjectKey: "e2630293-a714-4787-ab6d-e68254a6fb6a",
			Location:  nil,
		},
		{
			ID:        2,
			Name:      "Another Package",
			AipID:     "96e182a0-31ab-4738-a620-1ff1954d9ecb",
			Status:    "rejected",
			ObjectKey: "49b0a604-6c81-458c-852a-1afa713f1fd9",
			Location:  nil,
		},
	})
}

func TestReadPackage(t *testing.T) {
	t.Parallel()

	entc, c := setUpClient(t)

	entc.Pkg.Create().
		SetName("Package").
		SetAipID(uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e")).
		SetObjectKey(uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a")).
		SetStatus(status.StatusStored).
		SaveX(context.Background())

	pkg, err := c.ReadPackage(context.Background(), uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e"))
	assert.NilError(t, err)
	assert.DeepEqual(t, pkg, &storage.StoredStoragePackage{
		ID:        1,
		Name:      "Package",
		AipID:     "488c64cc-d89b-4916-9131-c94152dfb12e",
		Status:    "stored",
		ObjectKey: "e2630293-a714-4787-ab6d-e68254a6fb6a",
		Location:  nil,
	})
}

func TestUpdatePackageStatus(t *testing.T) {
	t.Parallel()

	entc, c := setUpClient(t)

	p := entc.Pkg.Create().
		SetName("Package").
		SetAipID(uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e")).
		SetObjectKey(uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a")).
		SetStatus(status.StatusStored).
		SaveX(context.Background())

	err := c.UpdatePackageStatus(context.Background(), status.StatusRejected, p.AipID)
	assert.NilError(t, err)

	entc.Pkg.Query().
		Where(
			pkg.ID(p.ID),
			pkg.StatusEQ(status.StatusRejected),
		).OnlyX(context.Background())
}

func TestUpdatePackageLocation(t *testing.T) {
	t.Parallel()

	entc, c := setUpClient(t)

	p := entc.Pkg.Create().
		SetName("Package").
		SetAipID(uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e")).
		SetObjectKey(uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a")).
		SetStatus(status.StatusStored).
		SetLocation("perma-aips-1").
		SaveX(context.Background())

	err := c.UpdatePackageLocation(context.Background(), "perma-aips-2", p.AipID)
	assert.NilError(t, err)

	entc.Pkg.Query().
		Where(
			pkg.ID(p.ID),
			pkg.Location("perma-aips-2"),
		).OnlyX(context.Background())
}

func TestCreateLocation(t *testing.T) {
	t.Parallel()

	entc, c := setUpClient(t)

	l, err := c.CreateLocation(
		context.Background(),
		"test_location",
		ref.New("location description"),
		source.LocationSourceMinIO,
		purpose.LocationPurposeAIPStore,
		uuid.MustParse("7a090f2c-7bd4-471c-8aa1-8c72125decd5"),
	)
	assert.NilError(t, err)

	dblocation := entc.Location.GetX(context.Background(), int(l.ID))
	assert.Equal(t, dblocation.Name, "test_location")
	assert.Equal(t, dblocation.Description, "location description")
	assert.Equal(t, dblocation.Source, source.LocationSourceMinIO)
	assert.Equal(t, dblocation.Purpose, purpose.LocationPurposeAIPStore)
	assert.Equal(t, dblocation.UUID.String(), "7a090f2c-7bd4-471c-8aa1-8c72125decd5")
}

func TestReadLocation(t *testing.T) {
	t.Parallel()

	entc, c := setUpClient(t)

	entc.Location.Create().
		SetName("test_location").
		SetDescription("location description").
		SetSource(source.LocationSourceMinIO).
		SetPurpose(purpose.LocationPurposeAIPStore).
		SetUUID(uuid.MustParse("7a090f2c-7bd4-471c-8aa1-8c72125decd5")).
		SaveX(context.Background())

	l, err := c.ReadLocation(context.Background(), uuid.MustParse("7a090f2c-7bd4-471c-8aa1-8c72125decd5"))
	assert.NilError(t, err)
	assert.DeepEqual(t, l, &storage.StoredLocation{
		ID:          1,
		Name:        "test_location",
		Description: ref.New("location description"),
		Source:      source.LocationSourceMinIO.String(),
		Purpose:     purpose.LocationPurposeAIPStore.String(),
		UUID:        ref.New("7a090f2c-7bd4-471c-8aa1-8c72125decd5"),
	})
}
