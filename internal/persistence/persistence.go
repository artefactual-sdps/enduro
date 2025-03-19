package persistence

import (
	"context"
	"errors"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
)

var (
	// ErrNotFound is the error returned if a resource cannot be found.
	ErrNotFound = errors.New("not found error")

	// ErrNotValid is the error returned if the data provided is invalid.
	ErrNotValid = errors.New("invalid data error")

	// ErrInternal is the error returned if an internal error occurred.
	ErrInternal = errors.New("internal error")
)

type (
	SIPUpdater  func(*datatypes.SIP) (*datatypes.SIP, error)
	TaskUpdater func(*datatypes.Task) (*datatypes.Task, error)
)

type Service interface {
	// CreateSIP persists the given SIP to the data store then updates
	// the SIP from the data store, adding auto-generated data
	// (e.g. ID, CreatedAt).
	CreateSIP(context.Context, *datatypes.SIP) error
	UpdateSIP(context.Context, int, SIPUpdater) (*datatypes.SIP, error)
	ListSIPs(context.Context, *SIPFilter) ([]*datatypes.SIP, *Page, error)

	CreateWorkflow(context.Context, *datatypes.Workflow) error

	CreateTask(context.Context, *datatypes.Task) error
	UpdateTask(ctx context.Context, id int, updater TaskUpdater) (*datatypes.Task, error)
}
