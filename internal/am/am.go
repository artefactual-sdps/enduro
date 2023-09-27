package am

import (
	context "context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/artefactual-sdps/enduro/internal/package_"
	"github.com/go-logr/logr"
	"github.com/oklog/run"
	"go.artefactual.dev/amclient"
	temporalsdk_activity "go.temporal.io/sdk/activity"
)

const CreateAIPActivityName = "create-aip-activity"

type CreateAIPActivity struct {
	logger logr.Logger
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
				client := http.Client{}
				childCtx, cancel := context.WithTimeout(ctx, time.Second*10)
				defer cancel()

				c := amclient.NewClient(&client, a.cfg.Address, a.cfg.User, a.cfg.Key)

				req, err := c.NewRequest(
					childCtx,
					"Get",
					fmt.Sprintf("file://%s", opts.Path),
					&Processing{
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
				)
				if err != nil {
					return err
				}

				resp, err := c.Do(
					ctx,
					req,
					nil,
				)
				if err != nil {
					return err
				}
				result.UUID = resp.Body.Read()
				for {
					err := amclient.CheckResponse(resp.Response)
					if err != nil {
						return errors.New("package failed or rejected")
					}

					if resp.StatusCode == am.PACKAGE_STATUS_PROCESSING {
						continue
					}

					err = savePreservationTasks(ctx, c.Jobs, a.pkgsvc, opts.PreservationActionID)
					if err != nil {
						return err
					}

					result.Path = fmt.Sprintf("%s/completed/%s-%s.7z", a.cfg.ShareDir, opts.Name, result.UUID)
					a.logger.Info("We have run Archivematica successfully", "path", result.Path)

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

func savePreservationTasks(ctx context.Context, jobs []*transferservice.Job, pkgsvc package_.Service, paID uint) error {
	jobStatusToPreservationTaskStatus := map[transferservice.Job_Status]package_.PreservationTaskStatus{
		Job_STATUS_UNSPECIFIED: package_.TaskStatusUnspecified,
		Job_STATUS_COMPLETE:    package_.TaskStatusDone,
		Job_STATUS_PROCESSING:  package_.TaskStatusInProgress,
		Job_STATUS_FAILED:      package_.TaskStatusError,
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
