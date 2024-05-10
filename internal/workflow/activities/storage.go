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

type CreateStoragePackageActivity struct {
	client storage.Client
}

type CreateStoragePackageActivityParams struct {
	Name       string
	AIPID      string
	ObjectKey  string
	Status     string
	LocationID *uuid.UUID
}

type CreateStoragePackageActivityResult struct {
	CreatedAt string
}

func NewCreateStoragePackageActivity(client storage.Client) *CreateStoragePackageActivity {
	return &CreateStoragePackageActivity{client: client}
}

func (a *CreateStoragePackageActivity) Execute(
	ctx context.Context,
	params *CreateStoragePackageActivityParams,
) (*CreateStoragePackageActivityResult, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(1).Info("Executing CreateStoragePackageActivity", "params", params)

	payload := goastorage.CreatePayload{
		AipID:      params.AIPID,
		Name:       params.Name,
		Status:     params.Status,
		ObjectKey:  params.ObjectKey,
		LocationID: params.LocationID,
	}

	pkg, err := a.client.Create(ctx, &payload)
	if err != nil {
		if errors.Is(err, goastorage.Unauthorized("unauthorized")) {
			return nil, temporal.NewNonRetryableError(
				fmt.Errorf("%s: %v", CreateStoragePackageActivityName, err),
			)
		}

		if serr, ok := err.(*goa.ServiceError); ok {
			if serr.Name == "not_valid" {
				return nil, temporal.NewNonRetryableError(
					fmt.Errorf("%s: %v", CreateStoragePackageActivityName, err),
				)
			}
		}

		return nil, fmt.Errorf("%s: %v", CreateStoragePackageActivityName, err)
	}

	return &CreateStoragePackageActivityResult{CreatedAt: pkg.CreatedAt}, nil
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

	err := a.storageClient.Move(childCtx, &goastorage.MovePayload{
		AipID:      params.AIPID,
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
					res, err := a.storageClient.MoveStatus(childCtx, &goastorage.MoveStatusPayload{
						AipID: params.AIPID,
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

type RejectPackageActivityParams struct {
	AIPID string
}

type RejectPackageActivity struct {
	storageClient *goastorage.Client
}

type RejectPackageActivityResult struct{}

func NewRejectPackageActivity(storageClient *goastorage.Client) *RejectPackageActivity {
	return &RejectPackageActivity{
		storageClient: storageClient,
	}
}

func (a *RejectPackageActivity) Execute(
	ctx context.Context,
	params *RejectPackageActivityParams,
) (*RejectPackageActivityResult, error) {
	childCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	err := a.storageClient.Reject(childCtx, &goastorage.RejectPayload{
		AipID: params.AIPID,
	})

	return &RejectPackageActivityResult{}, err
}
