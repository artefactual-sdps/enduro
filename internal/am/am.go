package am

import (
	context "context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/artefactual-sdps/enduro/internal/package_"
	"github.com/artefactual-sdps/enduro/internal/temporal"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
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

				// Start transfer
				payload, resp, err := c.Package.Create(childCtx, &amclient.PackageCreateRequest{
					Name: opts.Name,
					Type: "standard",
					Path: opts.Path,
					// ProcessingConfig:
					AutoApprove: true,
					Accession:   uuid.New().String(),
				})
				if err != nil {
					if resp != nil {
						switch {
						case resp.StatusCode == http.StatusForbidden:
							return temporal.NonRetryableError(fmt.Errorf("authentication error: %v", err))
						}
					}
					return err
				}

				result.UUID = payload.ID
				for {
					err := amclient.CheckResponse(resp.Response)
					if err != nil {
						return errors.New("package failed or rejected")
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

func savePreservationTasks(ctx context.Context, jobs amclient.JobsService, pkgsvc package_.Service, paID uint) error {
	jobStatusToPreservationTaskStatus := map[amclient.JobStatus]package_.PreservationTaskStatus{
		amclient.JobStatusUnknown: package_.TaskStatusUnspecified,
		// amclient.JobStatusUserInput: [TODO: this may be required]
		amclient.JobStatusComplete:   package_.TaskStatusDone,
		amclient.JobStatusProcessing: package_.TaskStatusInProgress,
		amclient.JobStatusFailed:     package_.TaskStatusError,
	}

	js, _, err := jobs.List(ctx)
	for _, job := range js {
		pt := package_.PreservationTask{
			TaskID:               job.ID,
			Name:                 job.Name,
			Status:               jobStatusToPreservationTaskStatus[job.Status],
			PreservationActionID: paID,
		}
		// TODO: We probably should get the startedAt time from the Job.
		pt.StartedAt.Time = time.Now()
		err := pkgsvc.CreatePreservationTask(ctx, &pt)
		if err != nil {
			return err
		}
	}

	return nil
}
