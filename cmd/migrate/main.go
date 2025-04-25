// Example (ingest):
//
//  1. Make changes to schema files (internal/persistence/ent/schema),
//  2. Re-generate (make gen-ent),
//  3. Use an empty MySQL database,
//  4. Run:
//     $ go run ./cmd/migrate/ \
//     --db="ingest" \
//     --dsn="mysql://root:root123@tcp(localhost:3306)/enduro_migrate" \
//     --path="./internal/db/migrations" \
//     --name="changes"
//
// Example (storage):
//
//  1. Make changes to schema files (internal/storage/persistence/ent/schema),
//  2. Re-generate (make gen-ent),
//  3. Use an empty MySQL database,
//  4. Run:
//     $ go run ./cmd/migrate/ \
//     --db="storage" \
//     --dsn="mysql://root:root123@tcp(localhost:3306)/enduro_migrate" \
//     --path="./internal/storage/persistence/migrations" \
//     --name="changes"
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/sqltool"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/go-sql-driver/mysql"
	"github.com/spf13/pflag"

	ingest_migrate "github.com/artefactual-sdps/enduro/internal/persistence/ent/db/migrate"
	storage_migrate "github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/migrate"
)

func main() {
	p := pflag.NewFlagSet("migrate", pflag.ExitOnError)
	p.String("db", "", "Enduro database ('ingest' or 'storage')")
	p.String("dsn", "", "MySQL DSN")
	p.String("path", "", "Migration directory")
	p.String("name", "changes", "Migration name")
	_ = p.Parse(os.Args[1:])

	db, _ := p.GetString("db")
	if db == "" {
		fmt.Printf("--db flag is missing")
		os.Exit(1)
	}
	if db != "ingest" && db != "storage" {
		fmt.Printf("--db flag has an unexpected value (use 'ingest' or 'storage')")
		os.Exit(1)
	}

	DSN, _ := p.GetString("dsn")
	if DSN == "" {
		fmt.Printf("--dsn flag is missing")
		os.Exit(1)
	}

	path, _ := p.GetString("path")
	if path == "" {
		fmt.Printf("--path flag is missing")
		os.Exit(1)
	}

	// MySQL's DSN format is not accepted by Ent, convert as needed (remove Net).
	DSN = strings.TrimPrefix(DSN, "mysql://")
	DSNConfig, err := mysql.ParseDSN(DSN)
	if err != nil {
		fmt.Printf("Failed to parse MySQL DSN: %v\n", err)
		os.Exit(1)
	}
	entDSN := fmt.Sprintf("%s://%s:%s@%s/%s",
		"mysql",
		DSNConfig.User,
		DSNConfig.Passwd,
		DSNConfig.Addr,
		DSNConfig.DBName,
	)

	ctx := context.Background()

	// Create a local migration directory able to understand golang-migrate migration files for replay.
	dir, err := sqltool.NewGolangMigrateDir(path)
	if err != nil {
		log.Fatalf("failed creating atlas migration directory: %v", err)
	}

	// Write migration diff.
	opts := []schema.MigrateOption{
		schema.WithDir(dir),                         // provide migration directory
		schema.WithMigrationMode(schema.ModeReplay), // provide migration mode
		schema.WithDialect(dialect.MySQL),           // Ent dialect to use
		schema.WithFormatter(migrate.DefaultFormatter),
		schema.WithDropIndex(true),
		schema.WithDropColumn(true),
	}

	// Append ".up" to the migration name if not already present, as a
	// workaround for the fact that Atlas omits ".up" from the ".up.sql" suffix
	// that is required when the migration is applied.
	name, _ := p.GetString("name")
	if !strings.HasSuffix(name, ".up") {
		name = fmt.Sprintf("%s.up", name)
	}

	// Generate migrations using Atlas support for MySQL (note the Ent dialect option passed above).
	if db == "ingest" {
		err = ingest_migrate.NamedDiff(ctx, entDSN, name, opts...)
	} else {
		err = storage_migrate.NamedDiff(ctx, entDSN, name, opts...)
	}
	if err != nil {
		log.Fatalf("failed generating migration file: %v", err)
	}
}
