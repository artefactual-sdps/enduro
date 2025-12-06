package client_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"entgo.io/ent/dialect"
)

// countingDriver wraps a dialect.Driver to record executed INSERT statements.
type countingDriver struct {
	dialect.Driver
	execs  []string
	begins int
}

type countingTx struct {
	dialect.Tx
	parent *countingDriver
}

func (tx *countingTx) Exec(ctx context.Context, query string, args, v any) error {
	if strings.Contains(strings.ToUpper(query), "INSERT") {
		tx.parent.execs = append(tx.parent.execs, query)
	}

	return tx.Tx.Exec(ctx, query, args, v)
}

func (tx *countingTx) Query(ctx context.Context, query string, args, v any) error {
	if strings.Contains(strings.ToUpper(query), "INSERT") {
		tx.parent.execs = append(tx.parent.execs, query)
	}

	return tx.Tx.Query(ctx, query, args, v)
}

func (d *countingDriver) BeginTx(ctx context.Context, opts *sql.TxOptions) (dialect.Tx, error) {
	d.begins++
	if drv, ok := d.Driver.(interface {
		BeginTx(context.Context, *sql.TxOptions) (dialect.Tx, error)
	}); ok {
		tx, err := drv.BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		return &countingTx{Tx: tx, parent: d}, nil
	}

	return nil, fmt.Errorf("BeginTx not supported")
}

func (d *countingDriver) Exec(ctx context.Context, query string, args, v any) error {
	if strings.Contains(strings.ToUpper(query), "INSERT") {
		d.execs = append(d.execs, query)
	}

	return d.Driver.Exec(ctx, query, args, v)
}

func (d *countingDriver) reset() {
	d.execs = nil
	d.begins = 0
}
