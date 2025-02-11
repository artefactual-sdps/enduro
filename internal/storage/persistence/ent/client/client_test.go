package client_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"entgo.io/ent"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"go.artefactual.dev/tools/ref"
	"gotest.tools/v3/assert"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/client"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/aip"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/enttest"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/hook"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/location"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

var (
	aipID      = uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e")
	locationID = uuid.MustParse("a06a155c-9cf0-4416-a2b6-e90e58ef3186")
	objectKey  = uuid.MustParse("e2630293-a714-4787-ab6d-e68254a6fb6a")
)

func fakeNow() time.Time {
	const longForm = "Jan 2, 2006 at 3:04pm (MST)"
	t, _ := time.Parse(longForm, "Feb 3, 2013 at 7:54pm (UTC)")
	return t
}

func setUpClient(t *testing.T) (*db.Client, *client.Client) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_fk=1", t.Name())
	entc := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { entc.Close() })

	c := client.NewClient(entc)

	// Use ent Hooks to set the create_at fields to a fixed value
	entc.AIP.Use(func(next ent.Mutator) ent.Mutator {
		return hook.AIPFunc(func(ctx context.Context, m *db.AIPMutation) (ent.Value, error) {
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

func TestCreateAIP(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		params  *goastorage.Package
		want    *goastorage.Package
		wantErr string
	}

	for _, tt := range []test{
		{
			name: "Creates an AIP with minimal data",
			params: &goastorage.Package{
				Name:      "test_aip",
				AipID:     aipID,
				ObjectKey: objectKey,
			},
			want: &goastorage.Package{
				Name:      "test_aip",
				AipID:     aipID,
				ObjectKey: objectKey,
				Status:    "unspecified",
				CreatedAt: fakeNow().Format(time.RFC3339),
			},
		},
		{
			name: "Creates an AIP with all data",
			params: &goastorage.Package{
				Name:       "test_aip",
				AipID:      aipID,
				ObjectKey:  objectKey,
				Status:     "stored",
				LocationID: ref.New(locationID),
			},
			want: &goastorage.Package{
				Name:       "test_aip",
				AipID:      aipID,
				ObjectKey:  objectKey,
				Status:     "stored",
				LocationID: ref.New(locationID),
				CreatedAt:  fakeNow().Format(time.RFC3339),
			},
		},
		{
			name: "Errors if locationID is not found",
			params: &goastorage.Package{
				Name:       "test_aip",
				AipID:      aipID,
				ObjectKey:  objectKey,
				LocationID: ref.New(uuid.MustParse("f1508f95-cab7-447f-b6a2-e01bf7c64558")),
			},
			wantErr: "Storage location not found.",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			entc, c := setUpClient(t)
			_, err := entc.Location.Create().
				SetName("Location 1").
				SetDescription("MinIO AIP store").
				SetSource(types.LocationSourceMinIO).
				SetPurpose(types.LocationPurposeAIPStore).
				SetUUID(locationID).
				SetConfig(types.LocationConfig{
					Value: &types.S3Config{
						Bucket: "perma-aips-1",
						Region: "eu-west-1",
					},
				}).
				Save(ctx)
			if err != nil {
				t.Fatalf("Couldn't create test location: %v", err)
			}

			got, err := c.CreateAIP(ctx, tt.params)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}

func TestListAIPs(t *testing.T) {
	t.Parallel()

	aipID2 := uuid.MustParse("96e182a0-31ab-4738-a620-1ff1954d9ecb")
	objectKey2 := uuid.MustParse("49b0a604-6c81-458c-852a-1afa713f1fd9")

	entc, c := setUpClient(t)

	entc.AIP.Create().
		SetName("AIP").
		SetAipID(aipID).
		SetObjectKey(objectKey).
		SetStatus(types.AIPStatusStored).
		SaveX(context.Background())
	entc.AIP.Create().
		SetName("Another AIP").
		SetAipID(aipID2).
		SetObjectKey(objectKey2).
		SetStatus(types.AIPStatusRejected).
		SaveX(context.Background())

	aips, err := c.ListAIPs(context.Background())
	assert.NilError(t, err)
	assert.DeepEqual(t, aips, goastorage.PackageCollection{
		{
			Name:       "AIP",
			AipID:      aipID,
			Status:     "stored",
			ObjectKey:  objectKey,
			LocationID: nil,
			CreatedAt:  "2013-02-03T19:54:00Z",
		},
		{
			Name:       "Another AIP",
			AipID:      aipID2,
			Status:     "rejected",
			ObjectKey:  objectKey2,
			LocationID: nil,
			CreatedAt:  "2013-02-03T19:54:00Z",
		},
	})
}

func TestReadAIP(t *testing.T) {
	t.Parallel()

	t.Run("Returns valid result", func(t *testing.T) {
		entc, c := setUpClient(t)

		entc.AIP.Create().
			SetName("AIP").
			SetAipID(aipID).
			SetObjectKey(objectKey).
			SetStatus(types.AIPStatusStored).
			SaveX(context.Background())

		aip, err := c.ReadAIP(context.Background(), aipID)
		assert.NilError(t, err)
		assert.DeepEqual(t, aip, &goastorage.Package{
			Name:       "AIP",
			AipID:      aipID,
			Status:     "stored",
			ObjectKey:  objectKey,
			LocationID: nil,
			CreatedAt:  "2013-02-03T19:54:00Z",
		})
	})

	t.Run("Returns error when AIP does not exist", func(t *testing.T) {
		t.Parallel()

		_, c := setUpClient(t)

		l, err := c.ReadAIP(context.Background(), aipID)
		assert.Assert(t, l == nil)
		assert.ErrorContains(t, err, "Storage package not found")
	})
}

func TestUpdateAIPStatus(t *testing.T) {
	t.Parallel()

	entc, c := setUpClient(t)

	a := entc.AIP.Create().
		SetName("AIP").
		SetAipID(aipID).
		SetObjectKey(objectKey).
		SetStatus(types.AIPStatusStored).
		SaveX(context.Background())

	err := c.UpdateAIPStatus(context.Background(), a.AipID, types.AIPStatusRejected)
	assert.NilError(t, err)

	entc.AIP.Query().
		Where(
			aip.ID(a.ID),
			aip.StatusEQ(types.AIPStatusRejected),
		).OnlyX(context.Background())
}

func TestUpdateAIPLocation(t *testing.T) {
	t.Parallel()

	entc, c := setUpClient(t)

	l1 := entc.Location.Create().
		SetName("perma-aips-1").
		SetDescription("").
		SetSource(types.LocationSourceMinIO).
		SetPurpose(types.LocationPurposeAIPStore).
		SetUUID(locationID).
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

	a := entc.AIP.Create().
		SetName("AIP").
		SetAipID(aipID).
		SetObjectKey(objectKey).
		SetStatus(types.AIPStatusStored).
		SetLocation(l1).
		SaveX(context.Background())

	err := c.UpdateAIPLocationID(context.Background(), a.AipID, l2.UUID)
	assert.NilError(t, err)

	entc.AIP.Query().
		Where(
			aip.ID(a.ID),
			aip.LocationID(l2.ID),
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
			UUID:        locationID,
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
	assert.Equal(t, dblocation.UUID, locationID)
	assert.Equal(t, dblocation.CreatedAt, time.Date(2013, time.February, 3, 19, 54, 0, 0, time.UTC))
	assert.DeepEqual(t, dblocation.Config.Value, &types.S3Config{Bucket: "perma-aips-1"})
}

func TestCreateURLLocation(t *testing.T) {
	t.Parallel()

	entc, c := setUpClient(t)
	ctx := context.Background()

	l, err := c.CreateLocation(
		ctx,
		&goastorage.Location{
			Name:        "test_url_location",
			Description: ref.New("location description"),
			Source:      types.LocationSourceMinIO.String(),
			Purpose:     types.LocationPurposeAIPStore.String(),
			UUID:        locationID,
		},
		&types.LocationConfig{
			Value: &types.URLConfig{
				URL: "mem://",
			},
		},
	)
	assert.NilError(t, err)

	dblocation, err := entc.Location.Query().Where(location.UUID(l.UUID)).Only(ctx)
	assert.NilError(t, err)
	assert.Equal(t, dblocation.Name, "test_url_location")
	assert.Equal(t, dblocation.Description, "location description")
	assert.Equal(t, dblocation.Source, types.LocationSourceMinIO)
	assert.Equal(t, dblocation.Purpose, types.LocationPurposeAIPStore)
	assert.Equal(t, dblocation.UUID, locationID)
	assert.Equal(t, dblocation.CreatedAt, time.Date(2013, time.February, 3, 19, 54, 0, 0, time.UTC))
	assert.DeepEqual(t, dblocation.Config.Value, &types.URLConfig{URL: "mem://"})
}

func TestListLocations(t *testing.T) {
	t.Parallel()

	locationIDs := [4]uuid.UUID{
		locationID,
		uuid.MustParse("7ba9a118-a662-4047-8547-64bc752b91c6"),
		uuid.MustParse("f0b91bce-dddc-4e15-b1ae-19007685204b"),
		uuid.MustParse("e0ed8b2a-8ae2-4546-b5d8-f0090919df04"),
	}

	entc, c := setUpClient(t)
	entc.Location.Create().
		SetName("Location").
		SetDescription("location").
		SetSource(types.LocationSourceMinIO).
		SetPurpose(types.LocationPurposeAIPStore).
		SetUUID(locationIDs[0]).
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
		SetUUID(locationIDs[1]).
		SetConfig(types.LocationConfig{
			Value: &types.SFTPConfig{
				Address:   "sftp:22",
				Username:  "user",
				Password:  "secret",
				Directory: "upload",
			},
		}).
		SaveX(context.Background())
	entc.Location.Create().
		SetName("URL Location").
		SetDescription("URL location").
		SetSource(types.LocationSourceMinIO).
		SetPurpose(types.LocationPurposeUnspecified).
		SetUUID(locationIDs[2]).
		SetConfig(types.LocationConfig{
			Value: &types.URLConfig{
				URL: "mem://",
			},
		}).
		SaveX(context.Background())
	entc.Location.Create().
		SetName("AMSS Location").
		SetDescription("AMSS Location").
		SetSource(types.LocationSourceAMSS).
		SetPurpose(types.LocationPurposeAIPStore).
		SetUUID(locationIDs[3]).
		SetConfig(types.LocationConfig{
			Value: &types.AMSSConfig{
				APIKey:   "Secret1",
				URL:      "http://127.0.0.1:62081/",
				Username: "analyst",
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
			UUID:        locationIDs[0],
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
			UUID:        locationIDs[1],
			CreatedAt:   "2013-02-03T19:54:00Z",
			Config: &goastorage.SFTPConfig{
				Address:   "sftp:22",
				Username:  "user",
				Password:  "secret",
				Directory: "upload",
			},
		},
		{
			Name:        "URL Location",
			Description: ref.New("URL location"),
			Source:      "minio",
			Purpose:     "unspecified",
			UUID:        locationIDs[2],
			CreatedAt:   "2013-02-03T19:54:00Z",
			Config: &goastorage.URLConfig{
				URL: "mem://",
			},
		},
		{
			Name:        "AMSS Location",
			Description: ref.New("AMSS Location"),
			Source:      "amss",
			Purpose:     "aip_store",
			UUID:        locationIDs[3],
			CreatedAt:   "2013-02-03T19:54:00Z",
			Config: &goastorage.AMSSConfig{
				APIKey:   "Secret1",
				URL:      "http://127.0.0.1:62081/",
				Username: "analyst",
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
			SetUUID(locationID).
			SetConfig(types.LocationConfig{
				Value: &types.S3Config{
					Bucket: "perma-aips-1",
				},
			}).
			SaveX(context.Background())

		l, err := c.ReadLocation(context.Background(), locationID)
		assert.NilError(t, err)
		assert.DeepEqual(t, l, &goastorage.Location{
			Name:        "test_location",
			Description: ref.New("location description"),
			Source:      types.LocationSourceMinIO.String(),
			Purpose:     types.LocationPurposeAIPStore.String(),
			UUID:        locationID,
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

		l, err := c.ReadLocation(context.Background(), locationID)
		assert.Assert(t, l == nil)
		assert.ErrorContains(t, err, "location not found")
	})
}

func TestLocationAIPs(t *testing.T) {
	t.Parallel()

	t.Run("Returns valid result", func(t *testing.T) {
		t.Parallel()

		entc, c := setUpClient(t)
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

		entc.AIP.Create().
			SetName("AIP").
			SetAipID(aipID).
			SetObjectKey(objectKey).
			SetStatus(types.AIPStatusStored).
			SetLocation(l).
			SaveX(context.Background())

		aips, err := c.LocationAIPs(context.Background(), locationID)
		assert.NilError(t, err)
		assert.DeepEqual(t, aips, goastorage.PackageCollection{
			{
				Name:       "AIP",
				AipID:      aipID,
				Status:     "stored",
				ObjectKey:  objectKey,
				LocationID: ref.New(locationID),
				CreatedAt:  "2013-02-03T19:54:00Z",
			},
		})
	})

	t.Run("Returns empty result", func(t *testing.T) {
		t.Parallel()

		entc, c := setUpClient(t)

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

		aips, err := c.LocationAIPs(context.Background(), locationID)
		assert.NilError(t, err)
		assert.Assert(t, len(aips) == 0)
	})

	t.Run("Returns empty result if location does not exist", func(t *testing.T) {
		t.Parallel()

		_, c := setUpClient(t)

		aips, err := c.LocationAIPs(context.Background(), uuid.Nil)
		assert.NilError(t, err)
		assert.Assert(t, len(aips) == 0)
	})
}
