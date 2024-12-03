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
	PackageUpdater  func(*datatypes.Package) (*datatypes.Package, error)
	PresTaskUpdater func(*datatypes.PreservationTask) (*datatypes.PreservationTask, error)
)

type Service interface {
	// CreatePackage persists the given Package to the data store then updates
	// the Package from the data store, adding auto-generated data
	// (e.g. ID, CreatedAt).
	CreatePackage(context.Context, *datatypes.Package) error
	UpdatePackage(context.Context, int, PackageUpdater) (*datatypes.Package, error)
	ListPackages(context.Context, *PackageFilter) ([]*datatypes.Package, *Page, error)

	CreatePreservationAction(context.Context, *datatypes.PreservationAction) error

	CreatePreservationTask(context.Context, *datatypes.PreservationTask) error
	CreatePreservationTasks(
		context.Context,
		func(yield func(*datatypes.PreservationTask) bool),
	) ([]*datatypes.PreservationTask, error)
	UpdatePreservationTask(ctx context.Context, id int, updater PresTaskUpdater) (*datatypes.PreservationTask, error)
}
