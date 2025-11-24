package client_test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/entfilter"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/timerange"
)

func TestCreateBatch(t *testing.T) {
	t.Parallel()

	batchUUID := uuid.New()
	started := sql.NullTime{Time: time.Now(), Valid: true}
	completed := sql.NullTime{Time: started.Time.Add(time.Second), Valid: true}
	uploaderID := uuid.New()
	uploaderID2 := uuid.New()

	tests := []struct {
		name    string
		batch   *datatypes.Batch
		want    *datatypes.Batch
		wantErr string
	}{
		{
			name: "Creates a new Batch in the DB",
			batch: &datatypes.Batch{
				UUID:        batchUUID,
				Identifier:  "Test Batch 1",
				Status:      enums.BatchStatusProcessing,
				SIPSCount:   5,
				StartedAt:   started,
				CompletedAt: completed,
			},
			want: &datatypes.Batch{
				ID:          1,
				UUID:        batchUUID,
				Identifier:  "Test Batch 1",
				Status:      enums.BatchStatusProcessing,
				SIPSCount:   5,
				CreatedAt:   time.Now(),
				StartedAt:   started,
				CompletedAt: completed,
			},
		},
		{
			name: "Creates a Batch with missing optional fields",
			batch: &datatypes.Batch{
				UUID:       batchUUID,
				Identifier: "Test Batch 2",
				Status:     enums.BatchStatusProcessing,
				SIPSCount:  5,
			},
			want: &datatypes.Batch{
				ID:         1,
				UUID:       batchUUID,
				Identifier: "Test Batch 2",
				Status:     enums.BatchStatusProcessing,
				SIPSCount:  5,
				CreatedAt:  time.Now(),
			},
		},
		{
			name: "Creates a Batch with an uploader (existing user)",
			batch: &datatypes.Batch{
				UUID:       batchUUID,
				Identifier: "Test Batch 3",
				Status:     enums.BatchStatusProcessing,
				SIPSCount:  5,
				Uploader: &datatypes.User{
					UUID:    uploaderID,
					Email:   "nobody@example.com",
					Name:    "Test User",
					OIDCIss: "https://example.com/oidc",
					OIDCSub: "1234567890",
				},
			},
			want: &datatypes.Batch{
				ID:         1,
				UUID:       batchUUID,
				Identifier: "Test Batch 3",
				Status:     enums.BatchStatusProcessing,
				SIPSCount:  5,
				CreatedAt:  time.Now(),
				Uploader: &datatypes.User{
					UUID:      uploaderID,
					Email:     "nobody@example.com",
					Name:      "Test User",
					CreatedAt: time.Now(),
					OIDCIss:   "https://example.com/oidc",
					OIDCSub:   "1234567890",
				},
			},
		},
		{
			name: "Creates a Batch with an uploader (new user)",
			batch: &datatypes.Batch{
				UUID:       batchUUID,
				Identifier: "Test Batch 4",
				Status:     enums.BatchStatusProcessing,
				SIPSCount:  5,
				Uploader: &datatypes.User{
					UUID:      uploaderID2,
					Email:     "nobody2@example.com",
					Name:      "Test User 2",
					CreatedAt: time.Now(),
					OIDCIss:   "https://example.com/oidc",
					OIDCSub:   "newuser",
				},
			},
			want: &datatypes.Batch{
				ID:         1,
				UUID:       batchUUID,
				Identifier: "Test Batch 4",
				Status:     enums.BatchStatusProcessing,
				SIPSCount:  5,
				CreatedAt:  time.Now(),
				Uploader: &datatypes.User{
					UUID:      uploaderID2,
					Email:     "nobody2@example.com",
					Name:      "Test User 2",
					CreatedAt: time.Now(),
					OIDCIss:   "https://example.com/oidc",
					OIDCSub:   "newuser",
				},
			},
		},
		{
			name:    "Required field error for missing UUID",
			batch:   &datatypes.Batch{},
			wantErr: "invalid data error: field \"UUID\" is required",
		},
		{
			name:    "Required field error for missing Identifier",
			batch:   &datatypes.Batch{UUID: batchUUID},
			wantErr: "invalid data error: field \"Identifier\" is required",
		},
		{
			name: "Required field error for missing SIPSCount",
			batch: &datatypes.Batch{
				UUID:       batchUUID,
				Identifier: "Test Batch",
			},
			wantErr: "invalid data error: field \"SIPSCount\" is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, svc := setUpClient(t, logr.Discard())
			ctx := t.Context()
			batch := *tt.batch // Make a local copy of batch.

			_, err := createUser(t, c, uploaderID)
			assert.NilError(t, err)

			err = svc.CreateBatch(ctx, &batch)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			assert.DeepEqual(t, &batch, tt.want,
				cmpopts.EquateApproxTime(time.Second),
				cmpopts.IgnoreUnexported(db.Batch{}, db.BatchEdges{}),
			)
		})
	}
}

func TestUpdateBatch(t *testing.T) {
	t.Parallel()

	batchUUID := uuid.New()
	batchUUID2 := uuid.New()
	uploaderID := uuid.New()

	started := sql.NullTime{Time: time.Now(), Valid: true}
	started2 := sql.NullTime{
		Time: func() time.Time {
			t, _ := time.Parse(time.RFC3339, "1980-01-01T09:30:00Z")
			return t
		}(),
		Valid: true,
	}

	completed := sql.NullTime{Time: started.Time.Add(time.Second), Valid: true}
	completed2 := sql.NullTime{Time: started2.Time.Add(time.Second), Valid: true}

	type params struct {
		batchUUID uuid.UUID
		batch     *datatypes.Batch
		updater   persistence.BatchUpdater
	}
	tests := []struct {
		name    string
		args    params
		want    *datatypes.Batch
		wantErr string
	}{
		{
			name: "Updates all Batch columns",
			args: params{
				batchUUID: batchUUID,
				batch: &datatypes.Batch{
					UUID:        batchUUID,
					Identifier:  "Test Batch",
					Status:      enums.BatchStatusProcessing,
					SIPSCount:   5,
					StartedAt:   started,
					CompletedAt: completed,
					Uploader: &datatypes.User{
						UUID:    uploaderID,
						Email:   "nobody@example.com",
						Name:    "Test User",
						OIDCIss: "https://example.com/oidc",
						OIDCSub: "1234567890",
					},
				},
				updater: func(s *datatypes.Batch) (*datatypes.Batch, error) {
					s.ID = 100          // No-op, can't update ID.
					s.UUID = batchUUID2 // No-op, can't update UUID.
					s.Identifier = "Updated Batch"
					s.Status = enums.BatchStatusIngested
					s.SIPSCount = 10
					s.CreatedAt = started2.Time // No-op, can't update CreatedAt.
					s.StartedAt = started2
					s.CompletedAt = completed2
					s.Uploader = &datatypes.User{UUID: uuid.New()} // No-op, can't update Uploader.
					return s, nil
				},
			},
			want: &datatypes.Batch{
				ID:          1,
				UUID:        batchUUID,
				Identifier:  "Updated Batch",
				Status:      enums.BatchStatusIngested,
				SIPSCount:   10,
				CreatedAt:   time.Now(),
				StartedAt:   started2,
				CompletedAt: completed2,
				Uploader: &datatypes.User{
					UUID:      uploaderID,
					Email:     "nobody@example.com",
					Name:      "Test User",
					CreatedAt: time.Now(),
					OIDCIss:   "https://example.com/oidc",
					OIDCSub:   "1234567890",
				},
			},
		},
		{
			name: "Only updates selected columns",
			args: params{
				batchUUID: batchUUID,
				batch: &datatypes.Batch{
					UUID:       batchUUID,
					Identifier: "Test Batch",
					Status:     enums.BatchStatusProcessing,
					SIPSCount:  5,
					StartedAt:  started,
				},
				updater: func(s *datatypes.Batch) (*datatypes.Batch, error) {
					// Immutable.
					s.ID = 1234

					// Mutable.
					s.Status = enums.BatchStatusIngested
					s.SIPSCount = 10
					s.CompletedAt = completed

					return s, nil
				},
			},
			want: &datatypes.Batch{
				ID:          1,
				UUID:        batchUUID,
				Identifier:  "Test Batch",
				Status:      enums.BatchStatusIngested,
				SIPSCount:   10,
				CreatedAt:   time.Now(),
				StartedAt:   started,
				CompletedAt: completed,
			},
		},
		{
			name: "Ignores invalid fields",
			args: params{
				batchUUID: batchUUID,
				batch: &datatypes.Batch{
					UUID:       batchUUID,
					Identifier: "Test Batch",
					Status:     enums.BatchStatusQueued,
					SIPSCount:  5,
				},
				updater: func(s *datatypes.Batch) (*datatypes.Batch, error) {
					// Invalid.
					s.Status = ""
					s.SIPSCount = 0

					// Valid.
					s.StartedAt = started

					return s, nil
				},
			},
			want: &datatypes.Batch{
				ID:         1,
				UUID:       batchUUID,
				Identifier: "Test Batch",
				Status:     enums.BatchStatusQueued,
				SIPSCount:  5,
				CreatedAt:  time.Now(),
				StartedAt:  started,
			},
		},
		{
			name:    "Errors when Batch to update is not found",
			args:    params{batchUUID: batchUUID},
			wantErr: "not found error: db: batch not found",
		},
		{
			name: "Errors when the updater errors",
			args: params{
				batchUUID: batchUUID,
				batch: &datatypes.Batch{
					UUID:       batchUUID,
					Identifier: "Test Batch",
					Status:     enums.BatchStatusProcessing,
					SIPSCount:  5,
				},
				updater: func(s *datatypes.Batch) (*datatypes.Batch, error) {
					return nil, fmt.Errorf("Bad input")
				},
			},
			wantErr: "invalid data error: updater error: Bad input",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, svc := setUpClient(t, logr.Discard())
			ctx := t.Context()

			if tt.args.batch != nil {
				batch := *tt.args.batch // Make a local copy of batch.

				_, err := createUser(t, c, uploaderID)
				assert.NilError(t, err)

				err = svc.CreateBatch(ctx, &batch)
				assert.NilError(t, err)
			}

			batch, err := svc.UpdateBatch(ctx, tt.args.batchUUID, tt.args.updater)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			assert.DeepEqual(t, batch, tt.want,
				cmpopts.EquateApproxTime(time.Second),
				cmpopts.IgnoreUnexported(db.Batch{}, db.BatchEdges{}),
			)
		})
	}
}

func TestDeleteBatch(t *testing.T) {
	t.Parallel()

	batchUUID := uuid.New()

	for _, tc := range []struct {
		name    string
		id      uuid.UUID
		wantErr string
	}{
		{
			name: "Deletes a Batch",
		},
		{
			name:    "Fails to delete a missing Batch",
			id:      uuid.New(),
			wantErr: "not found error: db: batch not found: delete Batch",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, svc := setUpClient(t, logr.Discard())
			ctx := t.Context()

			batch := &datatypes.Batch{
				UUID:       batchUUID,
				Identifier: "Test Batch",
				Status:     enums.BatchStatusQueued,
				SIPSCount:  5,
			}

			err := svc.CreateBatch(ctx, batch)
			assert.NilError(t, err)

			if tc.id == uuid.Nil {
				tc.id = batch.UUID
			}

			err = svc.DeleteBatch(ctx, tc.id)
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}
			assert.NilError(t, err)

			_, err = svc.ReadBatch(ctx, tc.id)
			assert.Error(t, err, "not found error: db: batch not found")
		})
	}
}

func TestReadBatch(t *testing.T) {
	t.Parallel()

	uploaderID := uuid.New()
	batchUUID := uuid.New()

	for _, tt := range []struct {
		name      string
		batchUUID uuid.UUID
		want      *datatypes.Batch
		wantErr   string
	}{
		{
			name:      "Reads a Batch",
			batchUUID: batchUUID,
			want: &datatypes.Batch{
				ID:          1,
				UUID:        batchUUID,
				Identifier:  "Test Batch",
				Status:      enums.BatchStatusCanceled,
				SIPSCount:   5,
				CreatedAt:   time.Now(),
				StartedAt:   sql.NullTime{Time: time.Now().Add(time.Second), Valid: true},
				CompletedAt: sql.NullTime{Time: time.Now().Add(time.Minute), Valid: true},
				Uploader: &datatypes.User{
					UUID:      uploaderID,
					Email:     "nobody@example.com",
					Name:      "Test User",
					CreatedAt: time.Now(),
					OIDCIss:   "https://example.com/oidc",
					OIDCSub:   "1234567890",
				},
			},
		},
		{
			name:      "Fails to read a missing Batch",
			batchUUID: batchUUID,
			wantErr:   "not found error: db: batch not found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			entc, svc := setUpClient(t, logr.Discard())
			ctx := t.Context()

			if tt.want != nil {
				user, err := createUser(t, entc, uploaderID)
				assert.NilError(t, err)

				_, err = entc.Batch.Create().
					SetUUID(tt.want.UUID).
					SetIdentifier(tt.want.Identifier).
					SetStatus(tt.want.Status).
					SetSipsCount(tt.want.SIPSCount).
					SetCreatedAt(tt.want.CreatedAt).
					SetStartedAt(tt.want.StartedAt.Time).
					SetCompletedAt(tt.want.CompletedAt.Time).
					SetUploader(user).
					Save(ctx)
				assert.NilError(t, err)
			}

			s, err := svc.ReadBatch(ctx, tt.batchUUID)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, s, tt.want, cmpopts.EquateApproxTime(time.Second))
		})
	}
}

func TestListBatches(t *testing.T) {
	t.Parallel()

	batchUUID := uuid.New()
	batchUUID2 := uuid.New()
	batchUUID3 := uuid.New()
	uploaderID := uuid.New()
	uploaderID2 := uuid.New()

	started := sql.NullTime{
		Time: func() time.Time {
			t, _ := time.Parse(time.RFC3339, "2024-09-25T09:31:11Z")
			return t
		}(),
		Valid: true,
	}
	started2 := sql.NullTime{
		Time: func() time.Time {
			t, _ := time.Parse(time.RFC3339, "2024-09-25T10:03:42Z")
			return t
		}(),
		Valid: true,
	}

	completed := sql.NullTime{Time: started.Time.Add(time.Second), Valid: true}
	completed2 := sql.NullTime{Time: started2.Time.Add(time.Second), Valid: true}

	type results struct {
		data []*datatypes.Batch
		page *persistence.Page
	}
	tests := []struct {
		name        string
		data        []*datatypes.Batch
		batchFilter *persistence.BatchFilter
		want        results
		wantErr     string
	}{
		{
			name: "Returns all Batches",
			data: []*datatypes.Batch{
				{
					UUID:        batchUUID,
					Identifier:  "Test Batch 1",
					Status:      enums.BatchStatusIngested,
					SIPSCount:   5,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					UUID:        batchUUID2,
					Identifier:  "Test Batch 2",
					Status:      enums.BatchStatusProcessing,
					SIPSCount:   5,
					StartedAt:   started2,
					CompletedAt: completed2,
					Uploader: &datatypes.User{
						UUID:      uploaderID,
						Email:     "nobody@example.com",
						Name:      "Test User",
						CreatedAt: time.Now(),
						OIDCIss:   "https://example.com/oidc",
						OIDCSub:   "1234567890",
					},
				},
				{
					UUID:        batchUUID3,
					Identifier:  "Test Batch 3",
					Status:      enums.BatchStatusCanceled,
					SIPSCount:   5,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			want: results{
				data: []*datatypes.Batch{
					{
						ID:          1,
						UUID:        batchUUID,
						Identifier:  "Test Batch 1",
						Status:      enums.BatchStatusIngested,
						SIPSCount:   5,
						CreatedAt:   time.Now(),
						StartedAt:   started,
						CompletedAt: completed,
					},
					{
						ID:          2,
						UUID:        batchUUID2,
						Identifier:  "Test Batch 2",
						Status:      enums.BatchStatusProcessing,
						SIPSCount:   5,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
						Uploader: &datatypes.User{
							UUID:      uploaderID,
							Email:     "nobody@example.com",
							Name:      "Test User",
							CreatedAt: time.Now(),
							OIDCIss:   "https://example.com/oidc",
							OIDCSub:   "1234567890",
						},
					},
					{
						ID:          3,
						UUID:        batchUUID3,
						Identifier:  "Test Batch 3",
						Status:      enums.BatchStatusCanceled,
						SIPSCount:   5,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
					},
				},
				page: &persistence.Page{
					Limit: entfilter.DefaultPageSize,
					Total: 3,
				},
			},
		},
		{
			name: "Returns first page of Batches",
			data: []*datatypes.Batch{
				{
					UUID:        batchUUID,
					Identifier:  "Test Batch 1",
					Status:      enums.BatchStatusIngested,
					SIPSCount:   5,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					UUID:        batchUUID2,
					Identifier:  "Test Batch 2",
					Status:      enums.BatchStatusProcessing,
					SIPSCount:   5,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			batchFilter: &persistence.BatchFilter{
				Page: persistence.Page{Limit: 1},
			},
			want: results{
				data: []*datatypes.Batch{
					{
						ID:          1,
						UUID:        batchUUID,
						Identifier:  "Test Batch 1",
						Status:      enums.BatchStatusIngested,
						SIPSCount:   5,
						CreatedAt:   time.Now(),
						StartedAt:   started,
						CompletedAt: completed,
					},
				},
				page: &persistence.Page{
					Limit: 1,
					Total: 2,
				},
			},
		},
		{
			name: "Returns second page of Batches",
			data: []*datatypes.Batch{
				{
					UUID:        batchUUID,
					Identifier:  "Test Batch 1",
					Status:      enums.BatchStatusIngested,
					SIPSCount:   5,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					UUID:        batchUUID2,
					Identifier:  "Test Batch 2",
					Status:      enums.BatchStatusProcessing,
					SIPSCount:   5,
					StartedAt:   started2,
					CompletedAt: completed2,
					Uploader: &datatypes.User{
						UUID:      uploaderID,
						Email:     "nobody@example.com",
						Name:      "Test User",
						CreatedAt: time.Now(),
						OIDCIss:   "https://example.com/oidc",
						OIDCSub:   "1234567890",
					},
				},
			},
			batchFilter: &persistence.BatchFilter{
				Page: persistence.Page{Limit: 1, Offset: 1},
			},
			want: results{
				data: []*datatypes.Batch{
					{
						ID:          2,
						UUID:        batchUUID2,
						Identifier:  "Test Batch 2",
						Status:      enums.BatchStatusProcessing,
						SIPSCount:   5,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
						Uploader: &datatypes.User{
							UUID:      uploaderID,
							Email:     "nobody@example.com",
							Name:      "Test User",
							CreatedAt: time.Now(),
							OIDCIss:   "https://example.com/oidc",
							OIDCSub:   "1234567890",
						},
					},
				},
				page: &persistence.Page{
					Limit:  1,
					Offset: 1,
					Total:  2,
				},
			},
		},
		{
			name: "Returns Batches whose names contain a string",
			data: []*datatypes.Batch{
				{
					UUID:        batchUUID,
					Identifier:  "Test Batch",
					Status:      enums.BatchStatusIngested,
					SIPSCount:   5,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					UUID:        batchUUID2,
					Identifier:  "small.zip",
					Status:      enums.BatchStatusProcessing,
					SIPSCount:   5,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			batchFilter: &persistence.BatchFilter{
				Identifier: ref.New("small"),
			},
			want: results{
				data: []*datatypes.Batch{
					{
						ID:          2,
						UUID:        batchUUID2,
						Identifier:  "small.zip",
						Status:      enums.BatchStatusProcessing,
						SIPSCount:   5,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
					},
				},
				page: &persistence.Page{
					Limit: entfilter.DefaultPageSize,
					Total: 1,
				},
			},
		},
		{
			name: "Returns Batches filtered by status",
			data: []*datatypes.Batch{
				{
					UUID:        batchUUID,
					Identifier:  "Test Batch 1",
					Status:      enums.BatchStatusIngested,
					SIPSCount:   5,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					UUID:        batchUUID2,
					Identifier:  "Test Batch 2",
					Status:      enums.BatchStatusProcessing,
					SIPSCount:   5,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			batchFilter: &persistence.BatchFilter{
				Status: ref.New(enums.BatchStatusProcessing),
			},
			want: results{
				data: []*datatypes.Batch{
					{
						ID:          2,
						UUID:        batchUUID2,
						Identifier:  "Test Batch 2",
						Status:      enums.BatchStatusProcessing,
						SIPSCount:   5,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
					},
				},
				page: &persistence.Page{
					Limit: entfilter.DefaultPageSize,
					Total: 1,
				},
			},
		},
		{
			name: "Returns Batches filtered by CreatedAt",
			data: []*datatypes.Batch{
				{
					UUID:        batchUUID,
					Identifier:  "Test Batch 1",
					Status:      enums.BatchStatusIngested,
					SIPSCount:   5,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					UUID:        batchUUID2,
					Identifier:  "Test Batch 2",
					Status:      enums.BatchStatusProcessing,
					SIPSCount:   5,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			batchFilter: &persistence.BatchFilter{
				CreatedAt: func(t *testing.T) *timerange.Range {
					r, err := timerange.New(
						time.Now().Add(-1*time.Minute),
						time.Now().Add(time.Minute),
					)
					if err != nil {
						t.Fatalf("Error: %v", err)
					}
					return &r
				}(t),
			},
			want: results{
				data: []*datatypes.Batch{
					{
						ID:          1,
						UUID:        batchUUID,
						Identifier:  "Test Batch 1",
						Status:      enums.BatchStatusIngested,
						SIPSCount:   5,
						CreatedAt:   time.Now(),
						StartedAt:   started,
						CompletedAt: completed,
					},
					{
						ID:          2,
						UUID:        batchUUID2,
						Identifier:  "Test Batch 2",
						Status:      enums.BatchStatusProcessing,
						SIPSCount:   5,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
					},
				},
				page: &persistence.Page{
					Limit: entfilter.DefaultPageSize,
					Total: 2,
				},
			},
		},
		{
			name: "Returns no results when no Batches match CreatedAt range",
			data: []*datatypes.Batch{
				{
					UUID:        batchUUID,
					Identifier:  "Test Batch 1",
					Status:      enums.BatchStatusIngested,
					SIPSCount:   5,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					UUID:        batchUUID2,
					Identifier:  "Test Batch 2",
					Status:      enums.BatchStatusProcessing,
					SIPSCount:   5,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			batchFilter: &persistence.BatchFilter{
				CreatedAt: func(t *testing.T) *timerange.Range {
					r, err := timerange.New(
						time.Now().Add(time.Minute),
						time.Now().Add(2*time.Minute),
					)
					if err != nil {
						t.Fatalf("Error: %v", err)
					}
					return &r
				}(t),
			},
			want: results{
				data: []*datatypes.Batch{},
				page: &persistence.Page{
					Limit: entfilter.DefaultPageSize,
					Total: 0,
				},
			},
		},
		{
			name: "Returns Batches filtered by UploaderID",
			data: []*datatypes.Batch{
				{
					UUID:        batchUUID,
					Identifier:  "Test Batch 1",
					Status:      enums.BatchStatusIngested,
					SIPSCount:   5,
					StartedAt:   started,
					CompletedAt: completed,
					Uploader: &datatypes.User{
						UUID:      uploaderID,
						Email:     "nobody@example.com",
						Name:      "Nobody Here",
						CreatedAt: time.Now(),
						OIDCIss:   "https://example.com/oidc",
						OIDCSub:   "1234567890",
					},
				},
				{
					UUID:        batchUUID2,
					Identifier:  "Test Batch 2",
					Status:      enums.BatchStatusProcessing,
					SIPSCount:   5,
					StartedAt:   started2,
					CompletedAt: completed2,
					Uploader: &datatypes.User{
						UUID:      uploaderID2,
						Email:     "test@example.com",
						Name:      "Test Example",
						CreatedAt: time.Now(),
						OIDCIss:   "https://example.com/oidc",
						OIDCSub:   "otheruser",
					},
				},
			},
			batchFilter: &persistence.BatchFilter{
				UploaderID: ref.New(uploaderID2),
			},
			want: results{
				data: []*datatypes.Batch{
					{
						ID:          2,
						UUID:        batchUUID2,
						Identifier:  "Test Batch 2",
						Status:      enums.BatchStatusProcessing,
						SIPSCount:   5,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
						Uploader: &datatypes.User{
							UUID:      uploaderID2,
							Email:     "test@example.com",
							Name:      "Test Example",
							CreatedAt: time.Now(),
							OIDCIss:   "https://example.com/oidc",
							OIDCSub:   "otheruser",
						},
					},
				},
				page: &persistence.Page{
					Limit: entfilter.DefaultPageSize,
					Total: 1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			entc, svc := setUpClient(t, logr.Discard())
			ctx := t.Context()

			if len(tt.data) > 0 {
				for _, batch := range tt.data {
					q := entc.Batch.Create().
						SetUUID(batch.UUID).
						SetIdentifier(batch.Identifier).
						SetStatus(batch.Status).
						SetSipsCount(batch.SIPSCount).
						SetCreatedAt(time.Now()).
						SetStartedAt(batch.StartedAt.Time).
						SetCompletedAt(batch.CompletedAt.Time)

					if batch.Uploader != nil {
						user, err := entc.User.Create().
							SetUUID(batch.Uploader.UUID).
							SetEmail(batch.Uploader.Email).
							SetName(batch.Uploader.Name).
							SetCreatedAt(batch.Uploader.CreatedAt).
							SetOidcIss(batch.Uploader.OIDCIss).
							SetOidcSub(batch.Uploader.OIDCSub).
							Save(t.Context())
						assert.NilError(t, err)
						q.SetUploader(user)
					}

					_, err := q.Save(ctx)
					assert.NilError(t, err)
				}
			}

			got, pg, err := svc.ListBatches(ctx, tt.batchFilter)
			assert.NilError(t, err)

			assert.DeepEqual(t, got, tt.want.data,
				cmpopts.EquateApproxTime(time.Second),
				cmpopts.IgnoreUnexported(db.Batch{}, db.BatchEdges{}),
			)
			assert.DeepEqual(t, pg, tt.want.page)
		})
	}
}
