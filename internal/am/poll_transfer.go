package am

import (
	context "context"
	"errors"
	"net/http"
	"time"

	"github.com/artefactual-sdps/enduro/internal/temporal"

	"github.com/cenkalti/backoff/v3"
	temporalsdk_activity "go.temporal.io/sdk/activity"
)

const defaultMaxElapsedTime = time.Minute * 10

type PollTransferParams struct {
	PipelineName string
	TransferID   string
}

func PollTransfer(ctx context.Context, params *PollTransferParams) (string, error) {
	deadline := defaultMaxElapsedTime
	client := http.Client{}

	var sipID string
	lastRetryableError := time.Time{}
	var backoffStrategy backoff.BackOff = backoff.NewConstantBackOff(time.Second * 5)

	err = backoff.RetryNotify(
		func() (err error) {
			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()

			sipID, err = pipeline.TransferStatus(ctx, amc.Transfer, params.TransferID)

			// Abandon when we see a non-retryable error.
			if errors.Is(err, pipeline.ErrStatusNonRetryable) {
				return backoff.Permanent(temporal.NewNonRetryableError(err))
			}

			// Looking good, keep polling.
			if errors.Is(err, pipeline.ErrStatusInProgress) {
				lastRetryableError = time.Time{} // Reset.
				return err
			}

			if err != nil {
				logger.Error("Failed to look up Transfer status.", "error", err)
			}

			// Retry unless the deadline was exceeded.
			if lastRetryableError.IsZero() {
				lastRetryableError = clock.Now()
			} else if clock.Since(lastRetryableError) > deadline {
				return backoff.Permanent(temporal.NewNonRetryableError(err))
			}

			return err
		},
		backoffStrategy,
		func(err error, duration time.Duration) {
			temporalsdk_activity.RecordHeartbeat(ctx, err.Error())
		},
	)

	return sipID, err
}
