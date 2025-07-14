package entclient

import (
	"fmt"

	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
)

func newDBError(err error) error {
	if err == nil {
		return nil
	}

	var pErr error
	switch {
	case db.IsNotFound(err):
		pErr = persistence.ErrNotFound
	case db.IsConstraintError(err):
		pErr = persistence.ErrNotValid
	case db.IsValidationError(err):
		pErr = persistence.ErrNotValid
	case db.IsNotLoaded(err):
		pErr = persistence.ErrInternal
	case db.IsNotSingular(err):
		pErr = persistence.ErrInternal
	default:
		pErr = persistence.ErrInternal
	}

	return fmt.Errorf("%w: %s", pErr, err)
}

func newDBErrorWithDetails(err error, details string) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%w: %s", newDBError(err), details)
}

func newRequiredFieldError(field string) error {
	return fmt.Errorf("%w: field %q is required", persistence.ErrNotValid, field)
}

func newInvalidFieldError(field, value string) error {
	return fmt.Errorf("%w: field %q is invalid %q", persistence.ErrNotValid, field, value)
}

func newUpdaterError(err error) error {
	return fmt.Errorf("%w: updater error: %v", persistence.ErrNotValid, err)
}
