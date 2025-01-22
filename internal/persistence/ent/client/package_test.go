package entclient_test

import (
	"context"
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
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	entclient "github.com/artefactual-sdps/enduro/internal/persistence/ent/client"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/timerange"
)

func TestCreatePackage(t *testing.T) {
	t.Parallel()

	runID := uuid.New()
	aipID := uuid.NullUUID{UUID: uuid.New(), Valid: true}
	locID := uuid.NullUUID{UUID: uuid.New(), Valid: true}
	started := sql.NullTime{Time: time.Now(), Valid: true}
	completed := sql.NullTime{Time: started.Time.Add(time.Second), Valid: true}

	type params struct {
		pkg *datatypes.Package
	}
	tests := []struct {
		name    string
		args    params
		want    *datatypes.Package
		wantErr string
	}{
		{
			name: "Saves a new package in the DB",
			args: params{
				pkg: &datatypes.Package{
					Name:        "Test package 1",
					WorkflowID:  "workflow-1",
					RunID:       runID.String(),
					AIPID:       aipID,
					LocationID:  locID,
					Status:      enums.PackageStatusInProgress,
					StartedAt:   started,
					CompletedAt: completed,
				},
			},
			want: &datatypes.Package{
				ID:          1,
				Name:        "Test package 1",
				WorkflowID:  "workflow-1",
				RunID:       runID.String(),
				AIPID:       aipID,
				LocationID:  locID,
				Status:      enums.PackageStatusInProgress,
				CreatedAt:   time.Now(),
				StartedAt:   started,
				CompletedAt: completed,
			},
		},
		{
			name: "Saves a package with missing optional fields",
			args: params{
				pkg: &datatypes.Package{
					Name:       "Test package 2",
					WorkflowID: "workflow-2",
					RunID:      runID.String(),
					Status:     enums.PackageStatusInProgress,
				},
			},
			want: &datatypes.Package{
				ID:         1,
				Name:       "Test package 2",
				WorkflowID: "workflow-2",
				RunID:      runID.String(),
				Status:     enums.PackageStatusInProgress,
				CreatedAt:  time.Now(),
			},
		},
		{
			name: "Required field error for missing Name",
			args: params{
				pkg: &datatypes.Package{},
			},
			wantErr: "invalid data error: field \"Name\" is required",
		},
		{
			name: "Required field error for missing WorkflowID",
			args: params{
				pkg: &datatypes.Package{
					Name: "Missing WorkflowID",
				},
			},
			wantErr: "invalid data error: field \"WorkflowID\" is required",
		},
		{
			name: "Required field error for missing RunID",
			args: params{
				pkg: &datatypes.Package{
					Name:       "Missing RunID",
					WorkflowID: "workflow-12345",
				},
			},
			wantErr: "invalid data error: field \"RunID\" is required",
		},
		{
			name: "Errors on invalid RunID",
			args: params{
				pkg: &datatypes.Package{
					Name:       "Invalid package 1",
					WorkflowID: "workflow-invalid",
					RunID:      "Bad UUID",
				},
			},
			wantErr: "invalid data error: parse error: field \"RunID\": invalid UUID length: 8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, svc := setUpClient(t, logr.Discard())
			ctx := context.Background()
			pkg := *tt.args.pkg // Make a local copy of pkg.

			err := svc.CreatePackage(ctx, &pkg)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			assert.DeepEqual(t, &pkg, tt.want,
				cmpopts.EquateApproxTime(time.Millisecond*100),
				cmpopts.IgnoreUnexported(db.Pkg{}, db.PkgEdges{}),
			)
		})
	}
}

func TestUpdatePackage(t *testing.T) {
	t.Parallel()

	runID := uuid.MustParse("c5f7c35a-d5a6-4e00-b4da-b036ce5b40bc")
	runID2 := uuid.MustParse("c04d0191-d7ce-46dd-beff-92d6830082ff")

	aipID := uuid.NullUUID{
		UUID:  uuid.MustParse("e2ace0da-8697-453d-9ea1-4c9b62309e54"),
		Valid: true,
	}
	aipID2 := uuid.NullUUID{
		UUID:  uuid.MustParse("7d085541-af56-4444-9ce2-d6401ff4c97b"),
		Valid: true,
	}

	locID := uuid.NullUUID{
		UUID:  uuid.MustParse("146182ff-9923-4869-bca1-0bbc0f822025"),
		Valid: true,
	}
	locID2 := uuid.NullUUID{
		UUID:  uuid.MustParse("6e30694b-6497-439f-bf99-83af165e02c3"),
		Valid: true,
	}

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
		pkg     *datatypes.Package
		updater persistence.PackageUpdater
	}
	tests := []struct {
		name    string
		args    params
		want    *datatypes.Package
		wantErr string
	}{
		{
			name: "Updates all package columns",
			args: params{
				pkg: &datatypes.Package{
					Name:        "Test package",
					WorkflowID:  "workflow-1",
					RunID:       runID.String(),
					AIPID:       aipID,
					LocationID:  locID,
					Status:      enums.PackageStatusInProgress,
					StartedAt:   started,
					CompletedAt: completed,
				},
				updater: func(p *datatypes.Package) (*datatypes.Package, error) {
					p.ID = 100 // No-op, can't update ID.
					p.Name = "Updated package"
					p.WorkflowID = "workflow-2"
					p.RunID = runID2.String()
					p.AIPID = aipID2
					p.LocationID = locID2
					p.Status = enums.PackageStatusDone
					p.CreatedAt = started2.Time // No-op, can't update CreatedAt.
					p.StartedAt = started2
					p.CompletedAt = completed2
					return p, nil
				},
			},
			want: &datatypes.Package{
				ID:          1,
				Name:        "Updated package",
				WorkflowID:  "workflow-2",
				RunID:       runID2.String(),
				AIPID:       aipID2,
				LocationID:  locID2,
				Status:      enums.PackageStatusDone,
				CreatedAt:   time.Now(),
				StartedAt:   started2,
				CompletedAt: completed2,
			},
		},
		{
			name: "Only updates selected columns",
			args: params{
				pkg: &datatypes.Package{
					Name:       "Test package",
					WorkflowID: "workflow-1",
					RunID:      runID.String(),
					AIPID:      aipID,
					Status:     enums.PackageStatusInProgress,
					StartedAt:  started,
				},
				updater: func(p *datatypes.Package) (*datatypes.Package, error) {
					p.Status = enums.PackageStatusDone
					p.CompletedAt = completed
					return p, nil
				},
			},
			want: &datatypes.Package{
				ID:          1,
				Name:        "Test package",
				WorkflowID:  "workflow-1",
				RunID:       runID.String(),
				AIPID:       aipID,
				Status:      enums.PackageStatusDone,
				CreatedAt:   time.Now(),
				StartedAt:   started,
				CompletedAt: completed,
			},
		},
		{
			name: "Errors when package to update is not found",
			args: params{
				updater: func(p *datatypes.Package) (*datatypes.Package, error) {
					return nil, fmt.Errorf("Bad input")
				},
			},
			wantErr: "not found error: db: pkg not found",
		},
		{
			name: "Errors when the updater errors",
			args: params{
				pkg: &datatypes.Package{
					Name:       "Test package",
					WorkflowID: "workflow-1",
					RunID:      runID.String(),
					AIPID:      aipID,
				},
				updater: func(p *datatypes.Package) (*datatypes.Package, error) {
					return nil, fmt.Errorf("Bad input")
				},
			},
			wantErr: "invalid data error: updater error: Bad input",
		},
		{
			name: "Errors when updater sets an invalid RunID",
			args: params{
				pkg: &datatypes.Package{
					Name:       "Test package",
					WorkflowID: "workflow-1",
					RunID:      runID.String(),
					AIPID:      aipID,
				},
				updater: func(p *datatypes.Package) (*datatypes.Package, error) {
					p.RunID = "Bad UUID"
					return p, nil
				},
			},
			wantErr: "invalid data error: parse error: field \"RunID\": invalid UUID length: 8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, svc := setUpClient(t, logr.Discard())
			ctx := context.Background()

			var id int
			if tt.args.pkg != nil {
				pkg := *tt.args.pkg // Make a local copy of pkg.
				err := svc.CreatePackage(ctx, &pkg)
				assert.NilError(t, err)

				id = pkg.ID
			}

			pp, err := svc.UpdatePackage(ctx, id, tt.args.updater)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}

			assert.DeepEqual(t, pp, tt.want,
				cmpopts.EquateApproxTime(time.Millisecond*100),
				cmpopts.IgnoreUnexported(db.Pkg{}, db.PkgEdges{}),
			)
		})
	}
}

func TestListPackages(t *testing.T) {
	t.Parallel()

	runID := uuid.MustParse("c5f7c35a-d5a6-4e00-b4da-b036ce5b40bc")
	runID2 := uuid.MustParse("c04d0191-d7ce-46dd-beff-92d6830082ff")

	aipID := uuid.NullUUID{
		UUID:  uuid.MustParse("e2ace0da-8697-453d-9ea1-4c9b62309e54"),
		Valid: true,
	}
	aipID2 := uuid.NullUUID{
		UUID:  uuid.MustParse("7d085541-af56-4444-9ce2-d6401ff4c97b"),
		Valid: true,
	}

	locID := uuid.NullUUID{
		UUID:  uuid.MustParse("146182ff-9923-4869-bca1-0bbc0f822025"),
		Valid: true,
	}
	locID2 := uuid.NullUUID{
		UUID:  uuid.MustParse("6e30694b-6497-439f-bf99-83af165e02c3"),
		Valid: true,
	}

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
		data []*datatypes.Package
		page *persistence.Page
	}
	tests := []struct {
		name          string
		data          []*datatypes.Package
		packageFilter *persistence.PackageFilter
		want          results
		wantErr       string
	}{
		{
			name: "Returns all packages",
			data: []*datatypes.Package{
				{
					Name:        "Test package 1",
					WorkflowID:  "workflow-1",
					RunID:       runID.String(),
					AIPID:       aipID,
					LocationID:  locID,
					Status:      enums.PackageStatusDone,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					Name:        "Test package 2",
					WorkflowID:  "workflow-1",
					RunID:       runID2.String(),
					AIPID:       aipID2,
					LocationID:  locID2,
					Status:      enums.PackageStatusInProgress,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			want: results{
				data: []*datatypes.Package{
					{
						ID:          1,
						Name:        "Test package 1",
						WorkflowID:  "workflow-1",
						RunID:       runID.String(),
						AIPID:       aipID,
						LocationID:  locID,
						Status:      enums.PackageStatusDone,
						CreatedAt:   time.Now(),
						StartedAt:   started,
						CompletedAt: completed,
					},
					{
						ID:          2,
						Name:        "Test package 2",
						WorkflowID:  "workflow-1",
						RunID:       runID2.String(),
						AIPID:       aipID2,
						LocationID:  locID2,
						Status:      enums.PackageStatusInProgress,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
					},
				},
				page: &persistence.Page{
					Limit: entclient.DefaultPageSize,
					Total: 2,
				},
			},
		},
		{
			name: "Returns first page of packages",
			data: []*datatypes.Package{
				{
					Name:        "Test package 1",
					WorkflowID:  "workflow-1",
					RunID:       runID.String(),
					AIPID:       aipID,
					LocationID:  locID,
					Status:      enums.PackageStatusDone,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					Name:        "Test package 2",
					WorkflowID:  "workflow-1",
					RunID:       runID2.String(),
					AIPID:       aipID2,
					LocationID:  locID2,
					Status:      enums.PackageStatusInProgress,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			packageFilter: &persistence.PackageFilter{
				Page: persistence.Page{Limit: 1},
			},
			want: results{
				data: []*datatypes.Package{
					{
						ID:          1,
						Name:        "Test package 1",
						WorkflowID:  "workflow-1",
						RunID:       runID.String(),
						AIPID:       aipID,
						LocationID:  locID,
						Status:      enums.PackageStatusDone,
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
			name: "Returns second page of packages",
			data: []*datatypes.Package{
				{
					Name:        "Test package 1",
					WorkflowID:  "workflow-1",
					RunID:       runID.String(),
					AIPID:       aipID,
					LocationID:  locID,
					Status:      enums.PackageStatusDone,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					Name:        "Test package 2",
					WorkflowID:  "workflow-1",
					RunID:       runID2.String(),
					AIPID:       aipID2,
					LocationID:  locID2,
					Status:      enums.PackageStatusInProgress,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			packageFilter: &persistence.PackageFilter{
				Page: persistence.Page{Limit: 1, Offset: 1},
			},
			want: results{
				data: []*datatypes.Package{
					{
						ID:          2,
						Name:        "Test package 2",
						WorkflowID:  "workflow-1",
						RunID:       runID2.String(),
						AIPID:       aipID2,
						LocationID:  locID2,
						Status:      enums.PackageStatusInProgress,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
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
			name: "Returns packages whose names contain a string",
			data: []*datatypes.Package{
				{
					Name:        "Test package",
					WorkflowID:  "workflow-1",
					RunID:       runID.String(),
					AIPID:       aipID,
					LocationID:  locID,
					Status:      enums.PackageStatusDone,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					Name:        "small.zip",
					WorkflowID:  "workflow-1",
					RunID:       runID2.String(),
					AIPID:       aipID2,
					LocationID:  locID2,
					Status:      enums.PackageStatusInProgress,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			packageFilter: &persistence.PackageFilter{
				Name: ref.New("small"),
			},
			want: results{
				data: []*datatypes.Package{
					{
						ID:          2,
						Name:        "small.zip",
						WorkflowID:  "workflow-1",
						RunID:       runID2.String(),
						AIPID:       aipID2,
						LocationID:  locID2,
						Status:      enums.PackageStatusInProgress,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
					},
				},
				page: &persistence.Page{
					Limit: entclient.DefaultPageSize,
					Total: 1,
				},
			},
		},
		{
			name: "Returns packages filtered by AIPID",
			data: []*datatypes.Package{
				{
					Name:        "Test package 1",
					WorkflowID:  "workflow-1",
					RunID:       runID.String(),
					AIPID:       aipID,
					LocationID:  locID,
					Status:      enums.PackageStatusDone,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					Name:        "Test package 2",
					WorkflowID:  "workflow-1",
					RunID:       runID2.String(),
					AIPID:       aipID2,
					LocationID:  locID2,
					Status:      enums.PackageStatusInProgress,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			packageFilter: &persistence.PackageFilter{
				AIPID: &aipID2.UUID,
			},
			want: results{
				data: []*datatypes.Package{
					{
						ID:          2,
						Name:        "Test package 2",
						WorkflowID:  "workflow-1",
						RunID:       runID2.String(),
						AIPID:       aipID2,
						LocationID:  locID2,
						Status:      enums.PackageStatusInProgress,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
					},
				},
				page: &persistence.Page{
					Limit: entclient.DefaultPageSize,
					Total: 1,
				},
			},
		},
		{
			name: "Returns packages filtered by LocationID",
			data: []*datatypes.Package{
				{
					Name:        "Test package 1",
					WorkflowID:  "workflow-1",
					RunID:       runID.String(),
					AIPID:       aipID,
					LocationID:  locID,
					Status:      enums.PackageStatusDone,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					Name:        "Test package 2",
					WorkflowID:  "workflow-1",
					RunID:       runID2.String(),
					AIPID:       aipID2,
					LocationID:  locID2,
					Status:      enums.PackageStatusInProgress,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			packageFilter: &persistence.PackageFilter{
				LocationID: &locID2.UUID,
			},
			want: results{
				data: []*datatypes.Package{
					{
						ID:          2,
						Name:        "Test package 2",
						WorkflowID:  "workflow-1",
						RunID:       runID2.String(),
						AIPID:       aipID2,
						LocationID:  locID2,
						Status:      enums.PackageStatusInProgress,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
					},
				},
				page: &persistence.Page{
					Limit: entclient.DefaultPageSize,
					Total: 1,
				},
			},
		},
		{
			name: "Returns packages filtered by status",
			data: []*datatypes.Package{
				{
					Name:        "Test package 1",
					WorkflowID:  "workflow-1",
					RunID:       runID.String(),
					AIPID:       aipID,
					LocationID:  locID,
					Status:      enums.PackageStatusDone,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					Name:        "Test package 2",
					WorkflowID:  "workflow-1",
					RunID:       runID2.String(),
					AIPID:       aipID2,
					LocationID:  locID2,
					Status:      enums.PackageStatusInProgress,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			packageFilter: &persistence.PackageFilter{
				Status: ref.New(enums.PackageStatusInProgress),
			},
			want: results{
				data: []*datatypes.Package{
					{
						ID:          2,
						Name:        "Test package 2",
						WorkflowID:  "workflow-1",
						RunID:       runID2.String(),
						AIPID:       aipID2,
						LocationID:  locID2,
						Status:      enums.PackageStatusInProgress,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
					},
				},
				page: &persistence.Page{
					Limit: entclient.DefaultPageSize,
					Total: 1,
				},
			},
		},
		{
			name: "Returns packages filtered by CreatedAt",
			data: []*datatypes.Package{
				{
					Name:        "Test package 1",
					WorkflowID:  "workflow-1",
					RunID:       runID.String(),
					AIPID:       aipID,
					LocationID:  locID,
					Status:      enums.PackageStatusDone,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					Name:        "Test package 2",
					WorkflowID:  "workflow-1",
					RunID:       runID2.String(),
					AIPID:       aipID2,
					LocationID:  locID2,
					Status:      enums.PackageStatusInProgress,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			packageFilter: &persistence.PackageFilter{
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
				data: []*datatypes.Package{
					{
						ID:          1,
						Name:        "Test package 1",
						WorkflowID:  "workflow-1",
						RunID:       runID.String(),
						AIPID:       aipID,
						LocationID:  locID,
						Status:      enums.PackageStatusDone,
						CreatedAt:   time.Now(),
						StartedAt:   started,
						CompletedAt: completed,
					},
					{
						ID:          2,
						Name:        "Test package 2",
						WorkflowID:  "workflow-1",
						RunID:       runID2.String(),
						AIPID:       aipID2,
						LocationID:  locID2,
						Status:      enums.PackageStatusInProgress,
						CreatedAt:   time.Now(),
						StartedAt:   started2,
						CompletedAt: completed2,
					},
				},
				page: &persistence.Page{
					Limit: entclient.DefaultPageSize,
					Total: 2,
				},
			},
		},
		{
			name: "Returns no results when no packages match CreatedAt range",
			data: []*datatypes.Package{
				{
					Name:        "Test package 1",
					WorkflowID:  "workflow-1",
					RunID:       runID.String(),
					AIPID:       aipID,
					LocationID:  locID,
					Status:      enums.PackageStatusDone,
					StartedAt:   started,
					CompletedAt: completed,
				},
				{
					Name:        "Test package 2",
					WorkflowID:  "workflow-1",
					RunID:       runID2.String(),
					AIPID:       aipID2,
					LocationID:  locID2,
					Status:      enums.PackageStatusInProgress,
					StartedAt:   started2,
					CompletedAt: completed2,
				},
			},
			packageFilter: &persistence.PackageFilter{
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
				data: []*datatypes.Package{},
				page: &persistence.Page{
					Limit: entclient.DefaultPageSize,
					Total: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, svc := setUpClient(t, logr.Discard())
			ctx := context.Background()

			if len(tt.data) > 0 {
				for _, pkg := range tt.data {
					err := svc.CreatePackage(ctx, pkg)
					assert.NilError(t, err)
				}
			}

			got, pg, err := svc.ListPackages(ctx, tt.packageFilter)
			assert.NilError(t, err)

			assert.DeepEqual(t, got, tt.want.data,
				cmpopts.EquateApproxTime(time.Millisecond*100),
				cmpopts.IgnoreUnexported(db.Pkg{}, db.PkgEdges{}),
			)
			assert.DeepEqual(t, pg, tt.want.page)
		})
	}
}
