package persistence

import (
	"context"
	"errors"
	"iter"

	"github.com/google/uuid"

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
	SIPUpdater   func(*datatypes.SIP) (*datatypes.SIP, error)
	TaskUpdater  func(*datatypes.Task) (*datatypes.Task, error)
	BatchUpdater func(*datatypes.Batch) (*datatypes.Batch, error)
)

// TaskSequence is a convenience alias for iter.Seq[*datatypes.Task]. Use
// helpers such as slices.Values to convert an existing slice into the iterator.
type TaskSequence = iter.Seq[*datatypes.Task]

type Service interface {
	// CreateSIP persists the given SIP to the data store then updates
	// the SIP from the data store, adding auto-generated data
	// (e.g. ID, CreatedAt).
	CreateSIP(context.Context, *datatypes.SIP) error
	UpdateSIP(context.Context, uuid.UUID, SIPUpdater) (*datatypes.SIP, error)
	DeleteSIP(context.Context, uuid.UUID) error
	ReadSIP(context.Context, uuid.UUID) (*datatypes.SIP, error)
	ListSIPs(context.Context, *SIPFilter) ([]*datatypes.SIP, *Page, error)

	CreateWorkflow(context.Context, *datatypes.Workflow) error

	CreateTask(context.Context, *datatypes.Task) error
	// CreateTasks persists all tasks yielded by the sequence. For very large
	// sequences, the transaction may remain open for an extended period,
	// potentially causing lock contention or timeout issues.
	//
	// Implementors must consume the sequence in batches and stop on the first
	// error. Implementors must insert all batches within a single transaction
	// for atomicity. On success, implementors must update yielded tasks in
	// place with generated fields (e.g. database IDs).
	CreateTasks(context.Context, TaskSequence) error
	UpdateTask(ctx context.Context, id int, updater TaskUpdater) (*datatypes.Task, error)

	// CreateUser persists a new user to the data store then updates the user
	// to add auto-generated data (e.g. ID, CreatedAt).
	CreateUser(context.Context, *datatypes.User) error

	// ReadUser retrieves a user by UUID.
	ReadUser(context.Context, uuid.UUID) (*datatypes.User, error)

	// ReadOIDCUser retrieves a user by OIDC issuer and subject.
	ReadOIDCUser(ctx context.Context, iss, sub string) (*datatypes.User, error)

	// ListUsers retrieves a list of users based on the provided filter.
	ListUsers(context.Context, *UserFilter) ([]*datatypes.User, *Page, error)

	CreateBatch(context.Context, *datatypes.Batch) error
	UpdateBatch(context.Context, uuid.UUID, BatchUpdater) (*datatypes.Batch, error)
	DeleteBatch(context.Context, uuid.UUID) error
	ReadBatch(context.Context, uuid.UUID) (*datatypes.Batch, error)
	ListBatches(context.Context, *BatchFilter) ([]*datatypes.Batch, *Page, error)
}
