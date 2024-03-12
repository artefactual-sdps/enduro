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
	PackageUpdater func(*datatypes.Package) (*datatypes.Package, error)
)

type Service interface {
	CreatePackage(context.Context, *datatypes.Package) (*datatypes.Package, error)
	UpdatePackage(context.Context, uint, PackageUpdater) (*datatypes.Package, error)
}
