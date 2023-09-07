package entclient

import (
	"fmt"

	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
)

func rollback(tx *db.Tx, err error) error {
	rerr := tx.Rollback()
	if rerr == nil {
		return err
	}

	return fmt.Errorf("%w: failed transaction rollback: %v", persistence.ErrInternal, err)
}
