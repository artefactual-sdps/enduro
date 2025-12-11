package workflow

import (
	"fmt"

	temporalsdk_workflow "go.temporal.io/sdk/workflow"
)

const (
	A3mSemaphoreWorkflowName  = "a3m-semaphore-workflow"
	A3mAcquireSignalName      = "a3m-acquire"
	A3mReleaseSignalName      = "a3m-release"
	A3mSemaphoreWorkflowID    = "a3m-semaphore"
	A3mSemaphoreQueryName     = "a3m-semaphore-state"
	A3mSemaphoreMaxConcurrent = 1
)

type A3mSemaphoreAcquireSignal struct {
	WorkflowID string
	RunID      string
}

type A3mSemaphoreReleaseSignal struct {
	WorkflowID string
	RunID      string
}

type A3mSemaphoreWorkflow struct{}

func NewA3mSemaphoreWorkflow() *A3mSemaphoreWorkflow {
	return &A3mSemaphoreWorkflow{}
}

// Execute runs a long-lived workflow that manages concurrent access to a3m.
func (w *A3mSemaphoreWorkflow) Execute(ctx temporalsdk_workflow.Context) error {
	logger := temporalsdk_workflow.GetLogger(ctx)

	// Queue of workflows waiting for access.
	queue := []A3mSemaphoreAcquireSignal{}

	// Current workflows that have acquired access.
	acquired := make(map[string]bool)

	// Set up query handler to check semaphore state.
	err := temporalsdk_workflow.SetQueryHandler(ctx, A3mSemaphoreQueryName, func() (int, int) {
		return len(acquired), len(queue)
	})
	if err != nil {
		return fmt.Errorf("failed to set query handler: %w", err)
	}

	acquireChan := temporalsdk_workflow.GetSignalChannel(ctx, A3mAcquireSignalName)
	releaseChan := temporalsdk_workflow.GetSignalChannel(ctx, A3mReleaseSignalName)

	for {
		selector := temporalsdk_workflow.NewSelector(ctx)

		// Always listen for acquire signals.
		selector.AddReceive(acquireChan, func(c temporalsdk_workflow.ReceiveChannel, more bool) {
			var signal A3mSemaphoreAcquireSignal
			c.Receive(ctx, &signal)

			key := fmt.Sprintf("%s/%s", signal.WorkflowID, signal.RunID)

			// If already acquired, ignore duplicate request.
			if acquired[key] {
				logger.Info("Workflow already has access", "workflowID", signal.WorkflowID)
				return
			}

			// If we have capacity, grant immediately.
			if len(acquired) < A3mSemaphoreMaxConcurrent {
				acquired[key] = true
				logger.Info("Access granted", "workflowID", signal.WorkflowID, "active", len(acquired))

				// Signal the workflow that it can proceed.
				_ = temporalsdk_workflow.SignalExternalWorkflow(
					ctx,
					signal.WorkflowID,
					signal.RunID,
					"a3m-ready",
					nil,
				).Get(ctx, nil)
			} else {
				// Otherwise, add to queue.
				queue = append(queue, signal)
				logger.Info("Added to queue", "workflowID", signal.WorkflowID, "queueSize", len(queue))
			}
		})

		// Always listen for release signals.
		selector.AddReceive(releaseChan, func(c temporalsdk_workflow.ReceiveChannel, more bool) {
			var signal A3mSemaphoreReleaseSignal
			c.Receive(ctx, &signal)

			key := fmt.Sprintf("%s/%s", signal.WorkflowID, signal.RunID)

			if !acquired[key] {
				logger.Warn("Workflow tried to release without acquiring", "workflowID", signal.WorkflowID)
				return
			}

			delete(acquired, key)
			logger.Info("Access released", "workflowID", signal.WorkflowID, "active", len(acquired))

			// Grant access to next workflow in queue if any.
			if len(queue) > 0 {
				next := queue[0]
				queue = queue[1:]

				nextKey := fmt.Sprintf("%s/%s", next.WorkflowID, next.RunID)
				acquired[nextKey] = true
				logger.Info("Access granted from queue", "workflowID", next.WorkflowID, "active", len(acquired))

				// Signal the workflow that it can proceed.
				_ = temporalsdk_workflow.SignalExternalWorkflow(
					ctx,
					next.WorkflowID,
					next.RunID,
					"a3m-ready",
					nil,
				).Get(ctx, nil)
			}
		})

		selector.Select(ctx)
	}
}
