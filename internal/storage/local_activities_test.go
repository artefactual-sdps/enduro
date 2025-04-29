package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	"go.artefactual.dev/tools/mockutil"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/fake"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func TestReadAIPLocalActivity(t *testing.T) {
	t.Parallel()

	svc := fake.NewMockService(gomock.NewController(t))
	ctx := context.Background()
	aip := &goastorage.AIP{Name: "Name", UUID: aipID}
	svc.EXPECT().ReadAip(ctx, aipID).Return(aip, nil)

	re, err := storage.ReadAIPLocalActivity(ctx, svc, aipID)
	assert.NilError(t, err)
	assert.DeepEqual(t, re, aip)
}

func TestUpdateAIPLocationLocalActivity(t *testing.T) {
	t.Parallel()

	svc := fake.NewMockService(gomock.NewController(t))
	ctx := context.Background()
	svc.EXPECT().UpdateAipLocationID(ctx, aipID, locationID).Return(nil)

	err := storage.UpdateAIPLocationLocalActivity(ctx, svc, &storage.UpdateAIPLocationLocalActivityParams{
		AIPID:      aipID,
		LocationID: locationID,
	})
	assert.NilError(t, err)
}

func TestUpdateAIPStatusLocalActivity(t *testing.T) {
	t.Parallel()

	svc := fake.NewMockService(gomock.NewController(t))
	ctx := context.Background()
	status := enums.AIPStatusStored
	svc.EXPECT().UpdateAipStatus(ctx, aipID, status).Return(nil)

	err := storage.UpdateAIPStatusLocalActivity(ctx, svc, &storage.UpdateAIPStatusLocalActivityParams{
		AIPID:  aipID,
		Status: status,
	})
	assert.NilError(t, err)
}

func TestDeleteFromMinIOLocationLocalActivity(t *testing.T) {
	t.Parallel()

	svc := fake.NewMockService(gomock.NewController(t))
	ctx := context.Background()
	svc.EXPECT().DeleteAip(ctx, aipID).Return(nil)

	err := storage.DeleteFromMinIOLocationLocalActivity(ctx, svc, &storage.DeleteFromMinIOLocationLocalActivityParams{
		AIPID: aipID,
	})
	assert.NilError(t, err)
}

func TestReadLocationInfoLocalActivity(t *testing.T) {
	t.Parallel()

	svc := fake.NewMockService(gomock.NewController(t))
	ctx := context.Background()
	location := &goastorage.Location{
		UUID:   locationID,
		Source: enums.LocationSourceMinio.String(),
		Config: &goastorage.URLConfig{URL: "mem://"},
	}
	svc.EXPECT().ReadLocation(ctx, locationID).Return(location, nil)

	re, err := storage.ReadLocationInfoLocalActivity(ctx, svc, locationID)
	assert.NilError(t, err)
	assert.DeepEqual(t, re, &storage.ReadLocationInfoLocalActivityResult{
		Source: enums.LocationSourceMinio,
		Config: types.LocationConfig{Value: &types.URLConfig{URL: "mem://"}},
	})
}

func TestCreateWorkflowLocalActivity(t *testing.T) {
	t.Parallel()

	svc := fake.NewMockService(gomock.NewController(t))
	ctx := context.Background()
	dbID := 1
	workflow := &types.Workflow{
		TemporalID: "temporal-id",
		Type:       enums.WorkflowTypeDeleteAip,
		Status:     enums.WorkflowStatusInProgress,
		StartedAt:  time.Now(),
		AIPUUID:    aipID,
	}
	svc.EXPECT().
		CreateWorkflow(
			ctx,
			mockutil.Eq(workflow, mockutil.EquateNearlySameTime(), cmpopts.IgnoreFields(types.Workflow{}, "UUID")),
		).
		DoAndReturn(func(ctx context.Context, w *types.Workflow) error {
			w.DBID = dbID
			return nil
		})

	re, err := storage.CreateWorkflowLocalActivity(ctx, svc, &storage.CreateWorkflowLocalActivityParams{
		AIPID:      aipID,
		TemporalID: workflow.TemporalID,
		Type:       workflow.Type,
	})
	assert.NilError(t, err)
	assert.Equal(t, re, dbID)
}

func TestUpdateWorkflowStatusLocalActivity(t *testing.T) {
	t.Parallel()

	svc := fake.NewMockService(gomock.NewController(t))
	ctx := context.Background()
	dbID := 1
	svc.EXPECT().
		UpdateWorkflow(
			ctx,
			dbID,
			mockutil.Func(
				"should update workflow",
				func(updater persistence.WorkflowUpdater) error {
					w, err := updater(&types.Workflow{})
					assert.NilError(t, err)
					assert.DeepEqual(t, w.Status, enums.WorkflowStatusDone)
					return nil
				},
			),
		).
		Return(nil, nil)

	err := storage.UpdateWorkflowStatusLocalActivity(ctx, svc, &storage.UpdateWorkflowStatusLocalActivityParams{
		DBID:   dbID,
		Status: enums.WorkflowStatusDone,
	})
	assert.NilError(t, err)
}

func TestCompleteWorkflowLocalActivity(t *testing.T) {
	t.Parallel()

	svc := fake.NewMockService(gomock.NewController(t))
	ctx := context.Background()
	dbID := 1
	svc.EXPECT().
		UpdateWorkflow(
			ctx,
			dbID,
			mockutil.Func(
				"should update workflow",
				func(updater persistence.WorkflowUpdater) error {
					w, err := updater(&types.Workflow{})
					assert.NilError(t, err)
					assert.DeepEqual(t, w.Status, enums.WorkflowStatusDone)
					assert.DeepEqual(t, w.CompletedAt, time.Now(), mockutil.EquateNearlySameTime())
					return nil
				},
			),
		).
		Return(nil, nil)

	err := storage.CompleteWorkflowLocalActivity(ctx, svc, &storage.CompleteWorkflowLocalActivityParams{
		DBID:   dbID,
		Status: enums.WorkflowStatusDone,
	})
	assert.NilError(t, err)
}

func TestCreateTaskLocalActivity(t *testing.T) {
	t.Parallel()

	svc := fake.NewMockService(gomock.NewController(t))
	ctx := context.Background()
	dbID := 1
	task := &types.Task{
		Name:         "name",
		Status:       enums.TaskStatusInProgress,
		StartedAt:    time.Now(),
		Note:         "note",
		WorkflowDBID: 1,
	}
	svc.EXPECT().
		CreateTask(
			ctx,
			mockutil.Eq(task, mockutil.EquateNearlySameTime(), cmpopts.IgnoreFields(types.Task{}, "UUID")),
		).
		DoAndReturn(func(ctx context.Context, t *types.Task) error {
			t.DBID = dbID
			return nil
		})

	re, err := storage.CreateTaskLocalActivity(ctx, svc, &storage.CreateTaskLocalActivityParams{
		WorkflowDBID: task.WorkflowDBID,
		Status:       task.Status,
		Name:         task.Name,
		Note:         task.Note,
	})
	assert.NilError(t, err)
	assert.Equal(t, re, dbID)
}

func TestCompleteTaskLocalActivity(t *testing.T) {
	t.Parallel()

	svc := fake.NewMockService(gomock.NewController(t))
	ctx := context.Background()
	dbID := 1
	note := "note"
	svc.EXPECT().
		UpdateTask(
			ctx,
			dbID,
			mockutil.Func(
				"should update task",
				func(updater persistence.TaskUpdater) error {
					task, err := updater(&types.Task{})
					assert.NilError(t, err)
					assert.DeepEqual(t, task.Status, enums.TaskStatusDone)
					assert.DeepEqual(t, task.CompletedAt, time.Now(), mockutil.EquateNearlySameTime())
					assert.DeepEqual(t, task.Note, note)
					return nil
				},
			),
		).
		Return(nil, nil)

	err := storage.CompleteTaskLocalActivity(ctx, svc, &storage.CompleteTaskLocalActivityParams{
		DBID:   dbID,
		Status: enums.TaskStatusDone,
		Note:   note,
	})
	assert.NilError(t, err)
}

func TestCreateDeletionRequestLocalActivity(t *testing.T) {
	t.Parallel()

	svc := fake.NewMockService(gomock.NewController(t))
	ctx := context.Background()
	dbID := 1
	dr := &types.DeletionRequest{
		Requester:    "requester@example.com",
		RequesterISS: "issuer",
		RequesterSub: "subject",
		RequestedAt:  time.Now(),
		Reason:       "Reason",
		Status:       enums.DeletionRequestStatusPending,
		AIPUUID:      aipID,
		WorkflowDBID: 1,
	}
	svc.EXPECT().
		CreateDeletionRequest(
			ctx,
			mockutil.Eq(dr, mockutil.EquateNearlySameTime(), cmpopts.IgnoreFields(types.DeletionRequest{}, "UUID")),
		).
		DoAndReturn(func(ctx context.Context, d *types.DeletionRequest) error {
			d.DBID = dbID
			return nil
		})

	re, err := storage.CreateDeletionRequestLocalActivity(ctx, svc, &storage.CreateDeletionRequestLocalActivityParams{
		Requester:    dr.Requester,
		RequesterISS: dr.RequesterISS,
		RequesterSub: dr.RequesterSub,
		Reason:       dr.Reason,
		AIPUUID:      dr.AIPUUID,
		WorkflowDBID: dr.WorkflowDBID,
	})
	assert.NilError(t, err)
	assert.Equal(t, re, dbID)
}

func TestReviewDeletionRequestLocalActivity(t *testing.T) {
	t.Parallel()

	svc := fake.NewMockService(gomock.NewController(t))
	ctx := context.Background()
	dbID := 1
	drs := storage.DeletionReviewedSignal{
		Approved:  true,
		UserEmail: "reviewer@example.com",
		UserISS:   "issuer",
		UserSub:   "subject-2",
	}
	svc.EXPECT().
		UpdateDeletionRequest(
			ctx,
			dbID,
			mockutil.Func(
				"should update deletion request",
				func(updater persistence.DeletionRequestUpdater) error {
					dr, err := updater(&types.DeletionRequest{})
					assert.NilError(t, err)
					assert.DeepEqual(t, dr.Reviewer, drs.UserEmail)
					assert.DeepEqual(t, dr.ReviewerISS, drs.UserISS)
					assert.DeepEqual(t, dr.ReviewerSub, drs.UserSub)
					assert.DeepEqual(t, dr.ReviewedAt, time.Now(), mockutil.EquateNearlySameTime())
					assert.DeepEqual(t, dr.Status, enums.DeletionRequestStatusApproved)
					return nil
				},
			),
		).
		Return(nil, nil)

	err := storage.ReviewDeletionRequestLocalActivity(ctx, svc, dbID, drs)
	assert.NilError(t, err)
}
