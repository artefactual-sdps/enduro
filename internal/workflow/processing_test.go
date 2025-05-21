package workflow

import (
	"database/sql"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/artefactual-sdps/temporal-activities/archiveextract"
	"github.com/artefactual-sdps/temporal-activities/archivezip"
	"github.com/artefactual-sdps/temporal-activities/bagcreate"
	"github.com/artefactual-sdps/temporal-activities/bagvalidate"
	"github.com/artefactual-sdps/temporal-activities/bucketupload"
	"github.com/artefactual-sdps/temporal-activities/removepaths"
	"github.com/artefactual-sdps/temporal-activities/xmlvalidate"
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
	"gocloud.dev/blob/memblob"

	"github.com/artefactual-sdps/enduro/internal/a3m"
	a3mfake "github.com/artefactual-sdps/enduro/internal/a3m/fake"
	"github.com/artefactual-sdps/enduro/internal/am"
	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	ingest_fake "github.com/artefactual-sdps/enduro/internal/ingest/fake"
	"github.com/artefactual-sdps/enduro/internal/poststorage"
	"github.com/artefactual-sdps/enduro/internal/premis"
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
	sipUUID        = uuid.MustParse("e2ace0da-8697-453d-9ea1-4c9b62309e54")
	locationID     = uuid.MustParse("f2cc963f-c14d-4eaa-b950-bd207189a1f1")
	amssLocationID = uuid.MustParse("e0ed8b2a-8ae2-4546-b5d8-f0090919df04")
	transferID     = uuid.MustParse("65233405-771e-4f7e-b2d9-b08439570ba2")
	aipUUID        = uuid.MustParse("9e8161cc-2815-4d6f-8a75-f003c41b257b")

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

func preprocessingChildWorkflow(
	ctx temporalsdk_workflow.Context,
	params *preprocessing.WorkflowParams,
) (*preprocessing.WorkflowResult, error) {
	return nil, nil
}

func poststorageChildWorkflow(
	ctx temporalsdk_workflow.Context,
	params *poststorage.WorkflowParams,
) (*any, error) {
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
	ingestsvc := ingest_fake.NewMockService(ctrl)
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
		xmlvalidate.New(xmlvalidate.NewXMLLintValidator()).Execute,
		temporalsdk_activity.RegisterOptions{Name: xmlvalidate.Name},
	)
	s.env.RegisterActivityWithOptions(
		bagvalidate.New(bagvalidate.NewNoopValidator()).Execute,
		temporalsdk_activity.RegisterOptions{Name: bagvalidate.Name},
	)
	s.env.RegisterActivityWithOptions(
		activities.NewClassifySIPActivity().Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.ClassifySIPActivityName},
	)

	// Set up AM taskqueue.
	if cfg.Preservation.TaskQueue == temporal.AmWorkerTaskQueue {
		s.setupAMWorkflowTest(&cfg.AM, ctrl, ingestsvc)
	} else {
		s.setupA3mWorkflowTest(ctrl, ingestsvc)
	}

	s.env.RegisterActivityWithOptions(
		removepaths.New().Execute,
		temporalsdk_activity.RegisterOptions{Name: removepaths.Name},
	)
	s.env.RegisterActivityWithOptions(
		activities.NewDeleteOriginalActivity(wsvc).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.DeleteOriginalActivityName},
	)
	s.env.RegisterActivityWithOptions(
		archivezip.New().Execute,
		temporalsdk_activity.RegisterOptions{Name: archivezip.Name},
	)
	s.env.RegisterActivityWithOptions(
		bucketupload.New(memblob.OpenBucket(nil)).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.SendToFailedSIPsName},
	)
	s.env.RegisterActivityWithOptions(
		bucketupload.New(memblob.OpenBucket(nil)).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.SendToFailedPIPsName},
	)

	s.env.RegisterWorkflowWithOptions(
		preprocessingChildWorkflow,
		temporalsdk_workflow.RegisterOptions{Name: "preprocessing"},
	)
	s.env.RegisterWorkflowWithOptions(
		poststorageChildWorkflow,
		temporalsdk_workflow.RegisterOptions{Name: "poststorage_1"},
	)
	s.env.RegisterWorkflowWithOptions(
		poststorageChildWorkflow,
		temporalsdk_workflow.RegisterOptions{Name: "poststorage_2"},
	)

	s.workflow = NewProcessingWorkflow(cfg, rng, ingestsvc, wsvc)
}

func (s *ProcessingWorkflowTestSuite) setupAMWorkflowTest(
	cfg *am.Config,
	ctrl *gomock.Controller,
	ingestsvc ingest.Service,
) {
	clock := clockwork.NewFakeClock()
	sftpc := sftp_fake.NewMockClient(ctrl)

	s.env.RegisterActivityWithOptions(
		bagcreate.New(bagcreate.Config{}).Execute,
		temporalsdk_activity.RegisterOptions{Name: bagcreate.Name},
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
			ingestsvc,
		).Execute,
		temporalsdk_activity.RegisterOptions{Name: am.PollTransferActivityName},
	)
	s.env.RegisterActivityWithOptions(
		am.NewPollIngestActivity(
			cfg,
			clock,
			amclienttest.NewMockIngestService(ctrl),
			amclienttest.NewMockJobsService(ctrl),
			ingestsvc,
		).Execute,
		temporalsdk_activity.RegisterOptions{Name: am.PollIngestActivityName},
	)
	s.env.RegisterActivityWithOptions(
		activities.NewCreateStorageAIPActivity(nil).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.CreateStorageAIPActivityName},
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
	ingestsvc ingest.Service,
) {
	tsvc := a3mfake.NewMockTransferServiceClient(ctrl)

	s.env.RegisterActivityWithOptions(
		activities.NewBundleActivity().Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.BundleActivityName},
	)
	s.env.RegisterActivityWithOptions(
		a3m.NewCreateAIPActivity(noop.Tracer{}, tsvc, &a3m.Config{}, ingestsvc).Execute,
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
		activities.NewRejectSIPActivity(nil).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.RejectSIPActivityName},
	)
}

func (s *ProcessingWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func TestProcessingWorkflow(t *testing.T) {
	suite.Run(t, new(ProcessingWorkflowTestSuite))
}

func (s *ProcessingWorkflowTestSuite) TestConfirmation() {
	cfg := config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Storage: storage.Config{
			DefaultPermanentLocationID: locationID,
		},
		ValidatePREMIS: premis.Config{
			Enabled: true,
			XSDPath: "/home/enduro/premis.xsd",
		},
	}
	s.SetupWorkflowTest(cfg)

	sipID := 1
	wID := 1
	valPREMISTaskID := 102
	ctx := mock.AnythingOfType("*context.valueCtx")
	sessionCtx := mock.AnythingOfType("*context.timerCtx")
	key := "transfer.zip"
	watcherName := "watcher"
	retentionPeriod := 1 * time.Second
	ingestsvc := s.workflow.ingestsvc

	// Signal handler that mimics SIP/AIP confirmation
	s.env.RegisterDelayedCallback(
		func() {
			s.env.SignalWorkflow(
				ingest.ReviewPerformedSignalName,
				ingest.ReviewPerformedSignal{Accepted: true, LocationID: &locationID},
			)
		},
		0,
	)

	// Activity mocks/assertions sequence
	s.env.OnActivity(createSIPLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(sipID, nil)
	s.env.OnActivity(setStatusInProgressLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(
		createWorkflowLocalActivity,
		mock.Anything,
		mock.Anything,
		&createWorkflowLocalActivityParams{
			TemporalID: "default-test-workflow-id",
			Type:       enums.WorkflowTypeCreateAndReviewAip,
			Status:     enums.WorkflowStatusInProgress,
			StartedAt:  startTime,
			SIPUUID:    sipUUID,
		},
	).Return(wID, nil)

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

	s.env.OnActivity(
		createTaskLocalActivity,
		ctx,
		&createTaskLocalActivityParams{
			Ingestsvc: ingestsvc,
			RNG:       s.workflow.rng,
			Task: datatypes.Task{
				Name:   "Validate PREMIS",
				Status: enums.TaskStatusInProgress,
				StartedAt: sql.NullTime{
					Time:  startTime,
					Valid: true,
				},
				WorkflowID: wID,
			},
		},
	).Return(valPREMISTaskID, nil)

	s.env.OnActivity(
		xmlvalidate.Name,
		sessionCtx,
		&xmlvalidate.Params{
			XMLPath: filepath.Join(transferPath, "metadata", "premis.xml"),
			XSDPath: cfg.ValidatePREMIS.XSDPath,
		},
	).Return(
		&xmlvalidate.Result{Failures: []string{}}, nil,
	)

	s.env.OnActivity(
		completeTaskLocalActivity,
		ctx,
		ingestsvc,
		&completeTaskLocalActivityParams{
			ID:          valPREMISTaskID,
			Status:      enums.TaskStatusDone,
			CompletedAt: startTime,
			Note:        ref.New("PREMIS is valid"),
		},
	).Return(&completeTaskLocalActivityResult{}, nil)

	s.env.OnActivity(a3m.CreateAIPActivityName, mock.Anything, mock.Anything).Return(nil, nil)
	s.env.OnActivity(updateSIPLocalActivity, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(createTaskLocalActivity, mock.Anything, mock.Anything, mock.Anything).
		Return(0, nil)
	s.env.OnActivity(activities.UploadActivityName, mock.Anything, mock.Anything).Return(nil, nil)
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(setWorkflowStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(completeTaskLocalActivity, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(activities.MoveToPermanentStorageActivityName, mock.Anything, mock.Anything).Return(nil, nil)
	s.env.OnActivity(activities.PollMoveToPermanentStorageActivityName, mock.Anything, mock.Anything).Return(nil, nil)
	s.env.OnActivity(completeWorkflowLocalActivity, ctx, ingestsvc, mock.AnythingOfType("*workflow.completeWorkflowLocalActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(
		removepaths.Name,
		sessionCtx,
		&removepaths.Params{Paths: []string{tempPath, transferPath}},
	).Return(&removepaths.Result{}, nil)
	s.env.OnActivity(activities.DeleteOriginalActivityName, sessionCtx, watcherName, key).Return(nil, nil).Once()
	s.env.OnActivity(updateSIPLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(completeWorkflowLocalActivity, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&ingest.ProcessingWorkflowRequest{
			Key:             "transfer.zip",
			WatcherName:     watcherName,
			RetentionPeriod: &retentionPeriod,
			AutoApproveAIP:  false,
			SIPUUID:         sipUUID,
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

	sipID := 1
	wID := 1
	valBagTaskID := 101
	moveAIPTaskID := 103
	watcherName := "watcher"
	key := "transfer.zip"
	retentionPeriod := 1 * time.Second
	ctx := mock.AnythingOfType("*context.valueCtx")
	sessionCtx := mock.AnythingOfType("*context.timerCtx")
	ingestsvc := s.workflow.ingestsvc
	rng := rand.New(rand.NewSource(1)) // #nosec: G404

	// Activity mocks/assertions sequence
	s.env.OnActivity(
		createSIPLocalActivity,
		ctx,
		ingestsvc,
		&createSIPLocalActivityParams{UUID: sipUUID, Name: key, Status: enums.SIPStatusQueued},
	).Return(sipID, nil).Once()
	s.env.OnActivity(
		setStatusInProgressLocalActivity,
		ctx,
		ingestsvc,
		sipUUID,
		mock.AnythingOfType("time.Time"),
	).Return(nil, nil)

	s.env.OnActivity(
		createWorkflowLocalActivity,
		ctx,
		ingestsvc,
		&createWorkflowLocalActivityParams{
			TemporalID: "default-test-workflow-id",
			Type:       enums.WorkflowTypeCreateAip,
			Status:     enums.WorkflowStatusInProgress,
			StartedAt:  startTime,
			SIPUUID:    sipUUID,
		},
	).Return(wID, nil)

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
		activities.ClassifySIPActivityName,
		sessionCtx,
		activities.ClassifySIPActivityParams{Path: extractPath},
	).Return(
		&activities.ClassifySIPActivityResult{Type: enums.SIPTypeBagIt}, nil,
	)

	s.env.OnActivity(
		createTaskLocalActivity,
		ctx,
		&createTaskLocalActivityParams{
			Ingestsvc: ingestsvc,
			RNG:       rng,
			Task: datatypes.Task{
				Name:   "Validate Bag",
				Status: enums.TaskStatusInProgress,
				StartedAt: sql.NullTime{
					Time:  startTime,
					Valid: true,
				},
				WorkflowID: wID,
			},
		},
	).Return(valBagTaskID, nil)

	s.env.OnActivity(
		bagvalidate.Name,
		sessionCtx,
		&bagvalidate.Params{Path: extractPath},
	).Return(
		&bagvalidate.Result{Valid: true},
		nil,
	)

	s.env.OnActivity(
		completeTaskLocalActivity,
		ctx,
		ingestsvc,
		&completeTaskLocalActivityParams{
			ID:          valBagTaskID,
			Status:      enums.TaskStatusDone,
			CompletedAt: startTime,
			Note:        ref.New("Bag successfully validated"),
		},
	).Return(&completeTaskLocalActivityResult{}, nil)

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
	s.env.OnActivity(updateSIPLocalActivity, ctx, ingestsvc, mock.AnythingOfType("*workflow.updateSIPLocalActivityParams")).
		Return(nil, nil).
		Times(2)

	s.env.OnActivity(
		createTaskLocalActivity,
		ctx,
		&createTaskLocalActivityParams{
			Ingestsvc: ingestsvc,
			RNG:       rng,
			Task: datatypes.Task{
				Name:       "Move AIP",
				Status:     enums.TaskStatusInProgress,
				StartedAt:  sql.NullTime{Time: startTime, Valid: true},
				Note:       "Moving to permanent storage",
				WorkflowID: wID,
			},
		},
	).Return(moveAIPTaskID, nil)

	s.env.OnActivity(activities.UploadActivityName, sessionCtx, mock.AnythingOfType("*activities.UploadActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil).
		Never()
	s.env.OnActivity(setWorkflowStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil).
		Never()

	s.env.OnActivity(
		completeTaskLocalActivity,
		ctx,
		ingestsvc,
		&completeTaskLocalActivityParams{
			ID:          moveAIPTaskID,
			Status:      enums.TaskStatusDone,
			CompletedAt: startTime,
			Note:        ref.New("Moved to location f2cc963f-c14d-4eaa-b950-bd207189a1f1"),
		},
	).Return(&completeTaskLocalActivityResult{}, nil)

	s.env.OnActivity(activities.MoveToPermanentStorageActivityName, sessionCtx, mock.AnythingOfType("*activities.MoveToPermanentStorageActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(activities.PollMoveToPermanentStorageActivityName, sessionCtx, mock.AnythingOfType("*activities.PollMoveToPermanentStorageActivityParams")).
		Return(nil, nil).
		Once()
	s.env.OnActivity(completeWorkflowLocalActivity, ctx, ingestsvc, mock.AnythingOfType("*workflow.completeWorkflowLocalActivityParams")).
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
		&ingest.ProcessingWorkflowRequest{
			Key:             key,
			WatcherName:     watcherName,
			RetentionPeriod: &retentionPeriod,
			AutoApproveAIP:  true,
			SIPUUID:         sipUUID,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowResult(nil))
}

func (s *ProcessingWorkflowTestSuite) TestAMWorkflow() {
	sipID := 1
	wID := 1
	valPREMISTaskID := 102
	watcherName := "watcher"
	key := "transfer.zip"
	retentionPeriod := 1 * time.Second
	ctx := mock.AnythingOfType("*context.valueCtx")
	sessionCtx := mock.AnythingOfType("*context.timerCtx")

	cfg := config.Configuration{
		A3m: a3m.Config{ShareDir: s.CreateTransferDir()},
		AM: am.Config{
			ZipPIP:           true,
			TransferDeadline: time.Second,
		},
		Preservation: pres.Config{TaskQueue: temporal.AmWorkerTaskQueue},
		Storage:      storage.Config{DefaultPermanentLocationID: amssLocationID},
		ValidatePREMIS: premis.Config{
			Enabled: true,
			XSDPath: "/home/enduro/premis.xsd",
		},
	}
	s.SetupWorkflowTest(cfg)

	ingestsvc := s.workflow.ingestsvc

	// Activity mocks/assertions sequence
	s.env.OnActivity(createSIPLocalActivity, ctx,
		ingestsvc,
		&createSIPLocalActivityParams{UUID: sipUUID, Name: key, Status: enums.SIPStatusQueued},
	).Return(sipID, nil)

	s.env.OnActivity(setStatusInProgressLocalActivity, ctx, ingestsvc, sipUUID, mock.AnythingOfType("time.Time")).
		Return(nil, nil)

	s.env.OnActivity(createWorkflowLocalActivity, ctx,
		ingestsvc, mock.AnythingOfType("*workflow.createWorkflowLocalActivityParams"),
	).Return(wID, nil)

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
		activities.ClassifySIPActivityName,
		sessionCtx,
		activities.ClassifySIPActivityParams{Path: extractPath},
	).Return(
		&activities.ClassifySIPActivityResult{Type: enums.SIPTypeUnknown}, nil,
	)

	// Archivematica specific activities.
	s.env.OnActivity(bagcreate.Name, sessionCtx,
		&bagcreate.Params{SourcePath: extractPath},
	).Return(
		&bagcreate.Result{BagPath: extractPath}, nil,
	)

	s.env.OnActivity(
		createTaskLocalActivity,
		ctx,
		&createTaskLocalActivityParams{
			Ingestsvc: ingestsvc,
			RNG:       s.workflow.rng,
			Task: datatypes.Task{
				Name:   "Validate PREMIS",
				Status: enums.TaskStatusInProgress,
				StartedAt: sql.NullTime{
					Time:  startTime,
					Valid: true,
				},
				WorkflowID: wID,
			},
		},
	).Return(valPREMISTaskID, nil)

	s.env.OnActivity(
		xmlvalidate.Name,
		sessionCtx,
		&xmlvalidate.Params{
			XMLPath: filepath.Join(extractPath, "data", "metadata", "premis.xml"),
			XSDPath: "/home/enduro/premis.xsd",
		},
	).Return(
		&xmlvalidate.Result{Failures: []string{}}, nil,
	)

	s.env.OnActivity(
		completeTaskLocalActivity,
		ctx,
		ingestsvc,
		&completeTaskLocalActivityParams{
			ID:          valPREMISTaskID,
			Status:      enums.TaskStatusDone,
			CompletedAt: startTime,
			Note:        ref.New("PREMIS is valid"),
		},
	).Return(&completeTaskLocalActivityResult{}, nil)

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
		&am.StartTransferActivityParams{Name: key, RelativePath: "transfer.zip", ZipPIP: true},
	).Return(
		&am.StartTransferActivityResult{TransferID: transferID.String()}, nil,
	)

	s.env.OnActivity(am.PollTransferActivityName, sessionCtx,
		&am.PollTransferActivityParams{TransferID: transferID.String(), WorkflowID: wID},
	).Return(
		&am.PollTransferActivityResult{SIPID: aipUUID.String()}, nil,
	)

	s.env.OnActivity(am.PollIngestActivityName, sessionCtx,
		&am.PollIngestActivityParams{SIPID: aipUUID.String(), WorkflowID: wID},
	).Return(
		&am.PollIngestActivityResult{Status: "COMPLETE"}, nil,
	)

	s.env.OnActivity(activities.CreateStorageAIPActivityName, sessionCtx,
		&activities.CreateStorageAIPActivityParams{
			Name:       key,
			AIPID:      aipUUID.String(),
			ObjectKey:  aipUUID.String(),
			Status:     "stored",
			LocationID: &amssLocationID,
		},
	).Return(&activities.CreateStorageAIPActivityResult{}, nil)

	s.env.OnActivity(am.DeleteTransferActivityName, sessionCtx,
		&am.DeleteTransferActivityParams{Destination: "transfer.zip"},
	).Return(nil, nil)

	// Post-preservation activities.
	s.env.OnActivity(
		updateSIPLocalActivity,
		ctx,
		ingestsvc,
		mock.AnythingOfType("*workflow.updateSIPLocalActivityParams"),
	).Return(nil, nil)
	s.env.OnActivity(
		completeWorkflowLocalActivity,
		ctx,
		ingestsvc,
		mock.AnythingOfType("*workflow.completeWorkflowLocalActivityParams"),
	).Return(nil, nil)
	s.env.OnActivity(
		removepaths.Name,
		sessionCtx,
		&removepaths.Params{Paths: []string{tempPath, extractPath + "/transfer.zip"}},
	).Return(&removepaths.Result{}, nil)
	s.env.OnActivity(
		activities.DeleteOriginalActivityName,
		sessionCtx,
		watcherName,
		key,
	).Return(nil, nil)

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&ingest.ProcessingWorkflowRequest{
			WatcherName:     watcherName,
			RetentionPeriod: &retentionPeriod,
			AutoApproveAIP:  true,
			Key:             key,
			SIPUUID:         sipUUID,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowResult(nil))
}

func (s *ProcessingWorkflowTestSuite) TestRejection() {
	cfg := config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Storage: storage.Config{
			DefaultPermanentLocationID: locationID,
		},
	}
	s.SetupWorkflowTest(cfg)

	sipID := 1
	key := "transfer.zip"
	watcherName := "watcher"
	retentionPeriod := 1 * time.Second
	sessionCtx := mock.AnythingOfType("*context.timerCtx")

	// Signal handler that mimics SIP/AIP rejection
	s.env.RegisterDelayedCallback(
		func() {
			s.env.SignalWorkflow(
				ingest.ReviewPerformedSignalName,
				ingest.ReviewPerformedSignal{Accepted: false},
			)
		},
		0,
	)

	// Activity mocks/assertions sequence
	s.env.OnActivity(createSIPLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(sipID, nil)
	s.env.OnActivity(setStatusInProgressLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(createWorkflowLocalActivity, mock.Anything, mock.Anything, mock.Anything).
		Return(0, nil)

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
		activities.ClassifySIPActivityName,
		sessionCtx,
		activities.ClassifySIPActivityParams{Path: extractPath},
	).Return(
		&activities.ClassifySIPActivityResult{Type: enums.SIPTypeUnknown}, nil,
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
	s.env.OnActivity(updateSIPLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(completeTaskLocalActivity, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(activities.UploadActivityName, mock.Anything, mock.Anything).Return(nil, nil)
	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(setWorkflowStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(createTaskLocalActivity, mock.Anything, mock.Anything, mock.Anything).
		Return(0, nil)
	s.env.OnActivity(activities.RejectSIPActivityName, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(
		removepaths.Name,
		sessionCtx,
		&removepaths.Params{Paths: []string{tempPath, transferPath}},
	).Return(&removepaths.Result{}, nil)
	s.env.OnActivity(activities.DeleteOriginalActivityName, sessionCtx, watcherName, key).Return(nil, nil).Once()
	s.env.OnActivity(updateSIPLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)
	s.env.OnActivity(completeWorkflowLocalActivity, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&ingest.ProcessingWorkflowRequest{
			Key:             "transfer.zip",
			WatcherName:     watcherName,
			RetentionPeriod: &retentionPeriod,
			AutoApproveAIP:  false,
			SIPUUID:         sipUUID,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowResult(nil))
}

func (s *ProcessingWorkflowTestSuite) TestChildWorkflows() {
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
		Poststorage: []poststorage.Config{
			{
				Namespace:    "default",
				TaskQueue:    "poststorage",
				WorkflowName: "poststorage_1",
			},
			{
				Namespace:    "default",
				TaskQueue:    "poststorage",
				WorkflowName: "poststorage_2",
			},
		},
		Storage: storage.Config{
			DefaultPermanentLocationID: locationID,
		},
	}
	s.SetupWorkflowTest(cfg)

	sipID := 1
	valBagTaskID := 101
	moveAIPTaskID := 103
	watcherName := "watcher"
	key := "transfer.zip"
	retentionPeriod := 1 * time.Second
	ctx := mock.AnythingOfType("*context.valueCtx")
	sessionCtx := mock.AnythingOfType("*context.timerCtx")
	ingestsvc := s.workflow.ingestsvc
	aipUUID := "56eebd45-5600-4768-a8c2-ec0114555a3d"

	downloadDir := strings.Replace(tempPath, "/tmp/", cfg.Preprocessing.SharedPath, 1)
	prepDest := strings.Replace(extractPath, "/tmp/", cfg.Preprocessing.SharedPath, 1)

	// Activity mocks/assertions sequence
	s.env.OnActivity(
		createSIPLocalActivity,
		ctx,
		ingestsvc,
		&createSIPLocalActivityParams{UUID: sipUUID, Name: key, Status: enums.SIPStatusQueued},
	).Return(sipID, nil)

	s.env.OnActivity(
		setStatusInProgressLocalActivity,
		ctx,
		ingestsvc,
		sipUUID,
		startTime,
	).Return(nil, nil)

	s.env.OnActivity(
		createWorkflowLocalActivity,
		ctx,
		ingestsvc,
		&createWorkflowLocalActivityParams{
			TemporalID: "default-test-workflow-id",
			Type:       enums.WorkflowTypeCreateAip,
			Status:     enums.WorkflowStatusInProgress,
			StartedAt:  startTime,
			SIPUUID:    sipUUID,
		},
	).Return(1, nil)

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
			Ingestsvc:  ingestsvc,
			RNG:        rand.New(rand.NewSource(1)), // #nosec: G404
			WorkflowID: 1,
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
		activities.ClassifySIPActivityName,
		sessionCtx,
		activities.ClassifySIPActivityParams{Path: prepDest},
	).Return(
		&activities.ClassifySIPActivityResult{Type: enums.SIPTypeBagIt}, nil,
	)

	s.env.OnActivity(
		createTaskLocalActivity,
		ctx,
		&createTaskLocalActivityParams{
			Ingestsvc: ingestsvc,
			RNG:       s.workflow.rng,
			Task: datatypes.Task{
				Name:   "Validate Bag",
				Status: enums.TaskStatusInProgress,
				StartedAt: sql.NullTime{
					Time:  startTime,
					Valid: true,
				},
				WorkflowID: 1,
			},
		},
	).Return(valBagTaskID, nil)

	s.env.OnActivity(
		bagvalidate.Name,
		sessionCtx,
		&bagvalidate.Params{Path: prepDest},
	).Return(
		&bagvalidate.Result{Valid: true},
		nil,
	)

	s.env.OnActivity(
		completeTaskLocalActivity,
		ctx,
		ingestsvc,
		&completeTaskLocalActivityParams{
			ID:          valBagTaskID,
			Status:      enums.TaskStatusDone,
			CompletedAt: startTime,
			Note:        ref.New("Bag successfully validated"),
		},
	).Return(&completeTaskLocalActivityResult{}, nil)

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
		Return(&a3m.CreateAIPActivityResult{UUID: aipUUID}, nil)

	s.env.OnActivity(updateSIPLocalActivity, ctx, ingestsvc, mock.AnythingOfType("*workflow.updateSIPLocalActivityParams")).
		Return(nil, nil).
		Times(2)

	s.env.OnActivity(
		createTaskLocalActivity,
		ctx,
		&createTaskLocalActivityParams{
			Ingestsvc: ingestsvc,
			RNG:       s.workflow.rng,
			Task: datatypes.Task{
				Name:       "Move AIP",
				Status:     enums.TaskStatusInProgress,
				StartedAt:  sql.NullTime{Time: startTime, Valid: true},
				WorkflowID: 1,
				Note:       "Moving to permanent storage",
			},
		},
	).Return(moveAIPTaskID, nil)

	s.env.OnActivity(
		activities.UploadActivityName,
		sessionCtx,
		mock.AnythingOfType("*activities.UploadActivityParams"),
	).Return(nil, nil)

	s.env.OnActivity(
		completeTaskLocalActivity,
		ctx,
		ingestsvc,
		&completeTaskLocalActivityParams{
			ID:          moveAIPTaskID,
			Status:      enums.TaskStatusDone,
			CompletedAt: startTime,
			Note:        ref.New("Moved to location f2cc963f-c14d-4eaa-b950-bd207189a1f1"),
		},
	).Return(&completeTaskLocalActivityResult{}, nil)

	s.env.OnActivity(setStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil).
		Never()
	s.env.OnActivity(setWorkflowStatusLocalActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil).
		Never()

	s.env.OnActivity(activities.MoveToPermanentStorageActivityName, sessionCtx, mock.AnythingOfType("*activities.MoveToPermanentStorageActivityParams")).
		Return(nil, nil).
		Once()

	s.env.OnActivity(activities.PollMoveToPermanentStorageActivityName, sessionCtx, mock.AnythingOfType("*activities.PollMoveToPermanentStorageActivityParams")).
		Return(nil, nil).
		Once()

	s.env.OnActivity(completeWorkflowLocalActivity, ctx, ingestsvc, mock.AnythingOfType("*workflow.completeWorkflowLocalActivityParams")).
		Return(nil, nil).
		Once()

	s.env.OnWorkflow(
		"poststorage_1",
		mock.AnythingOfType("*internal.valueCtx"),
		&poststorage.WorkflowParams{
			AIPUUID: aipUUID,
		},
	).Return(nil, nil)

	s.env.OnWorkflow(
		"poststorage_2",
		mock.AnythingOfType("*internal.valueCtx"),
		&poststorage.WorkflowParams{
			AIPUUID: aipUUID,
		},
	).Return(nil, nil)

	s.env.OnActivity(
		removepaths.Name,
		sessionCtx,
		&removepaths.Params{Paths: []string{downloadDir, transferPath}},
	).Return(&removepaths.Result{}, nil)

	s.env.OnActivity(activities.DeleteOriginalActivityName, sessionCtx, watcherName, key).Return(nil, nil).Once()

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&ingest.ProcessingWorkflowRequest{
			Key:             key,
			WatcherName:     watcherName,
			RetentionPeriod: &retentionPeriod,
			AutoApproveAIP:  true,
			SIPUUID:         sipUUID,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowResult(nil))
}

func (s *ProcessingWorkflowTestSuite) TestFailedSIP() {
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
		Storage: storage.Config{DefaultPermanentLocationID: locationID},
	}
	s.SetupWorkflowTest(cfg)

	sipID := 1
	watcherName := "watcher"
	key := "transfer.zip"
	retentionPeriod := 1 * time.Second
	ctx := mock.AnythingOfType("*context.valueCtx")
	sessionCtx := mock.AnythingOfType("*context.timerCtx")
	downloadDir := strings.Replace(tempPath, "/tmp/", cfg.Preprocessing.SharedPath, 1)
	prepDest := strings.Replace(extractPath, "/tmp/", cfg.Preprocessing.SharedPath, 1)
	ingestsvc := s.workflow.ingestsvc

	s.env.OnActivity(
		createSIPLocalActivity,
		ctx,
		ingestsvc,
		&createSIPLocalActivityParams{UUID: sipUUID, Name: key, Status: enums.SIPStatusQueued},
	).Return(sipID, nil)

	s.env.OnActivity(
		setStatusInProgressLocalActivity,
		ctx,
		ingestsvc,
		sipUUID,
		startTime,
	).Return(nil, nil)

	s.env.OnActivity(
		createWorkflowLocalActivity,
		ctx,
		ingestsvc,
		&createWorkflowLocalActivityParams{
			TemporalID: "default-test-workflow-id",
			Type:       enums.WorkflowTypeCreateAip,
			Status:     enums.WorkflowStatusInProgress,
			StartedAt:  startTime,
			SIPUUID:    sipUUID,
		},
	).Return(1, nil)

	s.env.OnActivity(
		activities.DownloadActivityName,
		sessionCtx,
		&activities.DownloadActivityParams{
			Key:             key,
			WatcherName:     watcherName,
			DestinationPath: cfg.Preprocessing.SharedPath,
		},
	).Return(&activities.DownloadActivityResult{Path: downloadDir + "/" + key}, nil)

	s.env.OnWorkflow(
		preprocessingChildWorkflow,
		mock.AnythingOfType("*internal.valueCtx"),
		&preprocessing.WorkflowParams{
			RelativePath: strings.TrimPrefix(downloadDir+"/"+key, cfg.Preprocessing.SharedPath),
		},
	).Return(
		&preprocessing.WorkflowResult{
			Outcome:      preprocessing.OutcomeContentError,
			RelativePath: strings.TrimPrefix(prepDest, cfg.Preprocessing.SharedPath),
		},
		nil,
	)

	s.env.OnActivity(
		activities.SendToFailedSIPsName,
		sessionCtx,
		&bucketupload.Params{
			Path:       downloadDir + "/" + key,
			Key:        fmt.Sprintf("Failed_%s", key),
			BufferSize: 100_000_000,
		},
	).Return(&bucketupload.Result{}, nil)

	s.env.OnActivity(
		updateSIPLocalActivity,
		ctx,
		ingestsvc,
		mock.AnythingOfType("*workflow.updateSIPLocalActivityParams"),
	).Return(nil, nil)

	s.env.OnActivity(
		completeWorkflowLocalActivity,
		ctx,
		ingestsvc,
		mock.AnythingOfType("*workflow.completeWorkflowLocalActivityParams"),
	).Return(nil, nil)

	s.env.OnActivity(
		removepaths.Name,
		sessionCtx,
		&removepaths.Params{Paths: []string{downloadDir}},
	).Return(&removepaths.Result{}, nil)

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&ingest.ProcessingWorkflowRequest{
			Key:             key,
			WatcherName:     watcherName,
			RetentionPeriod: &retentionPeriod,
			AutoApproveAIP:  true,
			SIPUUID:         sipUUID,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowResult(nil))
}

func (s *ProcessingWorkflowTestSuite) TestFailedPIPA3m() {
	cfg := config.Configuration{
		A3m:          a3m.Config{ShareDir: s.CreateTransferDir()},
		Preservation: pres.Config{TaskQueue: temporal.A3mWorkerTaskQueue},
		Storage:      storage.Config{DefaultPermanentLocationID: locationID},
	}
	s.SetupWorkflowTest(cfg)

	sipID := 1
	valBagTaskID := 101
	watcherName := "watcher"
	key := "transfer.zip"
	retentionPeriod := 1 * time.Second
	ctx := mock.AnythingOfType("*context.valueCtx")
	sessionCtx := mock.AnythingOfType("*context.timerCtx")
	downloadDir := strings.Replace(tempPath, "/tmp/", cfg.Preprocessing.SharedPath, 1)
	ingestsvc := s.workflow.ingestsvc

	s.env.OnActivity(
		createSIPLocalActivity,
		ctx,
		ingestsvc,
		&createSIPLocalActivityParams{UUID: sipUUID, Name: key, Status: enums.SIPStatusQueued},
	).Return(sipID, nil)

	s.env.OnActivity(
		setStatusInProgressLocalActivity,
		ctx,
		ingestsvc,
		sipUUID,
		startTime,
	).Return(nil, nil)

	s.env.OnActivity(
		createWorkflowLocalActivity,
		ctx,
		ingestsvc,
		&createWorkflowLocalActivityParams{
			TemporalID: "default-test-workflow-id",
			Type:       enums.WorkflowTypeCreateAip,
			Status:     enums.WorkflowStatusInProgress,
			StartedAt:  startTime,
			SIPUUID:    sipUUID,
		},
	).Return(1, nil)

	s.env.OnActivity(
		activities.DownloadActivityName,
		sessionCtx,
		&activities.DownloadActivityParams{
			Key:             key,
			WatcherName:     watcherName,
			DestinationPath: cfg.Preprocessing.SharedPath,
		},
	).Return(&activities.DownloadActivityResult{Path: downloadDir + "/" + key}, nil)

	s.env.OnActivity(
		archiveextract.Name,
		sessionCtx,
		&archiveextract.Params{SourcePath: downloadDir + "/" + key},
	).Return(&archiveextract.Result{ExtractPath: extractPath}, nil)

	s.env.OnActivity(
		activities.ClassifySIPActivityName,
		sessionCtx,
		activities.ClassifySIPActivityParams{Path: extractPath},
	).Return(&activities.ClassifySIPActivityResult{Type: enums.SIPTypeBagIt}, nil)

	s.env.OnActivity(
		createTaskLocalActivity,
		ctx,
		&createTaskLocalActivityParams{
			Ingestsvc: ingestsvc,
			RNG:       rand.New(rand.NewSource(1)), // #nosec: G404
			Task: datatypes.Task{
				Name:   "Validate Bag",
				Status: enums.TaskStatusInProgress,
				StartedAt: sql.NullTime{
					Time:  startTime,
					Valid: true,
				},
				WorkflowID: 1,
			},
		},
	).Return(valBagTaskID, nil)

	s.env.OnActivity(
		bagvalidate.Name,
		sessionCtx,
		&bagvalidate.Params{Path: extractPath},
	).Return(&bagvalidate.Result{Valid: true}, nil)

	s.env.OnActivity(
		completeTaskLocalActivity,
		ctx,
		ingestsvc,
		&completeTaskLocalActivityParams{
			ID:          valBagTaskID,
			Status:      enums.TaskStatusDone,
			CompletedAt: startTime,
			Note:        ref.New("Bag successfully validated"),
		},
	).Return(&completeTaskLocalActivityResult{}, nil)

	s.env.OnActivity(
		activities.BundleActivityName,
		sessionCtx,
		&activities.BundleActivityParams{
			SourcePath:  extractPath,
			TransferDir: s.transferDir,
			IsDir:       true,
		},
	).Return(&activities.BundleActivityResult{FullPath: transferPath}, nil)

	s.env.OnActivity(a3m.CreateAIPActivityName, sessionCtx, mock.AnythingOfType("*a3m.CreateAIPActivityParams")).
		Return(nil, fmt.Errorf("a3m error"))

	s.env.OnActivity(archivezip.Name, sessionCtx, &archivezip.Params{SourceDir: transferPath}).
		Return(&archivezip.Result{Path: transferPath + ".zip"}, nil)

	s.env.OnActivity(
		activities.SendToFailedPIPsName,
		sessionCtx,
		&bucketupload.Params{
			Path:       transferPath + ".zip",
			Key:        fmt.Sprintf("Failed_%s", key),
			BufferSize: 100_000_000,
		},
	).Return(&bucketupload.Result{}, nil)

	s.env.OnActivity(
		updateSIPLocalActivity,
		ctx,
		ingestsvc,
		mock.AnythingOfType("*workflow.updateSIPLocalActivityParams"),
	).Return(nil, nil)

	s.env.OnActivity(
		completeWorkflowLocalActivity,
		ctx,
		ingestsvc,
		mock.AnythingOfType("*workflow.completeWorkflowLocalActivityParams"),
	).Return(nil, nil)

	s.env.OnActivity(
		removepaths.Name,
		sessionCtx,
		&removepaths.Params{Paths: []string{downloadDir, transferPath}},
	).Return(&removepaths.Result{}, nil)

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&ingest.ProcessingWorkflowRequest{
			Key:             key,
			WatcherName:     watcherName,
			RetentionPeriod: &retentionPeriod,
			AutoApproveAIP:  true,
			SIPUUID:         sipUUID,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowResult(nil))
}

func (s *ProcessingWorkflowTestSuite) TestFailedPIPAM() {
	cfg := config.Configuration{
		AM:           am.Config{ZipPIP: true},
		Preservation: pres.Config{TaskQueue: temporal.AmWorkerTaskQueue},
		Storage:      storage.Config{DefaultPermanentLocationID: amssLocationID},
	}
	s.SetupWorkflowTest(cfg)

	sipID := 1
	wID := 1
	watcherName := "watcher"
	key := "transfer.zip"
	retentionPeriod := 1 * time.Second
	ctx := mock.AnythingOfType("*context.valueCtx")
	sessionCtx := mock.AnythingOfType("*context.timerCtx")
	ingestsvc := s.workflow.ingestsvc

	s.env.OnActivity(
		createSIPLocalActivity,
		ctx,
		ingestsvc,
		&createSIPLocalActivityParams{UUID: sipUUID, Name: key, Status: enums.SIPStatusQueued},
	).Return(sipID, nil)

	s.env.OnActivity(setStatusInProgressLocalActivity, ctx, ingestsvc, sipUUID, mock.AnythingOfType("time.Time")).
		Return(nil, nil)

	s.env.OnActivity(
		createWorkflowLocalActivity,
		ctx,
		ingestsvc,
		&createWorkflowLocalActivityParams{
			TemporalID: "default-test-workflow-id",
			Type:       enums.WorkflowTypeCreateAip,
			Status:     enums.WorkflowStatusInProgress,
			StartedAt:  startTime,
			SIPUUID:    sipUUID,
		},
	).Return(wID, nil)

	s.env.OnActivity(
		activities.DownloadActivityName,
		sessionCtx,
		&activities.DownloadActivityParams{Key: key, WatcherName: watcherName},
	).Return(&activities.DownloadActivityResult{Path: tempPath + "/" + key}, nil)

	s.env.OnActivity(
		archiveextract.Name,
		sessionCtx,
		&archiveextract.Params{SourcePath: tempPath + "/" + key},
	).Return(&archiveextract.Result{ExtractPath: extractPath}, nil)

	s.env.OnActivity(
		activities.ClassifySIPActivityName,
		sessionCtx,
		activities.ClassifySIPActivityParams{Path: extractPath},
	).Return(&activities.ClassifySIPActivityResult{Type: enums.SIPTypeUnknown}, nil)

	s.env.OnActivity(bagcreate.Name, sessionCtx, &bagcreate.Params{SourcePath: extractPath}).
		Return(&bagcreate.Result{BagPath: extractPath}, nil)

	s.env.OnActivity(archivezip.Name, sessionCtx, &archivezip.Params{SourceDir: extractPath}).
		Return(&archivezip.Result{Path: extractPath + "/transfer.zip"}, nil)

	s.env.OnActivity(
		am.UploadTransferActivityName,
		sessionCtx,
		&am.UploadTransferActivityParams{SourcePath: extractPath + "/transfer.zip"},
	).Return(nil, fmt.Errorf("AM error"))

	s.env.OnActivity(
		activities.SendToFailedPIPsName,
		sessionCtx,
		&bucketupload.Params{
			Path:       extractPath + "/transfer.zip",
			Key:        fmt.Sprintf("Failed_%s", key),
			BufferSize: 100_000_000,
		},
	).Return(&bucketupload.Result{}, nil)

	s.env.OnActivity(
		updateSIPLocalActivity,
		ctx,
		ingestsvc,
		mock.AnythingOfType("*workflow.updateSIPLocalActivityParams"),
	).Return(nil, nil)

	s.env.OnActivity(
		completeWorkflowLocalActivity,
		ctx,
		ingestsvc,
		mock.AnythingOfType("*workflow.completeWorkflowLocalActivityParams"),
	).Return(nil, nil)

	s.env.OnActivity(
		removepaths.Name,
		sessionCtx,
		&removepaths.Params{Paths: []string{tempPath, extractPath + "/transfer.zip"}},
	).Return(&removepaths.Result{}, nil)

	s.env.ExecuteWorkflow(
		s.workflow.Execute,
		&ingest.ProcessingWorkflowRequest{
			WatcherName:     watcherName,
			RetentionPeriod: &retentionPeriod,
			AutoApproveAIP:  true,
			Key:             key,
			SIPUUID:         sipUUID,
		},
	)

	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowResult(nil))
}
