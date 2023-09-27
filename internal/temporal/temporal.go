package temporal

import (
	"fmt"

	temporalsdk_temporal "go.temporal.io/sdk/temporal"
)

const (
	GlobalTaskQueue    = "global"
	A3mWorkerTaskQueue = "a3m"
	AmWorkerTaskQueue  = "am"
)

func NonRetryableError(err error) error {
	return temporalsdk_temporal.NewNonRetryableApplicationError(
		fmt.Sprintf("non retryable error: %v", err.Error()), "", nil, nil,
	)
}

func ContinuePollingError() error {
	return temporalsdk_temporal.NewApplicationError("Continue polling", "polling", nil)
}
