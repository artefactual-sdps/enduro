package workflow

import (
	"github.com/go-logr/logr"
	temporalsdk_workflow "go.temporal.io/sdk/workflow"

	"github.com/artefactual-labs/enduro/internal/package_"
)

type MoveWorkflow struct {
	logger logr.Logger
	pkgsvc package_.Service
}

func NewMoveWorkflow(logger logr.Logger, pkgsvc package_.Service) *MoveWorkflow {
	return &MoveWorkflow{
		logger: logger,
		pkgsvc: pkgsvc,
	}
}

func (w *MoveWorkflow) Execute(ctx temporalsdk_workflow.Context, req *package_.MoveWorkflowRequest) error {
	return nil
}
