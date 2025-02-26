package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/run"
	"go.artefactual.dev/tools/temporal"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	goa "goa.design/goa/v3/pkg"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage"
)

type CreateStorageAIPActivity struct {
	client storage.Client
}

type CreateStorageAIPActivityParams struct {
	Name       string
	AIPID      string
	ObjectKey  string
	Status     string
	LocationID *uuid.UUID
}

type CreateStorageAIPActivityResult struct {
	CreatedAt string
}

func NewCreateStorageAIPActivity(client storage.Client) *CreateStorageAIPActivity {
	return &CreateStorageAIPActivity{client: client}
}

func (a *CreateStorageAIPActivity) Execute(
	ctx context.Context,
	params *CreateStorageAIPActivityParams,
) (*CreateStorageAIPActivityResult, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(1).Info("Executing CreateStorageSIPActivity", "params", params)

	payload := goastorage.CreateAipPayload{
		UUID:       params.AIPID,
		Name:       params.Name,
		Status:     params.Status,
		ObjectKey:  params.ObjectKey,
		LocationID: params.LocationID,
	}

	aip, err := a.client.CreateAip(ctx, &payload)
	if err != nil {
		if errors.Is(err, goastorage.Unauthorized("Unauthorized")) {
			return nil, temporal.NewNonRetryableError(
				fmt.Errorf("%s: %v", CreateStorageAIPActivityName, err),
			)
		}

		if serr, ok := err.(*goa.ServiceError); ok {
			if serr.Name == "not_valid" {
				return nil, temporal.NewNonRetryableError(
					fmt.Errorf("%s: %v", CreateStorageAIPActivityName, err),
				)
			}
		}

		return nil, fmt.Errorf("%s: %v", CreateStorageAIPActivityName, err)
	}

	return &CreateStorageAIPActivityResult{CreatedAt: aip.CreatedAt}, nil
}

type MoveToPermanentStorageActivityParams struct {
	AIPID      string
	LocationID uuid.UUID
}

type MoveToPermanentStorageActivityResult struct{}

type MoveToPermanentStorageActivity struct {
	storageClient *goastorage.Client
}

func NewMoveToPermanentStorageActivity(storageClient *goastorage.Client) *MoveToPermanentStorageActivity {
	return &MoveToPermanentStorageActivity{
		storageClient: storageClient,
	}
}

func (a *MoveToPermanentStorageActivity) Execute(
	ctx context.Context,
	params *MoveToPermanentStorageActivityParams,
) (*MoveToPermanentStorageActivityResult, error) {
	childCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	err := a.storageClient.MoveAip(childCtx, &goastorage.MoveAipPayload{
		UUID:       params.AIPID,
		LocationID: params.LocationID,
	})

	return &MoveToPermanentStorageActivityResult{}, err
}

type PollMoveToPermanentStorageActivityParams struct {
	AIPID string
}

type PollMoveToPermanentStorageActivity struct {
	storageClient *goastorage.Client
}

type PollMoveToPermanentStorageActivityResult struct{}

func NewPollMoveToPermanentStorageActivity(storageClient *goastorage.Client) *PollMoveToPermanentStorageActivity {
	return &PollMoveToPermanentStorageActivity{
		storageClient: storageClient,
	}
}

func (a *PollMoveToPermanentStorageActivity) Execute(
	ctx context.Context,
	params *PollMoveToPermanentStorageActivityParams,
) (*PollMoveToPermanentStorageActivityResult, error) {
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
				childCtx, cancel := context.WithTimeout(ctx, time.Second*5)
				defer cancel()

				for {
					res, err := a.storageClient.MoveAipStatus(childCtx, &goastorage.MoveAipStatusPayload{
						UUID: params.AIPID,
					})
					if err != nil {
						return err
					}
					if res.Done {
						break
					}
				}

				return nil
			},
			func(error) {},
		)
	}

	err := g.Run()
	return &PollMoveToPermanentStorageActivityResult{}, err
}

type RejectSIPActivityParams struct {
	AIPID string
}

type RejectSIPActivity struct {
	storageClient *goastorage.Client
}

type RejectSIPActivityResult struct{}

func NewRejectSIPActivity(storageClient *goastorage.Client) *RejectSIPActivity {
	return &RejectSIPActivity{
		storageClient: storageClient,
	}
}

func (a *RejectSIPActivity) Execute(
	ctx context.Context,
	params *RejectSIPActivityParams,
) (*RejectSIPActivityResult, error) {
	childCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	err := a.storageClient.RejectAip(childCtx, &goastorage.RejectAipPayload{
		UUID: params.AIPID,
	})

	return &RejectSIPActivityResult{}, err
}
