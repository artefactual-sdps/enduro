package client

import (
	"fmt"

	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
)

func rollback(tx *db.Tx, err error) error {
	rerr := tx.Rollback()
	if rerr == nil {
		return err
	}

	return fmt.Errorf("failed transaction rollback: %v", err)
}
