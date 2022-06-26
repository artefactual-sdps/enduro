package activities

import (
	"context"
	"time"

	"github.com/oklog/run"
	temporalsdk_activity "go.temporal.io/sdk/activity"

	goastorage "github.com/artefactual-labs/enduro/internal/api/gen/storage"
)

type MoveToPermanentStorageActivityParams struct {
	AIPID    string
	Location string
}

type MoveToPermanentStorageActivity struct {
	storageClient *goastorage.Client
}

func NewMoveToPermanentStorageActivity(storageClient *goastorage.Client) *MoveToPermanentStorageActivity {
	return &MoveToPermanentStorageActivity{
		storageClient: storageClient,
	}
}

func (a *MoveToPermanentStorageActivity) Execute(ctx context.Context, params *MoveToPermanentStorageActivityParams) error {
	childCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	err := a.storageClient.Move(childCtx, &goastorage.MovePayload{
		AipID:    params.AIPID,
		Location: params.Location,
	})

	return err
}

type PollMoveToPermanentStorageActivityParams struct {
	AIPID string
}

type PollMoveToPermanentStorageActivity struct {
	storageClient *goastorage.Client
}

func NewPollMoveToPermanentStorageActivity(storageClient *goastorage.Client) *PollMoveToPermanentStorageActivity {
	return &PollMoveToPermanentStorageActivity{
		storageClient: storageClient,
	}
}

func (a *PollMoveToPermanentStorageActivity) Execute(ctx context.Context, params *PollMoveToPermanentStorageActivityParams) error {
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
	return err
}

type RejectPackageActivityParams struct {
	AIPID string
}

type RejectPackageActivity struct {
	storageClient *goastorage.Client
}

func NewRejectPackageActivity(storageClient *goastorage.Client) *RejectPackageActivity {
	return &RejectPackageActivity{
		storageClient: storageClient,
	}
}

func (a *RejectPackageActivity) Execute(ctx context.Context, params *RejectPackageActivityParams) error {
	childCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	err := a.storageClient.Reject(childCtx, &goastorage.RejectPayload{
		AipID: params.AIPID,
	})

	return err
}
