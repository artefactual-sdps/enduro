# Database migrations

Enduro uses [Atlas] to create and apply changes to the database schema.
Migrations for the "enduro" (ingest) database and "enduro_storage" database are
created and applied separately.

The database migration files use MySQL syntax, so they can only be applied to a
MySQL database.

## Create a migration database

You will need a completely empty migration database to generating a migration
file for either the enduro or enduro_storage database. The migrate command will
fail if the designated migration database contains any data. The migration
database name can have any valid name, as long as it is unique. We've used the
name "enduro_migrate" for the migration database in the examples below.

The migrate command deletes all data from the migration database after
generating a migration, so the same database can be used to generate multiple migrations.

### Start the Tilt development environment

```bash
tilt up
```

Wait for all the containers to complete their start-up before proceeding.

### Connect to MySQL with the mysql client

```bash
make db
```

### Create a new migration database and grant user permissions

```sql
create database enduro_migrate;
grant all on enduro_migrate to 'root'@'localhost';
```

### Exit the mysql client session

```sql
exit;
```

## Generate a new "enduro" database migration

### Update the "enduro" ent schema

Make the required schema changes to the appropriate file in the
`internal/persistence/ent/schema/` directory, then regenerate the [Ent]
schema:

```bash
make gen-ent
```

### Generate the "enduro" migration file

```bash
go run ./cmd/migrate/ \
--db="ingest" \
--dsn="mysql://root:root123@tcp(localhost:3306)/enduro_migrate" \
--path="./internal/db/migrations" \
--name="add_sip_note_column"
```

The `--name` option should summarize the changes made to the schema. Have a look
at the migration files in the `internal/db/migrations` directory for examples of
migration names.

If you need to change the name of a migration file you can manually change the
filename then regenerate the [Atlas] hash file (see below), or you can revert
all the changes to your local copy of Enduro and start again.

## Generate a new "enduro_storage" database migration

### Update the "storage" ent schema

Make the required schema changes to the appropriate file in the
`internal/storage/persistence/ent/schema/` directory, then regenerate the [Ent]
schema:

```bash
make gen-ent
```

### Generate the "storage" migration file

```bash
go run ./cmd/migrate/ \
--db="storage" \
--dsn="mysql://root:root123@tcp(localhost:3306)/enduro_migrate" \
--path="./internal/storage/persistence/migrations" \
--name="add_aip_note_column"
```

See the "enduro" migration generation instructions above for instructions on
naming the migration.

## Modifying a migration file

You will occasionally need to make changes to a generated migration file, such
as changing the file name, or adding additional queries to the file to modify
the database data (e.g. add a default value for a column, map an existing value
to a new value). After making changes to a migration file, you must regenerate
the `atlas.sum` file hashes for the migration to be recognized and applied.

### Regenerating the atlas.sum file hashes

Enduro has a make directive to regenerate the migration hashes for both the
"enduro" and "enduro_storage" databases.

```bash
make atlas-hash
```

## Apply the migrations

### When the enduro worker starts

The "enduro" and "storage" migrations will be applied to the respective
databases when the enduro worker starts, if the database schema is not already
at the latest version. You should see log messages about the migration in the
container start-up output.

```log
...
2025-04-23T23:13:43.263Z	V(2)	enduro.migrate	db/db.go:36	20250319035720/u rename_workflow_column (2.745526037s)
2025-04-23T23:13:43.278Z	V(2)	enduro.migrate	db/db.go:36	20250402173826/u remove_workflow_type (2.760657022s)
2025-04-23T23:13:43.363Z	V(2)	enduro.migrate	db/db.go:36	20250423213837/u add_sip_note_column (2.845041381s)
```

If no migration logs are output then the databases have the latest schema, or
and error occurred. If an error occurred applying the migrations, you should see
information about the error in the logs.

### When the Tilt "Flush" script runs

The Tilt "Flush" script can be run by clicking the trash can icon at the top
right of the Tilt dashboard. Running the "Flush" script will delete both the
"enduro" and "enduro_storage" databases, recreate them, and apply all
migrations. You should see the migrations being applied in the logs of the
"enduro" container console as shown above.

If you make changes to the migration files after starting the "enduro" Tilt
container, you will need to update the container to sync the changes before
running the "Flush" script.

## Troubleshooting

### My migration is not being applied

If you new migration is not being applied when the enduro worker starts or you
run the "Flush" script double check the following:

- All of the migration files in the "migrations" directory have an ".up.sql"
  extension. If any migration file doesn't have the correct extension the
  migration will stop at that file, and any subsequent migration files will be
  ignored.
- The name of the migration file in the "atlas.sum" file matches the name of the
  actual migration file.
- You've run `make atlas-hash` to update the "atlas.sum" file hashes.
- You've updated the "enduro" worker container after making any changes to the
  migration files.

[Atlas]: https://atlasgo.io/
[Ent]: https://entgo.io
