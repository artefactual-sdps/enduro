package workflow

import (
	"database/sql"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/artefactual-sdps/temporal-activities/archiveextract"
	"github.com/artefactual-sdps/temporal-activities/archivezip"
	"github.com/artefactual-sdps/temporal-activities/bagcreate"
	"github.com/artefactual-sdps/temporal-activities/bagvalidate"
	"github.com/artefactual-sdps/temporal-activities/removepaths"
	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.artefactual.dev/amclient/amclienttest"
	"go.artefactual.dev/tools/ref"
	"go.opentelemetry.io/otel/trace/noop"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	temporalsdk_worker "go.temporal.io/sdk/worker"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/a3m"
	a3mfake "github.com/artefactual-sdps/enduro/internal/a3m/fake"
	"github.com/artefactual-sdps/enduro/internal/am"
	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/package_"
	packagefake "github.com/artefactual-sdps/enduro/internal/package_/fake"
	"github.com/artefactual-sdps/enduro/internal/preprocessing"
	"github.com/artefactual-sdps/enduro/internal/pres"
	sftp_fake "github.com/artefactual-sdps/enduro/internal/sftp/fake"
	"github.com/artefactual-sdps/enduro/internal/storage"
	"github.com/artefactual-sdps/enduro/internal/temporal"
	watcherfake "github.com/artefactual-sdps/enduro/internal/watcher/fake"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
	"github.com/artefactual-sdps/enduro/internal/workflow/localact"
)

const (
	tempPath     string = "/tmp/enduro123456"
	extractPath  string = "/tmp/enduro123456/extract"
	transferPath string = "/home/a3m/.local/share/a3m/share/enduro2985726865"
)

var (
	locationID     = uuid.MustParse("f2cc963f-c14d-4eaa-b950-bd207189a1f1")
	amssLocationID = uuid.MustParse("e0ed8b2a-8ae2-4546-b5d8-f0090919df04")
	transferID     = uuid.MustParse("65233405-771e-4f7e-b2d9-b08439570ba2")
	sipID          = uuid.MustParse("9e8161cc-2815-4d6f-8a75-f003c41b257b")

	startTime time.Time = time.Date(2024, 7, 9, 16, 55, 13, 50, time.UTC)
)

type ProcessingWorkflowTestSuite struct {
	suite.Suite
	temporalsdk_testsuite.WorkflowTestSuite

	env *temporalsdk_testsuite.TestWorkflowEnvironment

	// Each test creates its own temporary transfer directory.
	transferDir string

	// Each test registers the workflow with a different name to avoid
	// duplicates.
	workflow *ProcessingWorkflow
}

func TestTransferInfo_Name(t *testing.T) {
	t.Run("Returns name of transfer", func(t *testing.T) {
		tinfo := TransferInfo{}
		tinfo.req.Key = "somename.tar.gz"
		assert.Equal(t, tinfo.Name(), "somename")
	})
}

func preprocessingChildWorkflow(
	ctx temporalsdk_workflow.Context,
	params *preprocessing.WorkflowParams,
) (*preprocessing.WorkflowResult, error) {
	return nil, nil
}

func (s *ProcessingWorkflowTestSuite) CreateTransferDir() string {
	s.transferDir = s.T().TempDir()

	return s.transferDir
}

func (s *ProcessingWorkflowTestSuite) SetupWorkflowTest(cfg config.Configuration) {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.SetWorkerOptions(temporalsdk_worker.Options{EnableSessionWorker: true})
	s.env.SetStartTime(startTime)

	ctrl := gomock.NewController(s.T())
	pkgsvc := packagefake.NewMockService(ctrl)
	wsvc := watcherfake.NewMockService(ctrl)
	rng := rand.New(rand.NewSource(1)) // #nosec: G404

	s.env.RegisterActivityWithOptions(
		activities.NewDownloadActivity(noop.Tracer{}, wsvc).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.DownloadActivityName},
	)
	s.env.RegisterActivityWithOptions(
		archiveextract.New(cfg.ExtractActivity).Execute,
		temporalsdk_activity.RegisterOptions{Name: archiveextract.Name},
	)
	s.env.RegisterActivityWithOptions(
		bagvalidate.New(bagvalidate.NewNoopValidator()).Execute,
		temporalsdk_activity.RegisterOptions{Name: bagvalidate.Name},
	)
	s.env.RegisterActivityWithOptions(
		activities.NewClassifyPackageActivity().Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.ClassifyPackageActivityName},
	)

	// Set up AM taskqueue.
	if cfg.Preservation.TaskQueue == temporal.AmWorkerTaskQueue {
		s.setupAMWorkflowTest(&cfg.AM, ctrl, pkgsvc)
	} else {
		s.setupA3mWorkflowTest(ctrl, pkgsvc)
	}

	s.env.RegisterWorkflowWithOptions(
		preprocessingChildWorkflow,
		temporalsdk_workflow.RegisterOptions{Name: "preprocessing"},
	)
	s.env.RegisterActivityWithOptions(
		removepaths.New().Execute,
		temporalsdk_activity.RegisterOptions{Name: removepaths.Name},
	)
	s.env.RegisterActivityWithOptions(
		activities.NewDeleteOriginalActivity(wsvc).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.DeleteOriginalActivityName},
	)

	s.workflow = NewProcessingWorkflow(cfg, rng, pkgsvc, wsvc)
}

func (s *ProcessingWorkflowTestSuite) setupAMWorkflowTest(
	cfg *am.Config,
	ctrl *gomock.Controller,
	pkgsvc package_.Service,
) {
	clock := clockwork.NewFakeClock()
	sftpc := sftp_fake.NewMockClient(ctrl)

	s.env.RegisterActivityWithOptions(
		bagcreate.New(bagcreate.Config{}).Execute,
		temporalsdk_activity.RegisterOptions{Name: bagcreate.Name},
	)
	s.env.RegisterActivityWithOptions(
		archivezip.New().Execute,
		temporalsdk_activity.RegisterOptions{Name: archivezip.Name},
	)
	s.env.RegisterActivityWithOptions(
		am.NewUploadTransferActivity(sftpc, 10*time.Second).Execute,
		temporalsdk_activity.RegisterOptions{Name: am.UploadTransferActivityName},
	)
	s.env.RegisterActivityWithOptions(
		am.NewStartTransferActivity(&am.Config{}, amclienttest.NewMockPackageService(ctrl)).Execute,
		temporalsdk_activity.RegisterOptions{Name: am.StartTransferActivityName},
	)
	s.env.RegisterActivityWithOptions(
		am.NewPollTransferActivity(
			cfg,
			clock,
			amclienttest.NewMockTransferService(ctrl),
			amclienttest.NewMockJobsService(ctrl),
			pkgsvc,
		).Execute,
		temporalsdk_activity.RegisterOptions{Name: am.PollTransferActivityName},
	)
	s.env.RegisterActivityWithOptions(
		am.NewPollIngestActivity(
			cfg,
			clock,
			amclienttest.NewMockIngestService(ctrl),
			amclienttest.NewMockJobsService(ctrl),
			pkgsvc,
		).Execute,
		temporalsdk_activity.RegisterOptions{Name: am.PollIngestActivityName},
	)
	s.env.RegisterActivityWithOptions(
		activities.NewCreateStoragePackageActivity(nil).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.CreateStoragePackageActivityName},
	)
	s.env.RegisterActivityWithOptions(
		am.NewDeleteTransferActivity(sftpc).Execute,
		temporalsdk_activity.RegisterOptions{
			Name: am.DeleteTransferActivityName,
		},
	)
}

func (s *ProcessingWorkflowTestSuite) setupA3mWorkflowTest(
	ctrl *gomock.Controller,
	pkgsvc package_.Service,
) {
	tsvc := a3mfake.NewMockTransferServiceClient(ctrl)

	s.env.RegisterActivityWithOptions(
		activities.NewBundleActivity().Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.BundleActivityName},
	)
	s.env.RegisterActivityWithOptions(
		a3m.NewCreateAIPActivity(noop.Tracer{}, tsvc, &a3m.Config{}, pkgsvc).Execute,
		temporalsdk_activity.RegisterOptions{Name: a3m.CreateAIPActivityName},
	)
	s.env.RegisterActivityWithOptions(
		activities.NewUploadActivity(nil).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.UploadActivityName},
	)
	s.env.RegisterActivityWithOptions(
		activities.NewMoveToPermanentStorageActivity(nil).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.MoveToPermanentStorageActivityName},
	)
	s.env.RegisterActivityWithOptions(
		activities.NewPollMoveToPermanentStorageActivity(nil).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.PollMoveToPermanentStorageActivityName},
	)
	s.env.RegisterActivityWithOptions(
		activities.NewRejectPackageActivity(nil).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.RejectPackageActivityName},
	)
}

func (s *ProcessingWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func TestProcessingWorkflow(t *testing.T) {
	suite.Run(t, new(ProcessingWorkflowTestSuite))
}

func (s *ProcessingWorkflowTestSuite) TestPackageConfirmation() {
	cfg := config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Storage: storage.Config{
			DefaultPermanentLocationID: locationID,
		},
	}
	s.SetupWorkflowTest(cfg)

	pkgID := uint(1)
	ctx := mock.AnythingOfType("*context.valueCtx")
	sessionCtx := mock.AnythingOfType("*context.timerCtx")
	key := "transfer.zip"
	watcherName := "watcher"
	retentionPeriod := 1 * time.Second
	pkgsvc := s.workflow.pkgsvc

	// Signal handler that mimics package confirmation
	s.env.RegisterDelayedCallback(
		func() {
			s.env.SignalWorkflow(
				package_.ReviewPerformedSignalName,
				package_.ReviewPerformedSignal{Accepted: true, LocationID: &locationID},
			)
		},
		0,
	)

	// Activity mocks/assertions sequence
	s.env.OnActivity(createPackageLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(pkgID, nil)
	s.env.OnActivity(setStatusInProgressLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(
		createPreservationActionLocalActivity,
		mock.Anything,
		mock.Anything,
		&createPreservationActionLocalActivityParams{
			WorkflowID: "default-test-workflow-id",
			Type:       enums.PreservationActionTypeCreateAndReviewAip,
			Status:     enums.PreservationActionStatusInProgress,
			StartedAt:  startTime,
			PackageID:  1,
		},
	).Return(uint(0), nil)

	s.env.OnActivity(activities.DownloadActivityName, sessionCtx,
		&activities.DownloadActivityParams{Key: key, WatcherName: watcherName},
	).Return(
		&activities.DownloadActivityResult{Path: tempPath + "/" + key}, nil,
	)

	s.env.OnActivity(archiveextract.Name, sessionCtx,
		&archiveextract.Params{SourcePath: tempPath + "/" + key},
	).Return(
		&archiveextract.Result{ExtractPath: extractPath}, nil,
	)

	s.env.OnActivity(activities.BundleActivityName, sessionCtx,
		&activities.BundleActivityParams{
			SourcePath:  extractPath,
			TransferDir: s.transferDir,
			IsDir:       true,
		},
	).Return(
		&activities.BundleActivityResult{FullPath: transferPath},
		nil,
	)

	s.env.OnActivity(a3m.CreateAIPActivityName, mock.Anything, mock.Anything).Return(nil, nil)
	s.env.OnActivity(updatePackageLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(createPreservationTaskLocalActivity, mock.Anything, mock.Anything, mock.Anything).
		Return(uint(0), nil)
	s.env.OnActivity(activities.UploadActivityName, mock.Anything, mock.Anything).Return(nil, nil)
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(setPreservationActionStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(completePreservationTaskLocalActivity, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(activities.MoveToPermanentStorageActivityName, mock.Anything, mock.Anything).Return(nil, nil)
	s.env.OnActivity(activities.PollMoveToPermanentStorageActivityName, mock.Anything, mock.Anything).Return(nil, nil)
	s.env.OnActivity(setLocationIDLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(completePreservationActionLocalActivity, ctx, pkgsvc, mock.AnythingOfType("*workflow.completePreservationActionLocalActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(
		removepaths.Name,
		sessionCtx,
		&removepaths.Params{Paths: []string{tempPath, transferPath}},
	).Return(&removepaths.Result{}, nil)
	s.env.OnActivity(activities.DeleteOriginalActivityName, sessionCtx, watcherName, key).Return(nil, nil).Once()
	s.env.OnActivity(updatePackageLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(completePreservationActionLocalActivity, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&package_.ProcessingWorkflowRequest{
			Key:             "transfer.zip",
			WatcherName:     watcherName,
			RetentionPeriod: &retentionPeriod,
			AutoApproveAIP:  false,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowResult(nil))
}

func (s *ProcessingWorkflowTestSuite) TestAutoApprovedAIP() {
	cfg := config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Storage: storage.Config{
			DefaultPermanentLocationID: locationID,
		},
	}
	s.SetupWorkflowTest(cfg)

	pkgID := uint(1)
	watcherName := "watcher"
	key := "transfer.zip"
	retentionPeriod := 1 * time.Second
	ctx := mock.AnythingOfType("*context.valueCtx")
	sessionCtx := mock.AnythingOfType("*context.timerCtx")
	pkgsvc := s.workflow.pkgsvc
	rng := rand.New(rand.NewSource(1)) // #nosec: G404

	// Activity mocks/assertions sequence
	s.env.OnActivity(
		createPackageLocalActivity,
		ctx,
		pkgsvc,
		&createPackageLocalActivityParams{Key: key, Status: enums.PackageStatusQueued},
	).Return(pkgID, nil).Once()
	s.env.OnActivity(
		setStatusInProgressLocalActivity,
		ctx,
		pkgsvc,
		pkgID,
		mock.AnythingOfType("time.Time"),
	).Return(nil, nil)

	s.env.OnActivity(
		createPreservationActionLocalActivity,
		ctx,
		pkgsvc,
		&createPreservationActionLocalActivityParams{
			WorkflowID: "default-test-workflow-id",
			Type:       enums.PreservationActionTypeCreateAip,
			Status:     enums.PreservationActionStatusInProgress,
			StartedAt:  startTime,
			PackageID:  1,
		},
	).Return(uint(0), nil)

	s.env.OnActivity(activities.DownloadActivityName, sessionCtx,
		&activities.DownloadActivityParams{Key: key, WatcherName: watcherName},
	).Return(
		&activities.DownloadActivityResult{Path: tempPath + "/" + key}, nil,
	)

	s.env.OnActivity(archiveextract.Name, sessionCtx,
		&archiveextract.Params{SourcePath: tempPath + "/" + key},
	).Return(
		&archiveextract.Result{ExtractPath: extractPath}, nil,
	)

	s.env.OnActivity(
		activities.ClassifyPackageActivityName,
		sessionCtx,
		activities.ClassifyPackageActivityParams{Path: extractPath},
	).Return(
		&activities.ClassifyPackageActivityResult{Type: enums.PackageTypeBagIt}, nil,
	)

	s.env.OnActivity(
		createPreservationTaskLocalActivity,
		ctx,
		&createPreservationTaskLocalActivityParams{
			PkgSvc: pkgsvc,
			RNG:    rng,
			PreservationTask: datatypes.PreservationTask{
				Name:   "Validate Bag",
				Status: enums.PreservationTaskStatusInProgress,
				StartedAt: sql.NullTime{
					Time:  startTime,
					Valid: true,
				},
				PreservationActionID: uint(0),
			},
		},
	).Return(uint(101), nil)

	s.env.OnActivity(
		bagvalidate.Name,
		sessionCtx,
		&bagvalidate.Params{Path: extractPath},
	).Return(
		&bagvalidate.Result{Valid: true},
		nil,
	)

	s.env.OnActivity(
		completePreservationTaskLocalActivity,
		ctx,
		pkgsvc,
		&completePreservationTaskLocalActivityParams{
			ID:          uint(101),
			Status:      enums.PreservationTaskStatusDone,
			CompletedAt: startTime,
			Note:        ref.New("Bag is valid"),
		},
	).Return(&completePreservationTaskLocalActivityResult{}, nil)

	s.env.OnActivity(activities.BundleActivityName, sessionCtx,
		&activities.BundleActivityParams{
			SourcePath:  extractPath,
			TransferDir: s.transferDir,
			IsDir:       true,
		},
	).Return(
		&activities.BundleActivityResult{FullPath: transferPath},
		nil,
	)

	s.env.OnActivity(a3m.CreateAIPActivityName, sessionCtx, mock.AnythingOfType("*a3m.CreateAIPActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(updatePackageLocalActivity, ctx, pkgsvc, mock.AnythingOfType("*workflow.updatePackageLocalActivityParams")).
		Return(nil, nil).
		Times(2)

	s.env.OnActivity(
		createPreservationTaskLocalActivity,
		ctx,
		&createPreservationTaskLocalActivityParams{
			PkgSvc: pkgsvc,
			RNG:    rng,
			PreservationTask: datatypes.PreservationTask{
				Name:                 "Move AIP",
				Status:               enums.PreservationTaskStatusInProgress,
				StartedAt:            sql.NullTime{Time: startTime, Valid: true},
				Note:                 "Moving to permanent storage",
				PreservationActionID: uint(0),
			},
		},
	).Return(uint(102), nil)

	s.env.OnActivity(activities.UploadActivityName, sessionCtx, mock.AnythingOfType("*activities.UploadActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil).
		Never()
	s.env.OnActivity(setPreservationActionStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil).
		Never()

	s.env.OnActivity(
		completePreservationTaskLocalActivity,
		ctx,
		pkgsvc,
		&completePreservationTaskLocalActivityParams{
			ID:          uint(102),
			Status:      enums.PreservationTaskStatusDone,
			CompletedAt: startTime,
			Note:        ref.New("Moved to location f2cc963f-c14d-4eaa-b950-bd207189a1f1"),
		},
	).Return(&completePreservationTaskLocalActivityResult{}, nil)

	s.env.OnActivity(activities.MoveToPermanentStorageActivityName, sessionCtx, mock.AnythingOfType("*activities.MoveToPermanentStorageActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(activities.PollMoveToPermanentStorageActivityName, sessionCtx, mock.AnythingOfType("*activities.PollMoveToPermanentStorageActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(setLocationIDLocalActivity, ctx, pkgsvc, pkgID, locationID).Return(nil, nil).Once()
	s.env.OnActivity(completePreservationActionLocalActivity, ctx, pkgsvc, mock.AnythingOfType("*workflow.completePreservationActionLocalActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(
		removepaths.Name,
		sessionCtx,
		&removepaths.Params{Paths: []string{tempPath, transferPath}},
	).Return(&removepaths.Result{}, nil)
	s.env.OnActivity(activities.DeleteOriginalActivityName, sessionCtx, watcherName, key).Return(nil, nil).Once()

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&package_.ProcessingWorkflowRequest{
			Key:                        key,
			WatcherName:                watcherName,
			RetentionPeriod:            &retentionPeriod,
			AutoApproveAIP:             true,
			DefaultPermanentLocationID: &cfg.Storage.DefaultPermanentLocationID,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowResult(nil))
}

func (s *ProcessingWorkflowTestSuite) TestAMWorkflow() {
	pkgID := uint(1)
	watcherName := "watcher"
	key := "transfer.zip"
	retentionPeriod := 1 * time.Second
	ctx := mock.AnythingOfType("*context.valueCtx")
	sessionCtx := mock.AnythingOfType("*context.timerCtx")

	cfg := config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.AmWorkerTaskQueue},
		Storage:      storage.Config{DefaultPermanentLocationID: amssLocationID},
	}
	s.SetupWorkflowTest(cfg)

	pkgsvc := s.workflow.pkgsvc

	// Activity mocks/assertions sequence
	s.env.OnActivity(createPackageLocalActivity, ctx,
		pkgsvc,
		&createPackageLocalActivityParams{Key: key, Status: enums.PackageStatusQueued},
	).Return(pkgID, nil)

	s.env.OnActivity(setStatusInProgressLocalActivity, ctx, pkgsvc, pkgID, mock.AnythingOfType("time.Time")).
		Return(nil, nil)

	s.env.OnActivity(createPreservationActionLocalActivity, ctx,
		pkgsvc, mock.AnythingOfType("*workflow.createPreservationActionLocalActivityParams"),
	).Return(uint(0), nil)

	s.env.OnActivity(activities.DownloadActivityName, sessionCtx,
		&activities.DownloadActivityParams{Key: key, WatcherName: watcherName},
	).Return(
		&activities.DownloadActivityResult{Path: tempPath + "/" + key}, nil,
	)

	s.env.OnActivity(archiveextract.Name, sessionCtx,
		&archiveextract.Params{SourcePath: tempPath + "/" + key},
	).Return(
		&archiveextract.Result{ExtractPath: extractPath}, nil,
	)

	s.env.OnActivity(
		activities.ClassifyPackageActivityName,
		sessionCtx,
		activities.ClassifyPackageActivityParams{Path: extractPath},
	).Return(
		&activities.ClassifyPackageActivityResult{Type: enums.PackageTypeUnknown}, nil,
	)

	// Archivematica specific activities.
	s.env.OnActivity(bagcreate.Name, sessionCtx,
		&bagcreate.Params{SourcePath: extractPath},
	).Return(
		&bagcreate.Result{BagPath: extractPath}, nil,
	)

	s.env.OnActivity(archivezip.Name, sessionCtx,
		&archivezip.Params{SourceDir: extractPath},
	).Return(
		&archivezip.Result{Path: extractPath + "/transfer.zip"}, nil,
	)

	s.env.OnActivity(am.UploadTransferActivityName, sessionCtx,
		&am.UploadTransferActivityParams{SourcePath: extractPath + "/transfer.zip"},
	).Return(
		&am.UploadTransferActivityResult{
			RemoteFullPath:     "transfer.zip",
			RemoteRelativePath: "transfer.zip",
		}, nil,
	)

	s.env.OnActivity(am.StartTransferActivityName, sessionCtx,
		&am.StartTransferActivityParams{Name: key, RelativePath: "transfer.zip"},
	).Return(
		&am.StartTransferActivityResult{TransferID: transferID.String()}, nil,
	)

	s.env.OnActivity(am.PollTransferActivityName, sessionCtx,
		&am.PollTransferActivityParams{TransferID: transferID.String()},
	).Return(
		&am.PollTransferActivityResult{SIPID: sipID.String()}, nil,
	)

	s.env.OnActivity(am.PollIngestActivityName, sessionCtx,
		&am.PollIngestActivityParams{SIPID: sipID.String()},
	).Return(
		&am.PollIngestActivityResult{Status: "COMPLETE"}, nil,
	)

	s.env.OnActivity(setLocationIDLocalActivity, ctx,
		pkgsvc,
		pkgID,
		amssLocationID,
	).Return(&setLocationIDLocalActivityResult{}, nil)

	s.env.OnActivity(activities.CreateStoragePackageActivityName, sessionCtx,
		&activities.CreateStoragePackageActivityParams{
			Name:       key,
			AIPID:      sipID.String(),
			ObjectKey:  sipID.String(),
			Status:     "stored",
			LocationID: &amssLocationID,
		},
	).Return(&activities.CreateStoragePackageActivityResult{}, nil)

	s.env.OnActivity(am.DeleteTransferActivityName, sessionCtx,
		&am.DeleteTransferActivityParams{Destination: "transfer.zip"},
	).Return(nil, nil)

	// Post-preservation activities.
	s.env.OnActivity(
		updatePackageLocalActivity,
		ctx,
		pkgsvc,
		mock.AnythingOfType("*workflow.updatePackageLocalActivityParams"),
	).Return(nil, nil)
	s.env.OnActivity(
		completePreservationActionLocalActivity,
		ctx,
		pkgsvc,
		mock.AnythingOfType("*workflow.completePreservationActionLocalActivityParams"),
	).Return(nil, nil)
	s.env.OnActivity(
		removepaths.Name,
		sessionCtx,
		&removepaths.Params{Paths: []string{tempPath}},
	).Return(&removepaths.Result{}, nil)
	s.env.OnActivity(
		activities.DeleteOriginalActivityName,
		sessionCtx,
		watcherName,
		key,
	).Return(nil, nil)

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&package_.ProcessingWorkflowRequest{
			WatcherName:                watcherName,
			RetentionPeriod:            &retentionPeriod,
			AutoApproveAIP:             true,
			DefaultPermanentLocationID: &cfg.Storage.DefaultPermanentLocationID,
			Key:                        key,
			TransferDeadline:           time.Second,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowResult(nil))
}

func (s *ProcessingWorkflowTestSuite) TestPackageRejection() {
	cfg := config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Storage: storage.Config{
			DefaultPermanentLocationID: locationID,
		},
	}
	s.SetupWorkflowTest(cfg)

	pkgID := uint(1)
	key := "transfer.zip"
	watcherName := "watcher"
	retentionPeriod := 1 * time.Second
	sessionCtx := mock.AnythingOfType("*context.timerCtx")

	// Signal handler that mimics package rejection
	s.env.RegisterDelayedCallback(
		func() {
			s.env.SignalWorkflow(
				package_.ReviewPerformedSignalName,
				package_.ReviewPerformedSignal{Accepted: false},
			)
		},
		0,
	)

	// Activity mocks/assertions sequence
	s.env.OnActivity(createPackageLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(pkgID, nil)
	s.env.OnActivity(setStatusInProgressLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(createPreservationActionLocalActivity, mock.Anything, mock.Anything, mock.Anything).
		Return(uint(0), nil)

	s.env.OnActivity(activities.DownloadActivityName, sessionCtx,
		&activities.DownloadActivityParams{Key: key, WatcherName: watcherName},
	).Return(
		&activities.DownloadActivityResult{Path: tempPath + "/" + key}, nil,
	)

	s.env.OnActivity(archiveextract.Name, sessionCtx,
		&archiveextract.Params{SourcePath: tempPath + "/" + key},
	).Return(
		&archiveextract.Result{ExtractPath: extractPath}, nil,
	)

	s.env.OnActivity(
		activities.ClassifyPackageActivityName,
		sessionCtx,
		activities.ClassifyPackageActivityParams{Path: extractPath},
	).Return(
		&activities.ClassifyPackageActivityResult{Type: enums.PackageTypeUnknown}, nil,
	)

	s.env.OnActivity(activities.BundleActivityName, sessionCtx,
		&activities.BundleActivityParams{
			SourcePath:  extractPath,
			TransferDir: s.transferDir,
			IsDir:       true,
		},
	).Return(
		&activities.BundleActivityResult{FullPath: transferPath},
		nil,
	)

	s.env.OnActivity(a3m.CreateAIPActivityName, mock.Anything, mock.Anything).Return(nil, nil)
	s.env.OnActivity(updatePackageLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(completePreservationTaskLocalActivity, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(activities.UploadActivityName, mock.Anything, mock.Anything).Return(nil, nil)
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(setPreservationActionStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(createPreservationTaskLocalActivity, mock.Anything, mock.Anything, mock.Anything).
		Return(uint(0), nil)
	s.env.OnActivity(activities.RejectPackageActivityName, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(
		removepaths.Name,
		sessionCtx,
		&removepaths.Params{Paths: []string{tempPath, transferPath}},
	).Return(&removepaths.Result{}, nil)
	s.env.OnActivity(activities.DeleteOriginalActivityName, sessionCtx, watcherName, key).Return(nil, nil).Once()
	s.env.OnActivity(updatePackageLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(completePreservationActionLocalActivity, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&package_.ProcessingWorkflowRequest{
			Key:             "transfer.zip",
			WatcherName:     watcherName,
			RetentionPeriod: &retentionPeriod,
			AutoApproveAIP:  false,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowResult(nil))
}

func (s *ProcessingWorkflowTestSuite) TestPreprocessingChildWorkflow() {
	cfg := config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Preprocessing: preprocessing.Config{
			Enabled:    true,
			Extract:    true,
			SharedPath: "/home/enduro/preprocessing/",
			Temporal: preprocessing.Temporal{
				Namespace:    "default",
				TaskQueue:    "preprocessing",
				WorkflowName: "preprocessing",
			},
		},
		Storage: storage.Config{
			DefaultPermanentLocationID: locationID,
		},
	}
	s.SetupWorkflowTest(cfg)

	pkgID := uint(1)
	watcherName := "watcher"
	key := "transfer.zip"
	retentionPeriod := 1 * time.Second
	ctx := mock.AnythingOfType("*context.valueCtx")
	sessionCtx := mock.AnythingOfType("*context.timerCtx")
	pkgsvc := s.workflow.pkgsvc

	downloadDir := strings.Replace(tempPath, "/tmp/", cfg.Preprocessing.SharedPath, 1)
	prepDest := strings.Replace(extractPath, "/tmp/", cfg.Preprocessing.SharedPath, 1)

	// Activity mocks/assertions sequence
	s.env.OnActivity(
		createPackageLocalActivity,
		ctx,
		pkgsvc,
		&createPackageLocalActivityParams{Key: key, Status: enums.PackageStatusQueued},
	).Return(pkgID, nil)

	s.env.OnActivity(
		setStatusInProgressLocalActivity,
		ctx,
		pkgsvc,
		pkgID,
		startTime,
	).Return(nil, nil)

	s.env.OnActivity(
		createPreservationActionLocalActivity,
		ctx,
		pkgsvc,
		&createPreservationActionLocalActivityParams{
			WorkflowID: "default-test-workflow-id",
			Type:       enums.PreservationActionTypeCreateAip,
			Status:     enums.PreservationActionStatusInProgress,
			StartedAt:  startTime,
			PackageID:  1,
		},
	).Return(uint(1), nil)

	s.env.OnActivity(activities.DownloadActivityName, sessionCtx,
		&activities.DownloadActivityParams{
			Key:             key,
			WatcherName:     watcherName,
			DestinationPath: cfg.Preprocessing.SharedPath,
		},
	).Return(
		&activities.DownloadActivityResult{Path: downloadDir + "/" + key}, nil,
	)

	s.env.OnWorkflow(
		preprocessingChildWorkflow,
		mock.AnythingOfType("*internal.valueCtx"),
		&preprocessing.WorkflowParams{
			RelativePath: strings.TrimPrefix(downloadDir+"/"+key, cfg.Preprocessing.SharedPath),
		},
	).Return(
		&preprocessing.WorkflowResult{
			Outcome:      preprocessing.OutcomeSuccess,
			RelativePath: strings.TrimPrefix(prepDest, cfg.Preprocessing.SharedPath),
			PreservationTasks: []preprocessing.Task{
				{
					Name:        "Identify SIP structure",
					Message:     "SIP structure identified: VecteurAIP",
					Outcome:     enums.PreprocessingTaskOutcomeSuccess,
					StartedAt:   time.Date(2024, 6, 14, 10, 5, 32, 0, time.UTC),
					CompletedAt: time.Date(2024, 6, 14, 10, 5, 33, 0, time.UTC),
				},
			},
		},
		nil,
	)

	s.env.OnActivity(
		localact.SavePreprocessingTasksActivity,
		ctx,
		localact.SavePreprocessingTasksActivityParams{
			PkgSvc:               pkgsvc,
			RNG:                  rand.New(rand.NewSource(1)), // #nosec: G404
			PreservationActionID: 1,
			Tasks: []preprocessing.Task{
				{
					Name:        "Identify SIP structure",
					Message:     "SIP structure identified: VecteurAIP",
					Outcome:     enums.PreprocessingTaskOutcomeSuccess,
					StartedAt:   time.Date(2024, 6, 14, 10, 5, 32, 0, time.UTC),
					CompletedAt: time.Date(2024, 6, 14, 10, 5, 33, 0, time.UTC),
				},
			},
		},
	).Return(
		&localact.SavePreprocessingTasksActivityResult{Count: 1},
		nil,
	)

	s.env.OnActivity(
		activities.ClassifyPackageActivityName,
		sessionCtx,
		activities.ClassifyPackageActivityParams{Path: prepDest},
	).Return(
		&activities.ClassifyPackageActivityResult{Type: enums.PackageTypeBagIt}, nil,
	)

	s.env.OnActivity(
		createPreservationTaskLocalActivity,
		ctx,
		&createPreservationTaskLocalActivityParams{
			PkgSvc: pkgsvc,
			RNG:    s.workflow.rng,
			PreservationTask: datatypes.PreservationTask{
				Name:   "Validate Bag",
				Status: enums.PreservationTaskStatusInProgress,
				StartedAt: sql.NullTime{
					Time:  startTime,
					Valid: true,
				},
				PreservationActionID: uint(1),
			},
		},
	).Return(uint(101), nil)

	s.env.OnActivity(
		bagvalidate.Name,
		sessionCtx,
		&bagvalidate.Params{Path: prepDest},
	).Return(
		&bagvalidate.Result{Valid: true},
		nil,
	)

	s.env.OnActivity(
		completePreservationTaskLocalActivity,
		ctx,
		pkgsvc,
		&completePreservationTaskLocalActivityParams{
			ID:          uint(101),
			Status:      enums.PreservationTaskStatusDone,
			CompletedAt: startTime,
			Note:        ref.New("Bag is valid"),
		},
	).Return(&completePreservationTaskLocalActivityResult{}, nil)

	s.env.OnActivity(activities.BundleActivityName, sessionCtx,
		&activities.BundleActivityParams{
			SourcePath:  prepDest,
			TransferDir: s.transferDir,
			IsDir:       true,
		},
	).Return(
		&activities.BundleActivityResult{FullPath: transferPath},
		nil,
	)

	s.env.OnActivity(a3m.CreateAIPActivityName, sessionCtx, mock.AnythingOfType("*a3m.CreateAIPActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(updatePackageLocalActivity, ctx, pkgsvc, mock.AnythingOfType("*workflow.updatePackageLocalActivityParams")).
		Return(nil, nil).
		Times(2)

	s.env.OnActivity(
		createPreservationTaskLocalActivity,
		ctx,
		&createPreservationTaskLocalActivityParams{
			PkgSvc: pkgsvc,
			RNG:    s.workflow.rng,
			PreservationTask: datatypes.PreservationTask{
				Name:                 "Move AIP",
				Status:               enums.PreservationTaskStatusInProgress,
				StartedAt:            sql.NullTime{Time: startTime, Valid: true},
				PreservationActionID: uint(1),
				Note:                 "Moving to permanent storage",
			},
		},
	).Return(uint(102), nil)

	s.env.OnActivity(
		activities.UploadActivityName,
		sessionCtx,
		mock.AnythingOfType("*activities.UploadActivityParams"),
	).Return(nil, nil)

	s.env.OnActivity(
		completePreservationTaskLocalActivity,
		ctx,
		pkgsvc,
		&completePreservationTaskLocalActivityParams{
			ID:          uint(102),
			Status:      enums.PreservationTaskStatusDone,
			CompletedAt: startTime,
			Note:        ref.New("Moved to location f2cc963f-c14d-4eaa-b950-bd207189a1f1"),
		},
	).Return(&completePreservationTaskLocalActivityResult{}, nil)

	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil).
		Never()
	s.env.OnActivity(setPreservationActionStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil).
		Never()

	s.env.OnActivity(activities.MoveToPermanentStorageActivityName, sessionCtx, mock.AnythingOfType("*activities.MoveToPermanentStorageActivityParams")).
		Return(nil, nil).
		Once()

	s.env.OnActivity(activities.PollMoveToPermanentStorageActivityName, sessionCtx, mock.AnythingOfType("*activities.PollMoveToPermanentStorageActivityParams")).
		Return(nil, nil).
		Once()

	s.env.OnActivity(setLocationIDLocalActivity, ctx, pkgsvc, pkgID, locationID).Return(nil, nil).Once()

	s.env.OnActivity(completePreservationActionLocalActivity, ctx, pkgsvc, mock.AnythingOfType("*workflow.completePreservationActionLocalActivityParams")).
		Return(nil, nil).
		Once()

	s.env.OnActivity(
		removepaths.Name,
		sessionCtx,
		&removepaths.Params{Paths: []string{downloadDir, transferPath}},
	).Return(&removepaths.Result{}, nil)

	s.env.OnActivity(activities.DeleteOriginalActivityName, sessionCtx, watcherName, key).Return(nil, nil).Once()

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&package_.ProcessingWorkflowRequest{
			Key:                        key,
			WatcherName:                watcherName,
			RetentionPeriod:            &retentionPeriod,
			AutoApproveAIP:             true,
			DefaultPermanentLocationID: &cfg.Storage.DefaultPermanentLocationID,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowResult(nil))
}
