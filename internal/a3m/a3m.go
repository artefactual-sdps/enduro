package a3m

import (
	context "context"
	"errors"
	"fmt"
	"time"

	a3m_transferservice "buf.build/gen/go/artefactual/a3m/protocolbuffers/go/a3m/api/transferservice/v1beta1"
	"github.com/go-logr/logr"
	"github.com/oklog/run"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	"google.golang.org/grpc"

	"github.com/artefactual-sdps/enduro/internal/package_"
)

const CreateAIPActivityName = "create-aip-activity"

type CreateAIPActivity struct {
	logger logr.Logger
	cfg    *Config
	pkgsvc package_.Service
}

type CreateAIPActivityParams struct {
	Path                 string
	PreservationActionID uint
}

type CreateAIPActivityResult struct {
	Path string
	UUID string
}

func NewCreateAIPActivity(logger logr.Logger, cfg *Config, pkgsvc package_.Service) *CreateAIPActivity {
	return &CreateAIPActivity{
		logger: logger,
		cfg:    cfg,
		pkgsvc: pkgsvc,
	}
}

func (a *CreateAIPActivity) Execute(ctx context.Context, opts *CreateAIPActivityParams) (*CreateAIPActivityResult, error) {
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
				childCtx, cancel := context.WithTimeout(ctx, time.Second*10)
				defer cancel()

				c, err := NewClient(childCtx, a.cfg.Address)
				if err != nil {
					return err
				}

				submitResp, err := c.TransferClient.Submit(
					ctx,
					&a3m_transferservice.SubmitRequest{
						Name: "enduro",
						Url:  fmt.Sprintf("file://%s", opts.Path),
						Config: &a3m_transferservice.ProcessingConfig{
							AssignUuidsToDirectories:                     true,
							ExamineContents:                              false,
							GenerateTransferStructureReport:              true,
							DocumentEmptyDirectories:                     true,
							ExtractPackages:                              true,
							DeletePackagesAfterExtraction:                false,
							IdentifyTransfer:                             true,
							IdentifySubmissionAndMetadata:                true,
							IdentifyBeforeNormalization:                  true,
							Normalize:                                    true,
							TranscribeFiles:                              true,
							PerformPolicyChecksOnOriginals:               true,
							PerformPolicyChecksOnPreservationDerivatives: true,
							AipCompressionLevel:                          1,
							AipCompressionAlgorithm:                      a3m_transferservice.ProcessingConfig_AIP_COMPRESSION_ALGORITHM_S7_BZIP2,
						},
					},
					grpc.WaitForReady(true),
				)
				if err != nil {
					return err
				}

				result.UUID = submitResp.Id

				for {
					readResp, err := c.TransferClient.Read(ctx, &a3m_transferservice.ReadRequest{Id: result.UUID})
					if err != nil {
						return err
					}

					if readResp.Status == a3m_transferservice.PackageStatus_PACKAGE_STATUS_PROCESSING {
						continue
					}

					err = savePreservationTasks(ctx, readResp.Jobs, a.pkgsvc, opts.PreservationActionID)
					if err != nil {
						return err
					}

					if readResp.Status == a3m_transferservice.PackageStatus_PACKAGE_STATUS_FAILED || readResp.Status == a3m_transferservice.PackageStatus_PACKAGE_STATUS_REJECTED {
						return errors.New("package failed or rejected")
					}

					result.Path = fmt.Sprintf("%s/completed/%s-%s.7z", a.cfg.ShareDir, "enduro", result.UUID)
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

func savePreservationTasks(ctx context.Context, jobs []*a3m_transferservice.Job, pkgsvc package_.Service, paID uint) error {
	jobStatusToPreservationTaskStatus := map[a3m_transferservice.Job_Status]package_.PreservationTaskStatus{
		a3m_transferservice.Job_STATUS_UNSPECIFIED: package_.TaskStatusUnspecified,
		a3m_transferservice.Job_STATUS_COMPLETE:    package_.TaskStatusDone,
		a3m_transferservice.Job_STATUS_PROCESSING:  package_.TaskStatusInProgress,
		a3m_transferservice.Job_STATUS_FAILED:      package_.TaskStatusError,
	}

	for _, job := range jobs {
		pt := package_.PreservationTask{
			TaskID:               job.Id,
			Name:                 job.Name,
			Status:               jobStatusToPreservationTaskStatus[job.Status],
			PreservationActionID: paID,
		}
		pt.StartedAt.Time = job.StartTime.AsTime()
		err := pkgsvc.CreatePreservationTask(ctx, &pt)
		if err != nil {
			return err
		}
	}

	return nil
}
