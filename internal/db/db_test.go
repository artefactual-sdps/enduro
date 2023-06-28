package db

import (
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"gotest.tools/v3/assert"
)

func TestConnect(t *testing.T) {
	t.Parallel()

	t.Run("Returns a SQLite connection", func(t *testing.T) {
		t.Parallel()

		dbfile := "/tmp/enduro-test.db"
		db, err := Connect("sqlite3", dbfile)
		assert.NilError(t, err)
		defer func() {
			db.Close()
			os.Remove(dbfile)
		}()

		_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS test (
	id INTEGER NOT NULL PRIMARY KEY,
	name VARCHAR(32)
);
`)
		assert.NilError(t, err)

		_, err = db.Exec(`INSERT INTO test VALUES (1, "Foo Bar");`)
		assert.NilError(t, err)

		var id int
		var name string
		r := db.QueryRow(`SELECT * FROM test;`)
		err = r.Scan(&id, &name)

		assert.NilError(t, err)
		assert.Equal(t, id, 1)
		assert.Equal(t, name, "Foo Bar")
	})
}

func TestMigrateEnduroDatabase(t *testing.T) {
	t.Parallel()

	t.Run("Error migrating SQLite db", func(t *testing.T) {
		t.Parallel()

		dbfile := "/tmp/enduro-migrate-test.db"
		db, err := Connect("sqlite3", dbfile)
		assert.NilError(t, err)
		defer func() {
			db.Close()
			os.Remove(dbfile)
		}()

		err = MigrateEnduroDatabase(db)
		assert.Error(t, err, "only MySQL migrations are supported")
	})
}
