// Example:
//
//  1. Make changes to schema files (internal/storage/persistence/ent/schema),
//  2. Re-generate (make gen-ent),
//  3. Run:
//     $ go run ./cmd/migrate/ \
//     --config="./enduro.toml" \
//     --dsn="mysql://enduro:enduro123@tcp(localhost:3306)/enduro_storage" \
//     --name="init" \
//     --path="./internal/storage/persistence/migrations"
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"ariga.io/atlas/sql/sqltool"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/go-sql-driver/mysql"
	"github.com/spf13/pflag"

	"github.com/artefactual-sdps/enduro/internal/config"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/migrate"
)

func main() {
	p := pflag.NewFlagSet("migrate", pflag.ExitOnError)
	p.String("config", "", "Configuration file")
	p.String("dsn", "", "MySQL DSN")
	p.String("path", "", "Migration directory")
	p.String("name", "changes", "Migration name")
	_ = p.Parse(os.Args[1:])

	path, _ := p.GetString("path")
	if path == "" {
		wd, err := os.Getwd()
		if err != nil {
			os.Exit(1)
		}
		// Guessing that running it from the root folder.
		path = filepath.Join(wd, "internal/storage/persistence/migrations")
	}

	var cfg config.Configuration
	configFile, _ := p.GetString("config")
	_, _, err := config.Read(&cfg, configFile)
	if err != nil {
		fmt.Printf("Failed to read configuration: %v\n", err)
		os.Exit(1)
	}

	DSN := cfg.Storage.Database.DSN
	flagDSN, _ := p.GetString("dsn")
	if flagDSN != "" {
		DSN = flagDSN
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
	}

	// Generate migrations using Atlas support for TiDB (note the Ent dialect option passed above).
	name, _ := p.GetString("name")
	err = migrate.NamedDiff(ctx, entDSN, name, opts...)
	if err != nil {
		log.Fatalf("failed generating migration file: %v", err)
	}
}
