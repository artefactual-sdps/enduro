package a3m

import (
	context "context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"buf.build/gen/go/artefactual/a3m/grpc/go/a3m/api/transferservice/v1beta1/transferservicev1beta1grpc"
	transferservice "buf.build/gen/go/artefactual/a3m/protocolbuffers/go/a3m/api/transferservice/v1beta1"
	"github.com/oklog/run"
	temporal_tools "go.artefactual.dev/tools/temporal"
	"go.opentelemetry.io/otel/trace"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	"google.golang.org/grpc"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/telemetry"
)

const CreateAIPActivityName = "create-aip-activity"

type CreateAIPActivity struct {
	tracer    trace.Tracer
	client    transferservicev1beta1grpc.TransferServiceClient
	cfg       *Config
	ingestsvc ingest.Service
}

type CreateAIPActivityParams struct {
	Name       string
	Path       string
	WorkflowID int
}

type CreateAIPActivityResult struct {
	Path string
	UUID string
}

func NewCreateAIPActivity(
	tracer trace.Tracer,
	client transferservicev1beta1grpc.TransferServiceClient,
	cfg *Config,
	ingestsvc ingest.Service,
) *CreateAIPActivity {
	return &CreateAIPActivity{
		tracer:    tracer,
		client:    client,
		cfg:       cfg,
		ingestsvc: ingestsvc,
	}
}

func (a *CreateAIPActivity) Execute(
	ctx context.Context,
	opts *CreateAIPActivityParams,
) (*CreateAIPActivityResult, error) {
	logger := temporal_tools.GetLogger(ctx)
	result := &CreateAIPActivityResult{}

	var g run.Group

	{
		cancel := make(chan struct{})

		g.Add(
			func() error {
				ticker := time.NewTicker(time.Second * 2)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-cancel:
						return nil
					case <-ticker.C:
						cp := "in progress"
						temporalsdk_activity.RecordHeartbeat(ctx, cp)
					}
				}
			},
			func(error) {
				close(cancel)
			},
		)
	}

	{
		g.Add(
			func() error {
				submitResp, err := a.client.Submit(
					ctx,
					&transferservice.SubmitRequest{
						Name: opts.Name,
						Url:  fmt.Sprintf("file://%s", opts.Path),
						Config: &transferservice.ProcessingConfig{
							AssignUuidsToDirectories:                     a.cfg.AssignUuidsToDirectories,
							ExamineContents:                              a.cfg.ExamineContents,
							GenerateTransferStructureReport:              a.cfg.GenerateTransferStructureReport,
							DocumentEmptyDirectories:                     a.cfg.DocumentEmptyDirectories,
							ExtractPackages:                              a.cfg.ExtractPackages,
							DeletePackagesAfterExtraction:                a.cfg.DeletePackagesAfterExtraction,
							IdentifyTransfer:                             a.cfg.IdentifyTransfer,
							IdentifySubmissionAndMetadata:                a.cfg.IdentifySubmissionAndMetadata,
							IdentifyBeforeNormalization:                  a.cfg.IdentifyBeforeNormalization,
							Normalize:                                    a.cfg.Normalize,
							TranscribeFiles:                              a.cfg.TranscribeFiles,
							PerformPolicyChecksOnOriginals:               a.cfg.PerformPolicyChecksOnOriginals,
							PerformPolicyChecksOnPreservationDerivatives: a.cfg.PerformPolicyChecksOnPreservationDerivatives,
							AipCompressionLevel:                          a.cfg.AipCompressionLevel,
							AipCompressionAlgorithm:                      a.cfg.AipCompressionAlgorithm,
						},
					},
					grpc.WaitForReady(true),
				)
				if err != nil {
					return err
				}

				result.UUID = submitResp.Id

				for {
					readResp, err := a.client.Read(ctx, &transferservice.ReadRequest{Id: result.UUID})
					if err != nil {
						return err
					}

					if readResp.Status == transferservice.PackageStatus_PACKAGE_STATUS_PROCESSING {
						time.Sleep(time.Second / 2)
						continue
					}

					err = saveTasks(ctx, a.tracer, readResp.Jobs, a.ingestsvc, opts.WorkflowID)
					if err != nil {
						return err
					}

					if readResp.Status == transferservice.PackageStatus_PACKAGE_STATUS_FAILED ||
						readResp.Status == transferservice.PackageStatus_PACKAGE_STATUS_REJECTED {
						return errors.New("package failed or rejected")
					}

					result.Path = fmt.Sprintf("%s/completed/%s-%s.7z", a.cfg.ShareDir, opts.Name, result.UUID)
					logger.Info("We have run a3m successfully", "path", result.Path)

					break
				}

				return nil
			},
			func(error) {},
		)
	}

	err := g.Run()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func saveTasks(
	ctx context.Context,
	tracer trace.Tracer,
	jobs []*transferservice.Job,
	ingestsvc ingest.Service,
	wID int,
) error {
	ctx, span := tracer.Start(ctx, "saveTasks")
	defer span.End()

	jobStatusToTaskStatus := map[transferservice.Job_Status]enums.TaskStatus{
		transferservice.Job_STATUS_UNSPECIFIED: enums.TaskStatusUnspecified,
		transferservice.Job_STATUS_COMPLETE:    enums.TaskStatusDone,
		transferservice.Job_STATUS_PROCESSING:  enums.TaskStatusInProgress,
		transferservice.Job_STATUS_FAILED:      enums.TaskStatusError,
	}

	for _, job := range jobs {
		task := datatypes.Task{
			TaskID: job.Id,
			Name:   job.Name,
			Status: jobStatusToTaskStatus[job.Status],
			StartedAt: sql.NullTime{
				Time:  job.StartTime.AsTime(),
				Valid: true,
			},
			WorkflowID: wID,
		}
		err := ingestsvc.CreateTask(ctx, &task)
		if err != nil {
			telemetry.RecordError(span, err)
			return err
		}
	}

	return nil
}
