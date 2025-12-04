package client_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"entgo.io/ent"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"go.artefactual.dev/tools/mockutil"
	"go.artefactual.dev/tools/ref"
	"gotest.tools/v3/assert"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/entfilter"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
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

	return entc, c
}

func setUpClientWithHooks(t *testing.T) (*db.Client, *client.Client) {
	t.Helper()

	entc, c := setUpClient(t)

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

	deletionReportKey := fmt.Sprintf("reports/aip_deletion_report_%s", aipID)

	type test struct {
		name    string
		params  *goastorage.AIP
		want    *goastorage.AIP
		wantErr string
	}

	for _, tt := range []test{
		{
			name: "Creates an AIP with minimal data",
			params: &goastorage.AIP{
				Name:      "test_aip",
				UUID:      aipID,
				ObjectKey: objectKey,
			},
			want: &goastorage.AIP{
				Name:      "test_aip",
				UUID:      aipID,
				ObjectKey: objectKey,
				Status:    "unspecified",
				CreatedAt: fakeNow().Format(time.RFC3339),
			},
		},
		{
			name: "Creates an AIP with all data",
			params: &goastorage.AIP{
				Name:              "test_aip",
				UUID:              aipID,
				ObjectKey:         objectKey,
				Status:            "stored",
				LocationUUID:      ref.New(locationID),
				DeletionReportKey: &deletionReportKey,
			},
			want: &goastorage.AIP{
				Name:              "test_aip",
				UUID:              aipID,
				ObjectKey:         objectKey,
				Status:            "stored",
				LocationUUID:      ref.New(locationID),
				CreatedAt:         fakeNow().Format(time.RFC3339),
				DeletionReportKey: &deletionReportKey,
			},
		},
		{
			name: "Errors if locationID is not found",
			params: &goastorage.AIP{
				Name:         "test_aip",
				UUID:         aipID,
				ObjectKey:    objectKey,
				LocationUUID: ref.New(uuid.MustParse("f1508f95-cab7-447f-b6a2-e01bf7c64558")),
			},
			wantErr: "Storage location not found.",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			entc, c := setUpClientWithHooks(t)
			_, err := entc.Location.Create().
				SetName("Location 1").
				SetDescription("MinIO AIP store").
				SetSource(enums.LocationSourceMinio).
				SetPurpose(enums.LocationPurposeAipStore).
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

	aipID := uuid.MustParse("488c64cc-d89b-4916-9131-c94152dfb12e")
	aipID2 := uuid.MustParse("7ba9a118-a662-4047-8547-64bc752b91c6")
	objectKey := aipID
	objectKey2 := aipID2

	tests := []struct {
		name    string
		data    func(t *testing.T, ctx context.Context, entc *db.Client)
		payload *goastorage.ListAipsPayload
		want    *goastorage.AIPs
		wantErr string
	}{
		{
			name: "Returns all AIPs",
			data: func(t *testing.T, ctx context.Context, entc *db.Client) {
				entc.AIP.Create().
					SetName("Test AIP 1").
					SetAipID(aipID).
					SetObjectKey(objectKey).
					SetStatus(enums.AIPStatusStored).
					SetCreatedAt(time.Date(2025, 5, 8, 10, 53, 12, 0, time.UTC)).
					ExecX(ctx)

				entc.AIP.Create().
					SetName("Test AIP 2").
					SetAipID(aipID2).
					SetObjectKey(objectKey2).
					SetStatus(enums.AIPStatusStored).
					SetCreatedAt(time.Date(2025, 5, 8, 10, 53, 48, 0, time.UTC)).
					ExecX(ctx)
			},
			want: &goastorage.AIPs{
				Items: []*goastorage.AIP{
					{
						Name:      "Test AIP 2",
						UUID:      aipID2,
						ObjectKey: objectKey2,
						Status:    "stored",
						CreatedAt: "2025-05-08T10:53:48Z",
					},
					{
						Name:      "Test AIP 1",
						UUID:      aipID,
						ObjectKey: objectKey,
						Status:    "stored",
						CreatedAt: "2025-05-08T10:53:12Z",
					},
				},
				Page: &goastorage.EnduroPage{
					Limit:  entfilter.DefaultPageSize,
					Offset: 0,
					Total:  2,
				},
			},
		},
		{
			name: "Returns paginated AIPs (first page)",
			data: func(t *testing.T, ctx context.Context, entc *db.Client) {
				entc.AIP.Create().
					SetName("Test AIP 1").
					SetAipID(aipID).
					SetObjectKey(objectKey).
					SetStatus(enums.AIPStatusStored).
					SetCreatedAt(time.Date(2025, 5, 8, 10, 53, 12, 0, time.UTC)).
					ExecX(ctx)

				entc.AIP.Create().
					SetName("Test AIP 2").
					SetAipID(aipID2).
					SetObjectKey(objectKey2).
					SetStatus(enums.AIPStatusStored).
					SetCreatedAt(time.Date(2025, 5, 8, 10, 53, 48, 0, time.UTC)).
					ExecX(ctx)
			},
			payload: &goastorage.ListAipsPayload{
				Limit: ref.New(1),
			},
			want: &goastorage.AIPs{
				Items: []*goastorage.AIP{
					{
						Name:      "Test AIP 2",
						UUID:      aipID2,
						ObjectKey: objectKey2,
						Status:    "stored",
						CreatedAt: "2025-05-08T10:53:48Z",
					},
				},
				Page: &goastorage.EnduroPage{
					Limit:  1,
					Offset: 0,
					Total:  2,
				},
			},
		},
		{
			name: "Returns paginated AIPs (second page)",
			data: func(t *testing.T, ctx context.Context, entc *db.Client) {
				entc.AIP.Create().
					SetName("Test AIP 1").
					SetAipID(aipID).
					SetObjectKey(objectKey).
					SetStatus(enums.AIPStatusStored).
					SetCreatedAt(time.Date(2025, 5, 8, 10, 53, 12, 0, time.UTC)).
					ExecX(ctx)

				entc.AIP.Create().
					SetName("Test AIP 2").
					SetAipID(aipID2).
					SetObjectKey(objectKey2).
					SetStatus(enums.AIPStatusStored).
					SetCreatedAt(time.Date(2025, 5, 8, 10, 53, 48, 0, time.UTC)).
					ExecX(ctx)
			},
			payload: &goastorage.ListAipsPayload{
				Offset: ref.New(1),
			},
			want: &goastorage.AIPs{
				Items: []*goastorage.AIP{
					{
						Name:      "Test AIP 1",
						UUID:      aipID,
						ObjectKey: objectKey,
						Status:    "stored",
						CreatedAt: "2025-05-08T10:53:12Z",
					},
				},
				Page: &goastorage.EnduroPage{
					Limit:  entfilter.DefaultPageSize,
					Offset: 1,
					Total:  2,
				},
			},
		},
		{
			name: "Returns AIPs filtered by status",
			data: func(t *testing.T, ctx context.Context, entc *db.Client) {
				entc.AIP.Create().
					SetName("Test AIP 1").
					SetAipID(aipID).
					SetObjectKey(objectKey).
					SetStatus(enums.AIPStatusStored).
					SetCreatedAt(time.Date(2025, 5, 8, 10, 53, 12, 0, time.UTC)).
					ExecX(ctx)

				entc.AIP.Create().
					SetName("Test AIP 2").
					SetAipID(aipID2).
					SetObjectKey(objectKey2).
					SetStatus(enums.AIPStatusDeleted).
					SetCreatedAt(time.Date(2025, 5, 8, 10, 53, 48, 0, time.UTC)).
					ExecX(ctx)
			},
			payload: &goastorage.ListAipsPayload{
				Status: ref.New("stored"),
			},
			want: &goastorage.AIPs{
				Items: []*goastorage.AIP{
					{
						Name:      "Test AIP 1",
						UUID:      aipID,
						ObjectKey: objectKey,
						Status:    "stored",
						CreatedAt: "2025-05-08T10:53:12Z",
					},
				},
				Page: &goastorage.EnduroPage{
					Limit:  entfilter.DefaultPageSize,
					Offset: 0,
					Total:  1,
				},
			},
		},
		{
			name: "Returns AIPs filtered by date range",
			data: func(t *testing.T, ctx context.Context, entc *db.Client) {
				entc.AIP.Create().
					SetName("Test AIP 1").
					SetAipID(aipID).
					SetObjectKey(objectKey).
					SetStatus(enums.AIPStatusStored).
					SetCreatedAt(time.Date(2025, 5, 6, 7, 8, 9, 0, time.UTC)).
					ExecX(ctx)

				entc.AIP.Create().
					SetName("Test AIP 2").
					SetAipID(aipID2).
					SetObjectKey(objectKey2).
					SetStatus(enums.AIPStatusStored).
					SetCreatedAt(time.Date(2025, 5, 7, 8, 9, 10, 0, time.UTC)).
					ExecX(ctx)
			},
			payload: &goastorage.ListAipsPayload{
				EarliestCreatedTime: ref.New("2025-05-07T00:00:00Z"),
			},
			want: &goastorage.AIPs{
				Items: []*goastorage.AIP{
					{
						Name:      "Test AIP 2",
						UUID:      aipID2,
						ObjectKey: objectKey2,
						Status:    "stored",
						CreatedAt: "2025-05-07T08:09:10Z",
					},
				},
				Page: &goastorage.EnduroPage{
					Limit:  entfilter.DefaultPageSize,
					Offset: 0,
					Total:  1,
				},
			},
		},
		{
			name: "Returns AIPs filtered by name",
			data: func(t *testing.T, ctx context.Context, entc *db.Client) {
				entc.AIP.Create().
					SetName("Test AIP 1").
					SetAipID(aipID).
					SetObjectKey(objectKey).
					SetStatus(enums.AIPStatusStored).
					SetCreatedAt(time.Date(2025, 5, 8, 10, 53, 12, 0, time.UTC)).
					ExecX(ctx)

				entc.AIP.Create().
					SetName("Test AIP 2").
					SetAipID(aipID2).
					SetObjectKey(objectKey2).
					SetStatus(enums.AIPStatusStored).
					SetCreatedAt(time.Date(2025, 5, 8, 10, 53, 48, 0, time.UTC)).
					ExecX(ctx)
			},
			payload: &goastorage.ListAipsPayload{
				Name: ref.New("Test AIP 1"),
			},
			want: &goastorage.AIPs{
				Items: []*goastorage.AIP{
					{
						Name:      "Test AIP 1",
						UUID:      aipID,
						ObjectKey: objectKey,
						Status:    "stored",
						CreatedAt: "2025-05-08T10:53:12Z",
					},
				},
				Page: &goastorage.EnduroPage{
					Limit:  entfilter.DefaultPageSize,
					Offset: 0,
					Total:  1,
				},
			},
		},
		{
			name: "Invalid status filter",
			payload: &goastorage.ListAipsPayload{
				Status: ref.New("invalid_status"),
			},
			wantErr: "status: invalid value",
		},
		{
			name: "Invalid date range filter",
			payload: &goastorage.ListAipsPayload{
				EarliestCreatedTime: ref.New("invalid_date"),
			},
			wantErr: "created at: time range: cannot parse start time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			entc, c := setUpClient(t)
			ctx := context.Background()

			if tt.data != nil {
				tt.data(t, ctx, entc)
			}

			aips, err := c.ListAIPs(ctx, tt.payload)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, aips, tt.want)
		})
	}
}

func TestReadAIP(t *testing.T) {
	t.Parallel()

	t.Run("Returns valid result", func(t *testing.T) {
		reportKey := fmt.Sprintf("reports/aip_deletion_report_%s", aipID)
		entc, c := setUpClientWithHooks(t)

		entc.AIP.Create().
			SetName("AIP").
			SetAipID(aipID).
			SetObjectKey(objectKey).
			SetStatus(enums.AIPStatusStored).
			SetDeletionReportKey(reportKey).
			SaveX(context.Background())

		aip, err := c.ReadAIP(context.Background(), aipID)
		assert.NilError(t, err)
		assert.DeepEqual(t, aip, &goastorage.AIP{
			Name:              "AIP",
			UUID:              aipID,
			Status:            "stored",
			ObjectKey:         objectKey,
			LocationUUID:      nil,
			CreatedAt:         "2013-02-03T19:54:00Z",
			DeletionReportKey: &reportKey,
		})
	})

	t.Run("Returns error when AIP does not exist", func(t *testing.T) {
		t.Parallel()

		_, c := setUpClientWithHooks(t)

		l, err := c.ReadAIP(context.Background(), aipID)
		assert.Assert(t, l == nil)
		assert.ErrorContains(t, err, "AIP not found")
	})
}

func TestUpdateAIP(t *testing.T) {
	t.Parallel()

	locID1 := uuid.MustParse("a06a155c-9cf0-4416-a2b6-e90e58ef3186")
	locID2 := uuid.MustParse("f1508f95-cab7-447f-b6a2-e01bf7c64558")
	locIDInvalid := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	type test struct {
		name         string
		aipID        uuid.UUID
		initLocation bool
		updater      persistence.AIPUpdater
		want         *types.AIP
		wantErr      string
	}
	for _, tc := range []test{
		{
			name:  "Updates an AIP",
			aipID: aipID,
			updater: func(aip *types.AIP) (*types.AIP, error) {
				aip.Status = enums.AIPStatusDeleted
				aip.LocationUUID = &locID1
				aip.DeletionReportKey = ref.New("reports/deletion_report.pdf")
				return aip, nil
			},
			want: &types.AIP{
				UUID:              aipID,
				Name:              "AIP",
				CreatedAt:         fakeNow(),
				ObjectKey:         objectKey,
				Status:            enums.AIPStatusDeleted,
				LocationUUID:      &locID1,
				DeletionReportKey: ref.New("reports/deletion_report.pdf"),
			},
		},
		{
			name:  "Keeps nil location when only status changes",
			aipID: aipID,
			updater: func(aip *types.AIP) (*types.AIP, error) {
				aip.Status = enums.AIPStatusDeleted
				return aip, nil
			},
			want: &types.AIP{
				UUID:         aipID,
				Name:         "AIP",
				CreatedAt:    fakeNow(),
				ObjectKey:    objectKey,
				Status:       enums.AIPStatusDeleted,
				LocationUUID: nil,
			},
		},
		{
			name:         "Updates an AIP with a previous location",
			aipID:        aipID,
			initLocation: true,
			updater: func(aip *types.AIP) (*types.AIP, error) {
				aip.Status = enums.AIPStatusDeleted
				aip.LocationUUID = &locID2
				aip.DeletionReportKey = ref.New("reports/deletion_report.pdf")
				return aip, nil
			},
			want: &types.AIP{
				UUID:              aipID,
				Name:              "AIP",
				CreatedAt:         fakeNow(),
				ObjectKey:         objectKey,
				Status:            enums.AIPStatusDeleted,
				LocationUUID:      &locID2,
				DeletionReportKey: ref.New("reports/deletion_report.pdf"),
			},
		},
		{
			name:         "Keeps existing location when updater sets it to nil",
			aipID:        aipID,
			initLocation: true,
			updater: func(aip *types.AIP) (*types.AIP, error) {
				aip.LocationUUID = nil
				return aip, nil
			},
			want: &types.AIP{
				UUID:         aipID,
				Name:         "AIP",
				CreatedAt:    fakeNow(),
				ObjectKey:    objectKey,
				Status:       enums.AIPStatusProcessing,
				LocationUUID: &locID1,
			},
		},
		{
			name:         "Ignores unchanged fields",
			aipID:        aipID,
			initLocation: true,
			updater: func(aip *types.AIP) (*types.AIP, error) {
				aip.Status = enums.AIPStatusProcessing
				aip.LocationUUID = &locID1
				return aip, nil
			},
			want: &types.AIP{
				UUID:         aipID,
				Name:         "AIP",
				CreatedAt:    fakeNow(),
				ObjectKey:    objectKey,
				Status:       enums.AIPStatusProcessing,
				LocationUUID: &locID1,
			},
		},
		{
			name:  "Errors if AIP not found",
			aipID: uuid.MustParse("f1508f95-cab7-447f-b6a2-e01bf7c64558"),
			updater: func(aip *types.AIP) (*types.AIP, error) {
				return aip, nil
			},
			wantErr: "load AIP: not found",
		},
		{
			name:  "Errors if location not found",
			aipID: aipID,
			updater: func(aip *types.AIP) (*types.AIP, error) {
				aip.LocationUUID = &locIDInvalid
				return aip, nil
			},
			wantErr: "load location: not found",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			entc, c := setUpClientWithHooks(t)
			ctx := t.Context()

			loc1 := entc.Location.Create().
				SetName("perma-aips-1").
				SetDescription("").
				SetSource(enums.LocationSourceMinio).
				SetPurpose(enums.LocationPurposeAipStore).
				SetConfig(types.LocationConfig{
					Value: &types.S3Config{Bucket: "perma-aips-1"},
				}).
				SetUUID(locID1).
				SaveX(ctx)

			entc.Location.Create().
				SetName("perma-aips-2").
				SetDescription("").
				SetSource(enums.LocationSourceMinio).
				SetPurpose(enums.LocationPurposeAipStore).
				SetConfig(types.LocationConfig{
					Value: &types.S3Config{Bucket: "perma-aips-1"},
				}).
				SetUUID(locID2).
				SaveX(ctx)

			q := entc.AIP.Create().
				SetName("AIP").
				SetAipID(aipID).
				SetObjectKey(objectKey).
				SetStatus(enums.AIPStatusProcessing)
			if tc.initLocation {
				q.SetLocationID(loc1.ID)
			}
			q.SaveX(ctx)

			got, _, err := c.UpdateAIP(context.Background(), tc.aipID, tc.updater)
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, got, tc.want)
		})
	}
}

func TestUpdateAIPStatus(t *testing.T) {
	t.Parallel()

	entc, c := setUpClientWithHooks(t)

	a := entc.AIP.Create().
		SetName("AIP").
		SetAipID(aipID).
		SetObjectKey(objectKey).
		SetStatus(enums.AIPStatusStored).
		SaveX(context.Background())

	err := c.UpdateAIPStatus(context.Background(), a.AipID, enums.AIPStatusProcessing)
	assert.NilError(t, err)

	entc.AIP.Query().
		Where(
			aip.ID(a.ID),
			aip.StatusEQ(enums.AIPStatusProcessing),
		).OnlyX(context.Background())
}

func TestUpdateAIPLocation(t *testing.T) {
	t.Parallel()

	entc, c := setUpClientWithHooks(t)

	l1 := entc.Location.Create().
		SetName("perma-aips-1").
		SetDescription("").
		SetSource(enums.LocationSourceMinio).
		SetPurpose(enums.LocationPurposeAipStore).
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
		SetSource(enums.LocationSourceMinio).
		SetPurpose(enums.LocationPurposeAipStore).
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
		SetStatus(enums.AIPStatusStored).
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

func TestListWorkflows(t *testing.T) {
	t.Parallel()

	workflowUUID1 := uuid.MustParse("a1b2c3d4-e5f6-7a8b-9c0d-e1f2a3b4c5d6")
	workflowUUID2 := uuid.MustParse("f6e5d4c3-b2a1-0c9b-8a7f-6e5d4c3b2a1f")
	taskUUID1 := uuid.MustParse("1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
	taskUUID2 := uuid.MustParse("6d5c4b3a-2f1e-0d9c-8b7a-6f5e4d3c2b1a")
	startedAt := time.Date(2023, 10, 1, 12, 0, 0, 0, time.UTC)
	completedAt := time.Date(2023, 10, 1, 12, 1, 0, 0, time.UTC)

	addWorkflowData := func(ctx context.Context, entc *db.Client) {
		entc.AIP.Create().
			SetName("AIP").
			SetAipID(aipID).
			SetObjectKey(objectKey).
			SetStatus(enums.AIPStatusStored).
			ExecX(ctx)

		workflow1 := entc.Workflow.Create().
			SetUUID(workflowUUID1).
			SetTemporalID("temporal-id-1").
			SetType(enums.WorkflowTypeMoveAip).
			SetStatus(enums.WorkflowStatusDone).
			SetStartedAt(startedAt).
			SetCompletedAt(completedAt).
			SetAipID(1).
			SaveX(ctx)

		entc.Task.Create().
			SetUUID(taskUUID1).
			SetName("Task 1").
			SetStatus(enums.TaskStatusDone).
			SetStartedAt(startedAt).
			SetCompletedAt(completedAt).
			SetNote("Note 1").
			SetWorkflowID(workflow1.ID).
			SaveX(ctx)

		workflow2 := entc.Workflow.Create().
			SetUUID(workflowUUID2).
			SetTemporalID("temporal-id-2").
			SetType(enums.WorkflowTypeDeleteAip).
			SetStatus(enums.WorkflowStatusInProgress).
			SetStartedAt(startedAt).
			SetAipID(1).
			SaveX(ctx)

		entc.Task.Create().
			SetUUID(taskUUID2).
			SetName("Task 2").
			SetStatus(enums.TaskStatusInProgress).
			SetStartedAt(startedAt).
			SetNote("Note 2").
			SetWorkflowID(workflow2.ID).
			SaveX(ctx)
	}

	type test struct {
		name   string
		filter persistence.WorkflowFilter
		want   goastorage.AIPWorkflowCollection
	}

	for _, tt := range []test{
		{
			name: "Returns all workflows",
			want: goastorage.AIPWorkflowCollection{
				{
					UUID:        workflowUUID1,
					TemporalID:  "temporal-id-1",
					Type:        enums.WorkflowTypeMoveAip.String(),
					Status:      enums.WorkflowStatusDone.String(),
					StartedAt:   ref.New(startedAt.Format(time.RFC3339)),
					CompletedAt: ref.New(completedAt.Format(time.RFC3339)),
					AipUUID:     aipID,
					Tasks: goastorage.AIPTaskCollection{
						{
							UUID:         taskUUID1,
							Name:         "Task 1",
							Status:       enums.TaskStatusDone.String(),
							StartedAt:    ref.New(startedAt.Format(time.RFC3339)),
							CompletedAt:  ref.New(completedAt.Format(time.RFC3339)),
							Note:         ref.New("Note 1"),
							WorkflowUUID: workflowUUID1,
						},
					},
				},
				{
					UUID:       workflowUUID2,
					TemporalID: "temporal-id-2",
					Type:       enums.WorkflowTypeDeleteAip.String(),
					Status:     enums.WorkflowStatusInProgress.String(),
					StartedAt:  ref.New(startedAt.Format(time.RFC3339)),
					AipUUID:    aipID,
					Tasks: goastorage.AIPTaskCollection{
						{
							UUID:         taskUUID2,
							Name:         "Task 2",
							Status:       enums.TaskStatusInProgress.String(),
							StartedAt:    ref.New(startedAt.Format(time.RFC3339)),
							Note:         ref.New("Note 2"),
							WorkflowUUID: workflowUUID2,
						},
					},
				},
			},
		},
		{
			name:   "Returns workflows matching AIP UUID",
			filter: persistence.WorkflowFilter{AIPUUID: &aipID},
			want: goastorage.AIPWorkflowCollection{
				{
					UUID:        workflowUUID1,
					TemporalID:  "temporal-id-1",
					Type:        enums.WorkflowTypeMoveAip.String(),
					Status:      enums.WorkflowStatusDone.String(),
					StartedAt:   ref.New(startedAt.Format(time.RFC3339)),
					CompletedAt: ref.New(completedAt.Format(time.RFC3339)),
					AipUUID:     aipID,
					Tasks: goastorage.AIPTaskCollection{
						{
							UUID:         taskUUID1,
							Name:         "Task 1",
							Status:       enums.TaskStatusDone.String(),
							StartedAt:    ref.New(startedAt.Format(time.RFC3339)),
							CompletedAt:  ref.New(completedAt.Format(time.RFC3339)),
							Note:         ref.New("Note 1"),
							WorkflowUUID: workflowUUID1,
						},
					},
				},
				{
					UUID:       workflowUUID2,
					TemporalID: "temporal-id-2",
					Type:       enums.WorkflowTypeDeleteAip.String(),
					Status:     enums.WorkflowStatusInProgress.String(),
					StartedAt:  ref.New(startedAt.Format(time.RFC3339)),
					AipUUID:    aipID,
					Tasks: goastorage.AIPTaskCollection{
						{
							UUID:         taskUUID2,
							Name:         "Task 2",
							Status:       enums.TaskStatusInProgress.String(),
							StartedAt:    ref.New(startedAt.Format(time.RFC3339)),
							Note:         ref.New("Note 2"),
							WorkflowUUID: workflowUUID2,
						},
					},
				},
			},
		},
		{
			name:   "Returns workflows matching status",
			filter: persistence.WorkflowFilter{Status: ref.New(enums.WorkflowStatusInProgress)},
			want: goastorage.AIPWorkflowCollection{
				{
					UUID:       workflowUUID2,
					TemporalID: "temporal-id-2",
					Type:       enums.WorkflowTypeDeleteAip.String(),
					Status:     enums.WorkflowStatusInProgress.String(),
					StartedAt:  ref.New(startedAt.Format(time.RFC3339)),
					AipUUID:    aipID,
					Tasks: goastorage.AIPTaskCollection{
						{
							UUID:         taskUUID2,
							Name:         "Task 2",
							Status:       enums.TaskStatusInProgress.String(),
							StartedAt:    ref.New(startedAt.Format(time.RFC3339)),
							Note:         ref.New("Note 2"),
							WorkflowUUID: workflowUUID2,
						},
					},
				},
			},
		},
		{
			name:   "Returns workflows matching type",
			filter: persistence.WorkflowFilter{Type: ref.New(enums.WorkflowTypeDeleteAip)},
			want: goastorage.AIPWorkflowCollection{
				{
					UUID:       workflowUUID2,
					TemporalID: "temporal-id-2",
					Type:       enums.WorkflowTypeDeleteAip.String(),
					Status:     enums.WorkflowStatusInProgress.String(),
					StartedAt:  ref.New(startedAt.Format(time.RFC3339)),
					AipUUID:    aipID,
					Tasks: goastorage.AIPTaskCollection{
						{
							UUID:         taskUUID2,
							Name:         "Task 2",
							Status:       enums.TaskStatusInProgress.String(),
							StartedAt:    ref.New(startedAt.Format(time.RFC3339)),
							Note:         ref.New("Note 2"),
							WorkflowUUID: workflowUUID2,
						},
					},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			entc, c := setUpClient(t)

			addWorkflowData(ctx, entc)

			got, err := c.ListWorkflows(ctx, &tt.filter)
			assert.NilError(t, err)
			assert.DeepEqual(
				t,
				got,
				tt.want,
				cmpopts.SortSlices(func(a, b goastorage.AIPWorkflow) bool {
					return a.UUID.String() < b.UUID.String()
				}),
				mockutil.EquateNearlySameTime(),
			)
		})
	}
}

func TestCreateLocation(t *testing.T) {
	t.Parallel()

	entc, c := setUpClientWithHooks(t)
	ctx := context.Background()

	l, err := c.CreateLocation(
		ctx,
		&goastorage.Location{
			Name:        "test_location",
			Description: ref.New("location description"),
			Source:      enums.LocationSourceMinio.String(),
			Purpose:     enums.LocationPurposeAipStore.String(),
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
	assert.Equal(t, dblocation.Source, enums.LocationSourceMinio)
	assert.Equal(t, dblocation.Purpose, enums.LocationPurposeAipStore)
	assert.Equal(t, dblocation.UUID, locationID)
	assert.Equal(t, dblocation.CreatedAt, time.Date(2013, time.February, 3, 19, 54, 0, 0, time.UTC))
	assert.DeepEqual(t, dblocation.Config.Value, &types.S3Config{Bucket: "perma-aips-1"})
}

func TestCreateURLLocation(t *testing.T) {
	t.Parallel()

	entc, c := setUpClientWithHooks(t)
	ctx := context.Background()

	l, err := c.CreateLocation(
		ctx,
		&goastorage.Location{
			Name:        "test_url_location",
			Description: ref.New("location description"),
			Source:      enums.LocationSourceMinio.String(),
			Purpose:     enums.LocationPurposeAipStore.String(),
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
	assert.Equal(t, dblocation.Source, enums.LocationSourceMinio)
	assert.Equal(t, dblocation.Purpose, enums.LocationPurposeAipStore)
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

	entc, c := setUpClientWithHooks(t)
	entc.Location.Create().
		SetName("Location").
		SetDescription("location").
		SetSource(enums.LocationSourceMinio).
		SetPurpose(enums.LocationPurposeAipStore).
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
		SetSource(enums.LocationSourceMinio).
		SetPurpose(enums.LocationPurposeAipStore).
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
		SetSource(enums.LocationSourceMinio).
		SetPurpose(enums.LocationPurposeUnspecified).
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
		SetSource(enums.LocationSourceAmss).
		SetPurpose(enums.LocationPurposeAipStore).
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

		entc, c := setUpClientWithHooks(t)

		entc.Location.Create().
			SetName("test_location").
			SetDescription("location description").
			SetSource(enums.LocationSourceMinio).
			SetPurpose(enums.LocationPurposeAipStore).
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
			Source:      enums.LocationSourceMinio.String(),
			Purpose:     enums.LocationPurposeAipStore.String(),
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

		_, c := setUpClientWithHooks(t)

		l, err := c.ReadLocation(context.Background(), locationID)
		assert.Assert(t, l == nil)
		assert.ErrorContains(t, err, "location not found")
	})
}

func TestLocationAIPs(t *testing.T) {
	t.Parallel()

	t.Run("Returns valid result", func(t *testing.T) {
		t.Parallel()

		entc, c := setUpClientWithHooks(t)
		l := entc.Location.Create().
			SetName("Location").
			SetDescription("location").
			SetSource(enums.LocationSourceMinio).
			SetPurpose(enums.LocationPurposeAipStore).
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
			SetStatus(enums.AIPStatusStored).
			SetLocation(l).
			SaveX(context.Background())

		aips, err := c.LocationAIPs(context.Background(), locationID)
		assert.NilError(t, err)
		assert.DeepEqual(t, aips, goastorage.AIPCollection{
			{
				Name:         "AIP",
				UUID:         aipID,
				Status:       "stored",
				ObjectKey:    objectKey,
				LocationUUID: ref.New(locationID),
				CreatedAt:    "2013-02-03T19:54:00Z",
			},
		})
	})

	t.Run("Returns empty result", func(t *testing.T) {
		t.Parallel()

		entc, c := setUpClientWithHooks(t)

		entc.Location.Create().
			SetName("Location").
			SetDescription("location").
			SetSource(enums.LocationSourceMinio).
			SetPurpose(enums.LocationPurposeAipStore).
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

		_, c := setUpClientWithHooks(t)

		aips, err := c.LocationAIPs(context.Background(), uuid.Nil)
		assert.NilError(t, err)
		assert.Assert(t, len(aips) == 0)
	})
}
