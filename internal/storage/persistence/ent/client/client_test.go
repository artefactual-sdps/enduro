package client_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"entgo.io/ent"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"gotest.tools/v3/assert"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/ref"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/client"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/enttest"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/hook"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/location"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/pkg"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func fakeNow() time.Time {
	const longForm = "Jan 2, 2006 at 3:04pm (MST)"
	t, _ := time.Parse(longForm, "Feb 3, 2013 at 7:54pm (PST)")
	return t
}

func setUpClient(t *testing.T) (*db.Client, *client.Client) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_fk=1", t.Name())
	entc := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { entc.Close() })

	c := client.NewClient(entc)

	// Use ent Hooks to set the create_at fields to a fixed value
	entc.Pkg.Use(func(next ent.Mutator) ent.Mutator {
		return hook.PkgFunc(func(ctx context.Context, m *db.PkgMutation) (ent.Value, error) {
			if m.Op() == db.OpCreate {
				m.SetCreatedAt(fakeNow())
			}
			return next.Mutate(ctx, m)
		})
	})
	entc.Location.Use(func(next ent.Mutator) ent.Mutator {
		return hook.LocationFunc(func(ctx context.Context, m *db.LocationMutation) (ent.Value, error) {
			if m.Op() == db.OpCreate {
				m.SetCreatedAt(fakeNow())
			}
			return next.Mutate(ctx, m)
		})
	})

	return entc, c
}

func TestCreatePackage(t *testing.T) {
	t.Parallel()

	entc, c := setUpClient(t)

	ctx := context.Background()
	p, err := c.CreatePackage(
		ctx,
		&goastorage.Package{
			Name:      "test_package",
			AipID:     uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e"),
			ObjectKey: uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a"),
		},
	)
	assert.NilError(t, err)

	dbpkg, err := entc.Pkg.Query().Where(pkg.AipID(p.AipID)).Only(ctx)
	assert.NilError(t, err)
	assert.Equal(t, dbpkg.Name, "test_package")
	assert.Equal(t, dbpkg.AipID.String(), "488c64cc-d89b-4916-9131-c94152dfb12e")
	assert.Equal(t, dbpkg.ObjectKey.String(), "e2630293-a714-4787-ab6d-e68254a6fb6a")
	assert.Equal(t, dbpkg.CreatedAt, time.Date(2013, time.February, 3, 19, 54, 0, 0, time.UTC))
}

func TestListPackages(t *testing.T) {
	t.Parallel()

	entc, c := setUpClient(t)

	entc.Pkg.Create().
		SetName("Package").
		SetAipID(uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e")).
		SetObjectKey(uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a")).
		SetStatus(types.StatusStored).
		SaveX(context.Background())
	entc.Pkg.Create().
		SetName("Another Package").
		SetAipID(uuid.MustParse("96e182a0-31ab-4738-a620-1ff1954d9ecb")).
		SetObjectKey(uuid.MustParse("49b0a604-6c81-458c-852a-1afa713f1fd9")).
		SetStatus(types.StatusRejected).
		SaveX(context.Background())

	pkgs, err := c.ListPackages(context.Background())
	assert.NilError(t, err)
	assert.DeepEqual(t, pkgs, goastorage.PackageCollection{
		{
			Name:       "Package",
			AipID:      uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e"),
			Status:     "stored",
			ObjectKey:  uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a"),
			LocationID: nil,
			CreatedAt:  "2013-02-03T19:54:00Z",
		},
		{
			Name:       "Another Package",
			AipID:      uuid.MustParse("96e182a0-31ab-4738-a620-1ff1954d9ecb"),
			Status:     "rejected",
			ObjectKey:  uuid.MustParse("49b0a604-6c81-458c-852a-1afa713f1fd9"),
			LocationID: nil,
			CreatedAt:  "2013-02-03T19:54:00Z",
		},
	})
}

func TestReadPackage(t *testing.T) {
	t.Parallel()

	t.Run("Returns valid result", func(t *testing.T) {
		entc, c := setUpClient(t)

		entc.Pkg.Create().
			SetName("Package").
			SetAipID(uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e")).
			SetObjectKey(uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a")).
			SetStatus(types.StatusStored).
			SaveX(context.Background())

		pkg, err := c.ReadPackage(context.Background(), uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e"))
		assert.NilError(t, err)
		assert.DeepEqual(t, pkg, &goastorage.Package{
			Name:       "Package",
			AipID:      uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e"),
			Status:     "stored",
			ObjectKey:  uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a"),
			LocationID: nil,
			CreatedAt:  "2013-02-03T19:54:00Z",
		})
	})

	t.Run("Returns error when package does not exist", func(t *testing.T) {
		t.Parallel()

		_, c := setUpClient(t)

		l, err := c.ReadPackage(context.Background(), uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e"))
		assert.Assert(t, l == nil)
		assert.ErrorContains(t, err, "package not found")
	})
}

func TestUpdatePackageStatus(t *testing.T) {
	t.Parallel()

	entc, c := setUpClient(t)

	p := entc.Pkg.Create().
		SetName("Package").
		SetAipID(uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e")).
		SetObjectKey(uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a")).
		SetStatus(types.StatusStored).
		SaveX(context.Background())

	err := c.UpdatePackageStatus(context.Background(), p.AipID, types.StatusRejected)
	assert.NilError(t, err)

	entc.Pkg.Query().
		Where(
			pkg.ID(p.ID),
			pkg.StatusEQ(types.StatusRejected),
		).OnlyX(context.Background())
}

func TestUpdatePackageLocation(t *testing.T) {
	t.Parallel()

	entc, c := setUpClient(t)

	l1 := entc.Location.Create().
		SetName("perma-aips-1").
		SetDescription("").
		SetSource(types.LocationSourceMinIO).
		SetPurpose(types.LocationPurposeAIPStore).
		SetUUID(uuid.MustParse("af2cd8cb-6f20-41c2-ab64-225d48312ac8")).
		SetConfig(types.LocationConfig{
			Value: &types.S3Config{
				Bucket: "perma-aips-1",
			},
		}).
		SaveX(context.Background())

	l2 := entc.Location.Create().
		SetName("perma-aips-2").
		SetDescription("").
		SetSource(types.LocationSourceMinIO).
		SetPurpose(types.LocationPurposeAIPStore).
		SetUUID(uuid.MustParse("aef501be-b726-4d32-820d-549541d29b64")).
		SetConfig(types.LocationConfig{
			Value: &types.S3Config{
				Bucket: "perma-aips-2",
			},
		}).
		SaveX(context.Background())

	p := entc.Pkg.Create().
		SetName("Package").
		SetAipID(uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e")).
		SetObjectKey(uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a")).
		SetStatus(types.StatusStored).
		SetLocation(l1).
		SaveX(context.Background())

	err := c.UpdatePackageLocationID(context.Background(), p.AipID, l2.UUID)
	assert.NilError(t, err)

	entc.Pkg.Query().
		Where(
			pkg.ID(p.ID),
			pkg.LocationID(l2.ID),
		).OnlyX(context.Background())
}

func TestCreateLocation(t *testing.T) {
	t.Parallel()

	entc, c := setUpClient(t)
	ctx := context.Background()

	l, err := c.CreateLocation(
		ctx,
		&goastorage.Location{
			Name:        "test_location",
			Description: ref.New("location description"),
			Source:      types.LocationSourceMinIO.String(),
			Purpose:     types.LocationPurposeAIPStore.String(),
			UUID:        uuid.MustParse("7a090f2c-7bd4-471c-8aa1-8c72125decd5"),
		},
		&types.LocationConfig{
			Value: &types.S3Config{
				Bucket: "perma-aips-1",
			},
		},
	)
	assert.NilError(t, err)

	dblocation, err := entc.Location.Query().Where(location.UUID(l.UUID)).Only(ctx)
	assert.NilError(t, err)
	assert.Equal(t, dblocation.Name, "test_location")
	assert.Equal(t, dblocation.Description, "location description")
	assert.Equal(t, dblocation.Source, types.LocationSourceMinIO)
	assert.Equal(t, dblocation.Purpose, types.LocationPurposeAIPStore)
	assert.Equal(t, dblocation.UUID.String(), "7a090f2c-7bd4-471c-8aa1-8c72125decd5")
	assert.Equal(t, dblocation.CreatedAt, time.Date(2013, time.February, 3, 19, 54, 0, 0, time.UTC))
	assert.DeepEqual(t, dblocation.Config.Value, &types.S3Config{Bucket: "perma-aips-1"})
}

func TestListLocations(t *testing.T) {
	t.Parallel()

	entc, c := setUpClient(t)

	entc.Location.Create().
		SetName("Location").
		SetDescription("location").
		SetSource(types.LocationSourceMinIO).
		SetPurpose(types.LocationPurposeAIPStore).
		SetUUID(uuid.MustParse("021f7ac2-5b0b-4620-b574-21f6a206cff3")).
		SetConfig(types.LocationConfig{
			Value: &types.S3Config{
				Bucket: "perma-aips-1",
			},
		}).
		SaveX(context.Background())
	entc.Location.Create().
		SetName("Another Location").
		SetDescription("another location").
		SetSource(types.LocationSourceMinIO).
		SetPurpose(types.LocationPurposeAIPStore).
		SetUUID(uuid.MustParse("7ba9a118-a662-4047-8547-64bc752b91c6")).
		SetConfig(types.LocationConfig{
			Value: &types.S3Config{
				Bucket: "perma-aips-2",
			},
		}).
		SaveX(context.Background())

	locations, err := c.ListLocations(context.Background())
	assert.NilError(t, err)
	assert.DeepEqual(t, locations, goastorage.LocationCollection{
		{
			Name:        "Location",
			Description: ref.New("location"),
			Source:      "minio",
			Purpose:     "aip_store",
			UUID:        uuid.MustParse("021f7ac2-5b0b-4620-b574-21f6a206cff3"),
			CreatedAt:   "2013-02-03T19:54:00Z",
			Config: &goastorage.S3Config{
				Bucket:    "perma-aips-1",
				Endpoint:  ref.New(""),
				PathStyle: ref.New(false),
				Profile:   ref.New(""),
				Key:       ref.New(""),
				Secret:    ref.New(""),
				Token:     ref.New(""),
			},
		},
		{
			Name:        "Another Location",
			Description: ref.New("another location"),
			Source:      "minio",
			Purpose:     "aip_store",
			UUID:        uuid.MustParse("7ba9a118-a662-4047-8547-64bc752b91c6"),
			CreatedAt:   "2013-02-03T19:54:00Z",
			Config: &goastorage.S3Config{
				Bucket:    "perma-aips-2",
				Endpoint:  ref.New(""),
				PathStyle: ref.New(false),
				Profile:   ref.New(""),
				Key:       ref.New(""),
				Secret:    ref.New(""),
				Token:     ref.New(""),
			},
		},
	})
}

func TestReadLocation(t *testing.T) {
	t.Parallel()

	t.Run("Returns valid result", func(t *testing.T) {
		t.Parallel()

		entc, c := setUpClient(t)

		entc.Location.Create().
			SetName("test_location").
			SetDescription("location description").
			SetSource(types.LocationSourceMinIO).
			SetPurpose(types.LocationPurposeAIPStore).
			SetUUID(uuid.MustParse("7a090f2c-7bd4-471c-8aa1-8c72125decd5")).
			SetConfig(types.LocationConfig{
				Value: &types.S3Config{
					Bucket: "perma-aips-1",
				},
			}).
			SaveX(context.Background())

		l, err := c.ReadLocation(context.Background(), uuid.MustParse("7a090f2c-7bd4-471c-8aa1-8c72125decd5"))
		assert.NilError(t, err)
		assert.DeepEqual(t, l, &goastorage.Location{
			Name:        "test_location",
			Description: ref.New("location description"),
			Source:      types.LocationSourceMinIO.String(),
			Purpose:     types.LocationPurposeAIPStore.String(),
			UUID:        uuid.MustParse("7a090f2c-7bd4-471c-8aa1-8c72125decd5"),
			CreatedAt:   "2013-02-03T19:54:00Z",
			Config: &goastorage.S3Config{
				Bucket:    "perma-aips-1",
				Endpoint:  ref.New(""),
				PathStyle: ref.New(false),
				Profile:   ref.New(""),
				Key:       ref.New(""),
				Secret:    ref.New(""),
				Token:     ref.New(""),
			},
		})
	})

	t.Run("Returns error when location does not exist", func(t *testing.T) {
		t.Parallel()

		_, c := setUpClient(t)

		l, err := c.ReadLocation(context.Background(), uuid.MustParse("7a090f2c-7bd4-471c-8aa1-8c72125decd5"))
		assert.Assert(t, l == nil)
		assert.ErrorContains(t, err, "location not found")
	})
}

func TestLocationPackages(t *testing.T) {
	t.Parallel()

	t.Run("Returns valid result", func(t *testing.T) {
		t.Parallel()

		entc, c := setUpClient(t)

		locationID := uuid.MustParse("021f7ac2-5b0b-4620-b574-21f6a206cff3")
		l := entc.Location.Create().
			SetName("Location").
			SetDescription("location").
			SetSource(types.LocationSourceMinIO).
			SetPurpose(types.LocationPurposeAIPStore).
			SetUUID(locationID).
			SetConfig(types.LocationConfig{
				Value: &types.S3Config{
					Bucket: "perma-aips-1",
				},
			}).
			SaveX(context.Background())

		entc.Pkg.Create().
			SetName("Package").
			SetAipID(uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e")).
			SetObjectKey(uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a")).
			SetStatus(types.StatusStored).
			SetLocation(l).
			SaveX(context.Background())

		pkgs, err := c.LocationPackages(context.Background(), locationID)
		assert.NilError(t, err)
		assert.DeepEqual(t, pkgs, goastorage.PackageCollection{
			{
				Name:       "Package",
				AipID:      uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e"),
				Status:     "stored",
				ObjectKey:  uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a"),
				LocationID: ref.New(locationID),
				CreatedAt:  "2013-02-03T19:54:00Z",
			},
		})
	})

	t.Run("Returns empty result", func(t *testing.T) {
		t.Parallel()

		entc, c := setUpClient(t)

		locationID := uuid.MustParse("021f7ac2-5b0b-4620-b574-21f6a206cff3")
		entc.Location.Create().
			SetName("Location").
			SetDescription("location").
			SetSource(types.LocationSourceMinIO).
			SetPurpose(types.LocationPurposeAIPStore).
			SetUUID(locationID).
			SetConfig(types.LocationConfig{
				Value: &types.S3Config{
					Bucket: "perma-aips-1",
				},
			}).
			SaveX(context.Background())

		pkgs, err := c.LocationPackages(context.Background(), locationID)
		assert.NilError(t, err)
		assert.Assert(t, len(pkgs) == 0)
	})

	t.Run("Returns empty result if location does not exist", func(t *testing.T) {
		t.Parallel()

		_, c := setUpClient(t)

		pkgs, err := c.LocationPackages(context.Background(), uuid.Nil)
		assert.NilError(t, err)
		assert.Assert(t, len(pkgs) == 0)
	})
}
