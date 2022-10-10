package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
)

// Wait up to five minutes is another process is already on it.
const lockTimeout = time.Minute * 5

// Connect returns the database handler which is safe for concurrent access.
func Connect(ds string) (db *sql.DB, err error) {
	config, err := mysqldriver.ParseDSN(ds)
	if err != nil {
		return nil, fmt.Errorf("error parsing dsn: %w (%s)", err, ds)
	}
	config.Collation = "utf8mb4_unicode_ci"
	config.Loc = time.UTC
	config.ParseTime = true
	config.MultiStatements = true
	config.Params = map[string]string{
		"time_zone": "UTC",
	}

	conn, err := mysqldriver.NewConnector(config)
	if err != nil {
		return nil, fmt.Errorf("error creating connector: %w", err)
	}

	db = sql.OpenDB(conn)

	// Set reasonable defaults in the built-in pool.
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	// Register Prometheus collector.
	c := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "src",
			Subsystem: "mysql_app_" + config.DBName,
			Name:      "open_connections",
			Help:      "Number of open connections to MySQL DB, as reported by mysql.DB.Stats()",
		},
		func() float64 {
			s := db.Stats()
			return float64(s.OpenConnections)
		},
	)
	prometheus.MustRegister(c)

	return db, nil
}

func MigrateEnduroDatabase(db *sql.DB) error {
	src, err := enduroSourceDriver()
	if err != nil {
		return fmt.Errorf("error loading migration sources: %v", err)
	}

	return up(db, src)
}

func MigrateEnduroStorageDatabase(db *sql.DB) error {
	src, err := persistence.SourceDriver()
	if err != nil {
		return fmt.Errorf("error loading migration sources: %v", err)
	}

	return up(db, src)
}

// up migrates the database.
func up(db *sql.DB, src source.Driver) error {
	m, err := newMigrate(db, src)
	if err != nil {
		return fmt.Errorf("error creating golang-migrate object: %v", err)
	}

	err = doMigrate(m)
	if err != nil {
		return fmt.Errorf("error during database migration: %v", err)
	}

	return nil
}

// newMigrate builds the golang-migrate object.
func newMigrate(db *sql.DB, src source.Driver) (*migrate.Migrate, error) {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return nil, fmt.Errorf("error creating migrate driver: %w", err)
	}

	m, err := migrate.NewWithInstance("source", src, "driver", driver)
	if err != nil {
		return nil, fmt.Errorf("error creating migrate instance: %w", err)
	}

	m.LockTimeout = lockTimeout

	return m, nil
}

func doMigrate(m *migrate.Migrate) error {
	err := m.Up()
	if err == nil || err == migrate.ErrNoChange {
		return nil
	}

	if os.IsNotExist(err) {
		_, dirty, verr := m.Version()
		if verr != nil {
			return verr
		}
		if dirty {
			return err
		}
		return nil
	}

	return err
}
