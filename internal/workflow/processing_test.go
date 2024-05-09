package workflow

import (
	"strings"
	"testing"
	"time"

	"github.com/artefactual-sdps/temporal-activities/archive"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.artefactual.dev/amclient/amclienttest"
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
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/package_"
	packagefake "github.com/artefactual-sdps/enduro/internal/package_/fake"
	"github.com/artefactual-sdps/enduro/internal/preprocessing"
	"github.com/artefactual-sdps/enduro/internal/pres"
	sftp_fake "github.com/artefactual-sdps/enduro/internal/sftp/fake"
	"github.com/artefactual-sdps/enduro/internal/temporal"
	watcherfake "github.com/artefactual-sdps/enduro/internal/watcher/fake"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

const (
	tempPath     string = "/tmp/enduro123456"
	extractPath  string = "/tmp/enduro123456/extract"
	transferPath string = "/tmp/2098266580-enduro-transfer/enduro4162369760/transfer"
)

var (
	locationID = uuid.MustParse("a06a155c-9cf0-4416-a2b6-e90e58ef3186")
	transferID = uuid.MustParse("65233405-771e-4f7e-b2d9-b08439570ba2")
	sipID      = uuid.MustParse("9e8161cc-2815-4d6f-8a75-f003c41b257b")
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

	ctrl := gomock.NewController(s.T())
	logger := logr.Discard()
	a3mTransferServiceClient := a3mfake.NewMockTransferServiceClient(ctrl)
	pkgsvc := packagefake.NewMockService(ctrl)
	wsvc := watcherfake.NewMockService(ctrl)

	s.env.RegisterActivityWithOptions(
		activities.NewDownloadActivity(logger, noop.Tracer{}, wsvc).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.DownloadActivityName},
	)
	s.env.RegisterActivityWithOptions(
		archive.NewExtractActivity(cfg.ExtractActivity).Execute,
		temporalsdk_activity.RegisterOptions{Name: archive.ExtractActivityName},
	)
	s.env.RegisterActivityWithOptions(
		activities.NewBundleActivity(logger).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.BundleActivityName},
	)
	s.env.RegisterActivityWithOptions(
		a3m.NewCreateAIPActivity(logger, noop.Tracer{}, a3mTransferServiceClient, &a3m.Config{}, pkgsvc).Execute,
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
	s.env.RegisterActivityWithOptions(
		activities.NewCleanUpActivity().Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.CleanUpActivityName},
	)
	s.env.RegisterActivityWithOptions(
		activities.NewDeleteOriginalActivity(wsvc).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.DeleteOriginalActivityName},
	)

	// Set up AM taskqueue.
	if cfg.Preservation.TaskQueue == temporal.AmWorkerTaskQueue {
		s.setupAMWorkflowTest(logger, &cfg.AM, ctrl, pkgsvc)
	}

	s.env.RegisterWorkflowWithOptions(
		preprocessingChildWorkflow,
		temporalsdk_workflow.RegisterOptions{Name: "preprocessing"},
	)

	s.workflow = NewProcessingWorkflow(logger, cfg, pkgsvc, wsvc)
}

func (s *ProcessingWorkflowTestSuite) setupAMWorkflowTest(
	logger logr.Logger,
	cfg *am.Config,
	ctrl *gomock.Controller,
	pkgsvc package_.Service,
) {
	clock := clockwork.NewFakeClock()
	sftpc := sftp_fake.NewMockClient(ctrl)

	s.env.RegisterActivityWithOptions(
		activities.NewZipActivity(logger).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.ZipActivityName},
	)
	s.env.RegisterActivityWithOptions(
		am.NewUploadTransferActivity(logger, sftpc, 10*time.Second).Execute,
		temporalsdk_activity.RegisterOptions{Name: am.UploadTransferActivityName},
	)
	s.env.RegisterActivityWithOptions(
		am.NewDeleteTransferActivity(logger, sftpc).Execute,
		temporalsdk_activity.RegisterOptions{
			Name: am.DeleteTransferActivityName,
		},
	)
	s.env.RegisterActivityWithOptions(
		am.NewStartTransferActivity(logger, &am.Config{}, amclienttest.NewMockPackageService(ctrl)).Execute,
		temporalsdk_activity.RegisterOptions{Name: am.StartTransferActivityName},
	)
	s.env.RegisterActivityWithOptions(
		am.NewPollTransferActivity(
			logger,
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
			logger,
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
	s.env.OnActivity(createPreservationActionLocalActivity, mock.Anything, mock.Anything, mock.Anything).
		Return(uint(0), nil)

	s.env.OnActivity(activities.DownloadActivityName, sessionCtx,
		&activities.DownloadActivityParams{Key: key, WatcherName: watcherName},
	).Return(
		&activities.DownloadActivityResult{Path: tempPath + "/" + key}, nil,
	)

	s.env.OnActivity(archive.ExtractActivityName, sessionCtx,
		&archive.ExtractActivityParams{SourcePath: tempPath + "/" + key},
	).Return(
		&archive.ExtractActivityResult{ExtractPath: extractPath}, nil,
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
	s.env.OnActivity(activities.CleanUpActivityName, sessionCtx, mock.AnythingOfType("*activities.CleanUpActivityParams")).
		Return(nil, nil).
		Once()
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
	}
	s.SetupWorkflowTest(cfg)

	pkgID := uint(1)
	watcherName := "watcher"
	key := "transfer.zip"
	retentionPeriod := 1 * time.Second
	ctx := mock.AnythingOfType("*context.valueCtx")
	sessionCtx := mock.AnythingOfType("*context.timerCtx")
	logger := s.workflow.logger
	pkgsvc := s.workflow.pkgsvc

	// Activity mocks/assertions sequence
	s.env.OnActivity(
		createPackageLocalActivity,
		ctx,
		logger,
		pkgsvc,
		&createPackageLocalActivityParams{Key: key, Status: enums.PackageStatusQueued},
	).Return(pkgID, nil).Once()
	s.env.OnActivity(setStatusInProgressLocalActivity, ctx, pkgsvc, pkgID, mock.AnythingOfType("time.Time")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(createPreservationActionLocalActivity, ctx, pkgsvc, mock.AnythingOfType("*workflow.createPreservationActionLocalActivityParams")).
		Return(uint(0), nil).
		Once()

	s.env.OnActivity(activities.DownloadActivityName, sessionCtx,
		&activities.DownloadActivityParams{Key: key, WatcherName: watcherName},
	).Return(
		&activities.DownloadActivityResult{Path: tempPath + "/" + key}, nil,
	)

	s.env.OnActivity(archive.ExtractActivityName, sessionCtx,
		&archive.ExtractActivityParams{SourcePath: tempPath + "/" + key},
	).Return(
		&archive.ExtractActivityResult{ExtractPath: extractPath}, nil,
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

	s.env.OnActivity(a3m.CreateAIPActivityName, sessionCtx, mock.AnythingOfType("*a3m.CreateAIPActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(updatePackageLocalActivity, ctx, logger, pkgsvc, mock.AnythingOfType("*workflow.updatePackageLocalActivityParams")).
		Return(nil, nil).
		Times(2)
	s.env.OnActivity(createPreservationTaskLocalActivity, ctx, pkgsvc, mock.AnythingOfType("*workflow.createPreservationTaskLocalActivityParams")).
		Return(uint(0), nil).
		Once()
	s.env.OnActivity(activities.UploadActivityName, sessionCtx, mock.AnythingOfType("*activities.UploadActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil).
		Never()
	s.env.OnActivity(setPreservationActionStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil).
		Never()
	s.env.OnActivity(completePreservationTaskLocalActivity, ctx, pkgsvc, mock.AnythingOfType("*workflow.completePreservationTaskLocalActivityParams")).
		Return(nil, nil).
		Once()
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
	s.env.OnActivity(activities.CleanUpActivityName, sessionCtx, mock.AnythingOfType("*activities.CleanUpActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(activities.DeleteOriginalActivityName, sessionCtx, watcherName, key).Return(nil, nil).Once()

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&package_.ProcessingWorkflowRequest{
			Key:                        key,
			WatcherName:                watcherName,
			RetentionPeriod:            &retentionPeriod,
			AutoApproveAIP:             true,
			DefaultPermanentLocationID: &locationID,
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
		AM:           am.Config{AMSSLocationID: "cf3059dd-4565-4fe9-92fe-b16d1a777403"},
		Preservation: pres.Config{TaskQueue: temporal.AmWorkerTaskQueue},
	}
	s.SetupWorkflowTest(cfg)

	logger := s.workflow.logger
	pkgsvc := s.workflow.pkgsvc

	// Activity mocks/assertions sequence
	s.env.OnActivity(createPackageLocalActivity, ctx,
		logger,
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

	s.env.OnActivity(archive.ExtractActivityName, sessionCtx,
		&archive.ExtractActivityParams{SourcePath: tempPath + "/" + key},
	).Return(
		&archive.ExtractActivityResult{ExtractPath: extractPath}, nil,
	)

	s.env.OnActivity(
		activities.BundleActivityName,
		sessionCtx,
		&activities.BundleActivityParams{
			SourcePath: extractPath,
			IsDir:      true,
		},
	).Return(
		&activities.BundleActivityResult{
			FullPath: transferPath,
		},
		nil,
	)

	// Archivematica specific activities.
	s.env.OnActivity(activities.ZipActivityName, sessionCtx,
		&activities.ZipActivityParams{SourceDir: transferPath},
	).Return(
		&activities.ZipActivityResult{Path: transferPath + "/transfer.zip"}, nil,
	)

	s.env.OnActivity(am.UploadTransferActivityName, sessionCtx,
		&am.UploadTransferActivityParams{SourcePath: transferPath + "/transfer.zip"},
	).Return(
		&am.UploadTransferActivityResult{
			RemoteFullPath:     "transfer.zip",
			RemoteRelativePath: "transfer.zip",
		}, nil,
	)

	s.env.OnActivity(am.StartTransferActivityName, sessionCtx,
		&am.StartTransferActivityParams{Name: key, Path: "transfer.zip"},
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
		uuid.MustParse(cfg.AM.AMSSLocationID),
	).Return(&setLocationIDLocalActivityResult{}, nil)

	s.env.OnActivity(activities.CreateStoragePackageActivityName, sessionCtx,
		&activities.CreateStoragePackageActivityParams{
			Name:       key,
			AIPID:      sipID.String(),
			ObjectKey:  sipID.String(),
			Status:     "stored",
			LocationID: cfg.AM.AMSSLocationID,
		},
	).Return(&activities.CreateStoragePackageActivityResult{}, nil)

	s.env.OnActivity(am.DeleteTransferActivityName, sessionCtx,
		&am.DeleteTransferActivityParams{Destination: "transfer.zip"},
	).Return(nil, nil)

	// Post-preservation activities.
	s.env.OnActivity(updatePackageLocalActivity, ctx, logger, pkgsvc, mock.AnythingOfType("*workflow.updatePackageLocalActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil).
		Never()
	s.env.OnActivity(completePreservationActionLocalActivity, ctx, pkgsvc, mock.AnythingOfType("*workflow.completePreservationActionLocalActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(activities.CleanUpActivityName, sessionCtx, mock.AnythingOfType("*activities.CleanUpActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(activities.DeleteOriginalActivityName, sessionCtx, watcherName, key).Return(nil, nil).Once()

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&package_.ProcessingWorkflowRequest{
			WatcherName:                watcherName,
			RetentionPeriod:            &retentionPeriod,
			AutoApproveAIP:             true,
			DefaultPermanentLocationID: &locationID,
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

	s.env.OnActivity(archive.ExtractActivityName, sessionCtx,
		&archive.ExtractActivityParams{SourcePath: tempPath + "/" + key},
	).Return(
		&archive.ExtractActivityResult{ExtractPath: extractPath}, nil,
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
	s.env.OnActivity(activities.CleanUpActivityName, sessionCtx, mock.AnythingOfType("*activities.CleanUpActivityParams")).
		Return(nil, nil).
		Once()
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
	}
	s.SetupWorkflowTest(cfg)

	pkgID := uint(1)
	watcherName := "watcher"
	key := "transfer.zip"
	retentionPeriod := 1 * time.Second
	ctx := mock.AnythingOfType("*context.valueCtx")
	sessionCtx := mock.AnythingOfType("*context.timerCtx")
	logger := s.workflow.logger
	pkgsvc := s.workflow.pkgsvc

	// Activity mocks/assertions sequence
	s.env.OnActivity(
		createPackageLocalActivity,
		ctx,
		logger,
		pkgsvc,
		&createPackageLocalActivityParams{Key: key, Status: enums.PackageStatusQueued},
	).Return(pkgID, nil).Once()
	s.env.OnActivity(setStatusInProgressLocalActivity, ctx, pkgsvc, pkgID, mock.AnythingOfType("time.Time")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(createPreservationActionLocalActivity, ctx, pkgsvc, mock.AnythingOfType("*workflow.createPreservationActionLocalActivityParams")).
		Return(uint(0), nil).
		Once()

	downloadDest := strings.Replace(tempPath, "/tmp/", cfg.Preprocessing.SharedPath, 1) + "/" + key
	s.env.OnActivity(activities.DownloadActivityName, sessionCtx,
		&activities.DownloadActivityParams{
			Key:             key,
			WatcherName:     watcherName,
			DestinationPath: cfg.Preprocessing.SharedPath,
		},
	).Return(
		&activities.DownloadActivityResult{Path: downloadDest}, nil,
	)

	prepDest := strings.Replace(extractPath, "/tmp/", cfg.Preprocessing.SharedPath, 1)
	s.env.OnWorkflow(
		preprocessingChildWorkflow,
		mock.Anything,
		&preprocessing.WorkflowParams{
			RelativePath: strings.TrimPrefix(downloadDest, cfg.Preprocessing.SharedPath),
		},
	).Return(
		&preprocessing.WorkflowResult{
			RelativePath: strings.TrimPrefix(prepDest, cfg.Preprocessing.SharedPath),
		},
		nil,
	)

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
	s.env.OnActivity(updatePackageLocalActivity, ctx, logger, pkgsvc, mock.AnythingOfType("*workflow.updatePackageLocalActivityParams")).
		Return(nil, nil).
		Times(2)
	s.env.OnActivity(createPreservationTaskLocalActivity, ctx, pkgsvc, mock.AnythingOfType("*workflow.createPreservationTaskLocalActivityParams")).
		Return(uint(0), nil).
		Once()
	s.env.OnActivity(activities.UploadActivityName, sessionCtx, mock.AnythingOfType("*activities.UploadActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil).
		Never()
	s.env.OnActivity(setPreservationActionStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil).
		Never()
	s.env.OnActivity(completePreservationTaskLocalActivity, ctx, pkgsvc, mock.AnythingOfType("*workflow.completePreservationTaskLocalActivityParams")).
		Return(nil, nil).
		Once()
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
	s.env.OnActivity(activities.CleanUpActivityName, sessionCtx, mock.AnythingOfType("*activities.CleanUpActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(activities.DeleteOriginalActivityName, sessionCtx, watcherName, key).Return(nil, nil).Once()

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&package_.ProcessingWorkflowRequest{
			Key:                        key,
			WatcherName:                watcherName,
			RetentionPeriod:            &retentionPeriod,
			AutoApproveAIP:             true,
			DefaultPermanentLocationID: &locationID,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowResult(nil))
}
