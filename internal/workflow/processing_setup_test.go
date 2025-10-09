package workflow

import (
	"math/rand"
	"time"

	"github.com/artefactual-sdps/temporal-activities/archiveextract"
	"github.com/artefactual-sdps/temporal-activities/archivezip"
	"github.com/artefactual-sdps/temporal-activities/bagcreate"
	"github.com/artefactual-sdps/temporal-activities/bagvalidate"
	"github.com/artefactual-sdps/temporal-activities/bucketcopy"
	"github.com/artefactual-sdps/temporal-activities/bucketdelete"
	"github.com/artefactual-sdps/temporal-activities/bucketdownload"
	"github.com/artefactual-sdps/temporal-activities/bucketupload"
	"github.com/artefactual-sdps/temporal-activities/removepaths"
	"github.com/artefactual-sdps/temporal-activities/xmlvalidate"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/suite"
	"go.artefactual.dev/amclient/amclienttest"
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
	"github.com/artefactual-sdps/enduro/internal/ingest"
	ingest_fake "github.com/artefactual-sdps/enduro/internal/ingest/fake"
	"github.com/artefactual-sdps/enduro/internal/poststorage"
	"github.com/artefactual-sdps/enduro/internal/preprocessing"
	sftp_fake "github.com/artefactual-sdps/enduro/internal/sftp/fake"
	"github.com/artefactual-sdps/enduro/internal/temporal"
	watcherfake "github.com/artefactual-sdps/enduro/internal/watcher/fake"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
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
		bucketdownload.New(nil).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.DownloadFromSIPSourceActivityName},
	)
	s.env.RegisterActivityWithOptions(
		bucketdownload.New(nil).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.DownloadFromInternalBucketActivityName},
	)
	s.env.RegisterActivityWithOptions(
		activities.NewGetSIPExtensionActivity().Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.GetSIPExtensionActivityName},
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
		bucketcopy.New(memblob.OpenBucket(nil)).Execute,
		temporalsdk_activity.RegisterOptions{Name: bucketcopy.Name},
	)
	s.env.RegisterActivityWithOptions(
		bucketdelete.New(memblob.OpenBucket(nil)).Execute,
		temporalsdk_activity.RegisterOptions{Name: bucketdelete.Name},
	)
	s.env.RegisterActivityWithOptions(
		bucketupload.New(memblob.OpenBucket(nil)).Execute,
		temporalsdk_activity.RegisterOptions{Name: bucketupload.Name},
	)
	s.env.RegisterActivityWithOptions(
		bucketdelete.New(nil).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.DeleteOriginalFromSIPSourceActivityName},
	)
	s.env.RegisterActivityWithOptions(
		bucketdelete.New(nil).Execute,
		temporalsdk_activity.RegisterOptions{Name: activities.DeleteOriginalFromInternalBucketActivityName},
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

func (s *ProcessingWorkflowTestSuite) ExecuteAndValidateWorkflow(
	req *ingest.ProcessingWorkflowRequest,
	shouldError bool,
) {
	s.env.ExecuteWorkflow(s.workflow.Execute, req)
	s.True(s.env.IsWorkflowCompleted())
	if shouldError {
		s.Error(s.env.GetWorkflowResult(nil))
	} else {
		s.NoError(s.env.GetWorkflowResult(nil))
	}
}
