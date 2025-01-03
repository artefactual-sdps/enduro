package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

const (
	// How many records we want to generate.
	datasetSize = 105

	// Size of each batch that we write to the CSV.
	batchSize = 10
)

type pkg struct {
	ID          string
	Name        string
	WorkflowID  string
	RunID       uuid.UUID
	AIPID       uuid.UUID
	LocationID  uuid.UUID
	Status      string
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
}

func (p pkg) toSlice() []any {
	return []any{
		p.ID,
		p.Name,
		p.WorkflowID,
		p.RunID,
		p.AIPID,
		p.LocationID,
		p.Status,
		p.CreatedAt,
		p.StartedAt,
		p.CompletedAt,
	}
}

func main() {
	var db *sql.DB

	cfg := mysql.Config{
		User:      "enduro",
		Passwd:    "enduro123",
		Net:       "tcp",
		Addr:      "127.0.0.1:3306",
		DBName:    "enduro",
		ParseTime: true,
	}

	// Get a database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")

	if _, err := db.Exec("DELETE FROM package;"); err != nil {
		log.Fatalf("Delete failed: %v", err)
	}

	batch := make([]pkg, 0, batchSize)
	var i int
	for i = range datasetSize {
		batch = append(batch, gen(i+1))

		if i > 0 && i%batchSize == 0 {
			if err := insertPkgs(db, batch); err != nil {
				log.Fatalf("Insert failed: %v", err)
			}
			batch = make([]pkg, 0, batchSize)
		}
	}

	if len(batch) > 0 {
		if err := insertPkgs(db, batch); err != nil {
			log.Fatalf("Insert failed: %v", err)
		}
	}

	fmt.Printf("%d packages inserted!\n", i+1)
}

func id() string {
	return uuid.New().String()
}

func gen(i int) pkg {
	const doneStatus string = "2"
	return pkg{
		ID:          strconv.Itoa(i),
		Name:        fmt.Sprintf("DPJ-SIP-%s.tar", id()),
		WorkflowID:  fmt.Sprintf("processing-workflow-%s", id()),
		RunID:       uuid.New(),
		AIPID:       uuid.New(),
		LocationID:  uuid.New(),
		Status:      doneStatus,
		CreatedAt:   time.Date(2019, 11, 21, 17, 36, 10, 0, time.UTC),
		StartedAt:   time.Date(2019, 11, 21, 17, 36, 11, 0, time.UTC),
		CompletedAt: time.Date(2019, 11, 21, 17, 42, 12, 0, time.UTC),
	}
}

func insertPkgs(db *sql.DB, pkgs []pkg) error {
	args := make([]any, 0, 10*len(pkgs))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	q := "INSERT INTO package VALUES "
	t := "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"

	for _, p := range pkgs {
		if p.ID == "" {
			break
		}

		q += t
		args = append(args, p.toSlice()...)
	}

	// Trim final comma.
	q = q[0 : len(q)-1]
	q += ";"

	_, err := db.ExecContext(ctx, q, args...)
	if err != nil {
		return err
	}

	return nil
}
