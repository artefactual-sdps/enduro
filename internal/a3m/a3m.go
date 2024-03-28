package a3m

import (
	context "context"
	"errors"
	"fmt"
	"time"

	"buf.build/gen/go/artefactual/a3m/grpc/go/a3m/api/transferservice/v1beta1/transferservicev1beta1grpc"
	transferservice "buf.build/gen/go/artefactual/a3m/protocolbuffers/go/a3m/api/transferservice/v1beta1"
	"github.com/go-logr/logr"
	"github.com/oklog/run"
	"go.opentelemetry.io/otel/trace"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	"google.golang.org/grpc"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/package_"
	"github.com/artefactual-sdps/enduro/internal/telemetry"
)

const CreateAIPActivityName = "create-aip-activity"

type CreateAIPActivity struct {
	logger logr.Logger
	tracer trace.Tracer
	client transferservicev1beta1grpc.TransferServiceClient
	cfg    *Config
	pkgsvc package_.Service
}

type CreateAIPActivityParams struct {
	Name                 string
	Path                 string
	PreservationActionID uint
}

type CreateAIPActivityResult struct {
	Path string
	UUID string
}

func NewCreateAIPActivity(
	logger logr.Logger,
	tracer trace.Tracer,
	client transferservicev1beta1grpc.TransferServiceClient,
	cfg *Config,
	pkgsvc package_.Service,
) *CreateAIPActivity {
	return &CreateAIPActivity{
		logger: logger,
		tracer: tracer,
		client: client,
		cfg:    cfg,
		pkgsvc: pkgsvc,
	}
}

func (a *CreateAIPActivity) Execute(
	ctx context.Context,
	opts *CreateAIPActivityParams,
) (*CreateAIPActivityResult, error) {
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
							AipCompressionLevel:                          int32(a.cfg.AipCompressionLevel),
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
						continue
					}

					err = savePreservationTasks(ctx, a.tracer, readResp.Jobs, a.pkgsvc, opts.PreservationActionID)
					if err != nil {
						return err
					}

					if readResp.Status == transferservice.PackageStatus_PACKAGE_STATUS_FAILED ||
						readResp.Status == transferservice.PackageStatus_PACKAGE_STATUS_REJECTED {
						return errors.New("package failed or rejected")
					}

					result.Path = fmt.Sprintf("%s/completed/%s-%s.7z", a.cfg.ShareDir, opts.Name, result.UUID)
					a.logger.Info("We have run a3m successfully", "path", result.Path)

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

func savePreservationTasks(
	ctx context.Context,
	tracer trace.Tracer,
	jobs []*transferservice.Job,
	pkgsvc package_.Service,
	paID uint,
) error {
	ctx, span := tracer.Start(ctx, "savePreservationTasks")
	defer span.End()

	jobStatusToPreservationTaskStatus := map[transferservice.Job_Status]enums.PreservationTaskStatus{
		transferservice.Job_STATUS_UNSPECIFIED: enums.PreservationTaskStatusUnspecified,
		transferservice.Job_STATUS_COMPLETE:    enums.PreservationTaskStatusDone,
		transferservice.Job_STATUS_PROCESSING:  enums.PreservationTaskStatusInProgress,
		transferservice.Job_STATUS_FAILED:      enums.PreservationTaskStatusError,
	}

	for _, job := range jobs {
		pt := datatypes.PreservationTask{
			TaskID:               job.Id,
			Name:                 job.Name,
			Status:               jobStatusToPreservationTaskStatus[job.Status],
			PreservationActionID: paID,
		}
		pt.StartedAt.Time = job.StartTime.AsTime()
		err := pkgsvc.CreatePreservationTask(ctx, &pt)
		if err != nil {
			telemetry.RecordError(span, err)
			return err
		}
	}

	return nil
}
