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
	_ "github.com/mattn/go-sqlite3"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/package_"
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
			SetStatus(int8(package_.NewStatus("in progress"))).
			Save(context.Background())

		assert.NilError(t, err)
		assert.Equal(t, p.Name, "testing 1-2-3")
		assert.Equal(t, p.WorkflowID, "12345")
		assert.Equal(t, p.RunID, runID)
		assert.Equal(t, p.AipID, aipID)
		assert.Equal(t, p.Status, int8(package_.StatusInProgress))
	})
}

func TestCreatePackage(t *testing.T) {
	runID := uuid.New()
	aipID := uuid.New()
	locID := uuid.NullUUID{UUID: uuid.New(), Valid: true}
	started := sql.NullTime{Time: time.Now(), Valid: true}
	completed := sql.NullTime{Time: started.Time.Add(time.Second), Valid: true}

	type params struct {
		pkg *package_.Package
	}
	tests := []struct {
		name    string
		args    params
		want    *package_.Package
		wantErr string
	}{
		{
			name: "Saves a new package in the DB",
			args: params{
				pkg: &package_.Package{
					Name:        "Test package 1",
					WorkflowID:  "workflow-1",
					RunID:       runID.String(),
					AIPID:       aipID.String(),
					LocationID:  locID,
					Status:      package_.StatusInProgress,
					StartedAt:   started,
					CompletedAt: completed,
				},
			},
			want: &package_.Package{
				ID:          1,
				Name:        "Test package 1",
				WorkflowID:  "workflow-1",
				RunID:       runID.String(),
				AIPID:       aipID.String(),
				LocationID:  locID,
				Status:      package_.StatusInProgress,
				CreatedAt:   time.Now(),
				StartedAt:   started,
				CompletedAt: completed,
			},
		},
		{
			name: "Saves a package with missing optional fields",
			args: params{
				pkg: &package_.Package{
					Name:       "Test package 2",
					WorkflowID: "workflow-2",
					RunID:      runID.String(),
					AIPID:      aipID.String(),
					Status:     package_.StatusInProgress,
				},
			},
			want: &package_.Package{
				ID:         1,
				Name:       "Test package 2",
				WorkflowID: "workflow-2",
				RunID:      runID.String(),
				AIPID:      aipID.String(),
				Status:     package_.StatusInProgress,
				CreatedAt:  time.Now(),
			},
		},
		{
			name: "Required field error for missing Name",
			args: params{
				pkg: &package_.Package{},
			},
			wantErr: "invalid data error: field \"Name\" is required",
		},
		{
			name: "Required field error for missing WorkflowID",
			args: params{
				pkg: &package_.Package{
					Name: "Missing WorkflowID",
				},
			},
			wantErr: "invalid data error: field \"WorkflowID\" is required",
		},
		{
			name: "Required field error for missing AIPID",
			args: params{
				pkg: &package_.Package{
					Name:       "Missing AIPID",
					WorkflowID: "workflow-12345",
					RunID:      runID.String(),
				},
			},
			wantErr: "invalid data error: field \"AIPID\" is required",
		},
		{
			name: "Required field error for missing RunID",
			args: params{
				pkg: &package_.Package{
					Name:       "Missing RunID",
					WorkflowID: "workflow-12345",
				},
			},
			wantErr: "invalid data error: field \"RunID\" is required",
		},
		{
			name: "Errors on invalid RunID",
			args: params{
				pkg: &package_.Package{
					Name:       "Invalid package 1",
					WorkflowID: "workflow-invalid",
					RunID:      "Bad UUID",
				},
			},
			wantErr: "invalid data error: parse error: field \"RunID\": invalid UUID length: 8",
		},
		{
			name: "Errors on invalid AIPID",
			args: params{
				pkg: &package_.Package{
					Name:       "Invalid package 2",
					WorkflowID: "workflow-invalid",
					RunID:      runID.String(),
					AIPID:      "Bad UUID",
				},
			},
			wantErr: "invalid data error: parse error: field \"AIPID\": invalid UUID length: 8",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, svc := setUpClient(t, logr.Discard())
			ctx := context.Background()

			pkg, err := svc.CreatePackage(ctx, tt.args.pkg)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			assert.DeepEqual(t, pkg, tt.want,
				cmpopts.EquateApproxTime(time.Millisecond*100),
				cmpopts.IgnoreUnexported(db.Pkg{}, db.PkgEdges{}),
			)
		})
	}
}
