package a3m_test

import (
	"fmt"
	"testing"
	"time"

	transferservice "buf.build/gen/go/artefactual/a3m/protocolbuffers/go/a3m/api/transferservice/v1beta1"
	"github.com/google/uuid"
	"go.artefactual.dev/tools/mockutil"
	"go.opentelemetry.io/otel/trace/noop"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/a3m"
	a3mfake "github.com/artefactual-sdps/enduro/internal/a3m/fake"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	ingest_fake "github.com/artefactual-sdps/enduro/internal/ingest/fake"
)

func TestCreateAIPActivity(t *testing.T) {
	t.Parallel()

	taskUUID := uuid.New()
	aipID := "55f00def-cdf7-4e9c-97fd-700980b993b3"

	ts := &temporalsdk_testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()
	ctrl := gomock.NewController(t)

	a3mTransferServiceClient := a3mfake.NewMockTransferServiceClient(ctrl)
	a3mTransferServiceClient.EXPECT().
		Submit(
			mockutil.Context(),
			gomock.AssignableToTypeOf(&transferservice.SubmitRequest{}),
			grpc.WaitForReady(true),
		).
		Return(
			&transferservice.SubmitResponse{
				Id: aipID,
			},
			nil,
		)
	a3mTransferServiceClient.EXPECT().
		Read(
			mockutil.Context(),
			&transferservice.ReadRequest{
				Id: aipID,
			},
		).
		Return(
			&transferservice.ReadResponse{
				Status: transferservice.PackageStatus_PACKAGE_STATUS_COMPLETE,
				Jobs: []*transferservice.Job{
					{
						Id:        taskUUID.String(),
						Status:    transferservice.Job_STATUS_COMPLETE,
						StartTime: timestamppb.New(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
					},
				},
			},
			nil,
		)

	ingestsvc := ingest_fake.NewMockService(ctrl)
	ingestsvc.EXPECT().CreateTasks(
		mockutil.Context(),
		[]*datatypes.Task{
			{
				UUID:         taskUUID,
				Status:       enums.TaskStatusDone,
				StartedAt:    time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				WorkflowUUID: uuid.MustParse("bbb608ab-95e9-4e60-be8e-0372195d8e05"),
			},
		},
	).Return(nil)

	shareDir := t.TempDir()

	env.RegisterActivityWithOptions(
		a3m.NewCreateAIPActivity(
			noop.Tracer{},
			a3mTransferServiceClient,
			&a3m.Config{
				Name:     "sip.zip",
				ShareDir: shareDir,
			},
			ingestsvc,
		).Execute,
		temporalsdk_activity.RegisterOptions{
			Name: a3m.CreateAIPActivityName,
		},
	)

	future, err := env.ExecuteActivity(a3m.CreateAIPActivityName, &a3m.CreateAIPActivityParams{
		Name:         "sip.zip",
		Path:         shareDir + "/sip.zip",
		WorkflowUUID: uuid.MustParse("bbb608ab-95e9-4e60-be8e-0372195d8e05"),
	})
	assert.NilError(t, err)

	var got *a3m.CreateAIPActivityResult
	assert.NilError(t, future.Get(&got))
	assert.DeepEqual(t, got, &a3m.CreateAIPActivityResult{
		Name: fmt.Sprintf("sip.zip-%s.7z", aipID),
		Path: fmt.Sprintf("%s/completed/sip.zip-%s.7z", shareDir, aipID),
		UUID: aipID,
	})
}
