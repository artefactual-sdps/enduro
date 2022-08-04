// Code generated by ent, DO NOT EDIT.

package db

import (
	"context"
	"fmt"
	"log"

	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/migrate"

	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/pkg"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
)

// Client is the client that holds all ent builders.
type Client struct {
	config
	// Schema is the client for creating, migrating and dropping schema.
	Schema *migrate.Schema
	// Pkg is the client for interacting with the Pkg builders.
	Pkg *PkgClient
}

// NewClient creates a new client configured with the given options.
func NewClient(opts ...Option) *Client {
	cfg := config{log: log.Println, hooks: &hooks{}}
	cfg.options(opts...)
	client := &Client{config: cfg}
	client.init()
	return client
}

func (c *Client) init() {
	c.Schema = migrate.NewSchema(c.driver)
	c.Pkg = NewPkgClient(c.config)
}

// Open opens a database/sql.DB specified by the driver name and
// the data source name, and returns a new client attached to it.
// Optional parameters can be added for configuring the client.
func Open(driverName, dataSourceName string, options ...Option) (*Client, error) {
	switch driverName {
	case dialect.MySQL, dialect.Postgres, dialect.SQLite:
		drv, err := sql.Open(driverName, dataSourceName)
		if err != nil {
			return nil, err
		}
		return NewClient(append(options, Driver(drv))...), nil
	default:
		return nil, fmt.Errorf("unsupported driver: %q", driverName)
	}
}

// Tx returns a new transactional client. The provided context
// is used until the transaction is committed or rolled back.
func (c *Client) Tx(ctx context.Context) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, fmt.Errorf("db: cannot start a transaction within a transaction")
	}
	tx, err := newTx(ctx, c.driver)
	if err != nil {
		return nil, fmt.Errorf("db: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = tx
	return &Tx{
		ctx:    ctx,
		config: cfg,
		Pkg:    NewPkgClient(cfg),
	}, nil
}

// BeginTx returns a transactional client with specified options.
func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, fmt.Errorf("ent: cannot start a transaction within a transaction")
	}
	tx, err := c.driver.(interface {
		BeginTx(context.Context, *sql.TxOptions) (dialect.Tx, error)
	}).BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("ent: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = &txDriver{tx: tx, drv: c.driver}
	return &Tx{
		ctx:    ctx,
		config: cfg,
		Pkg:    NewPkgClient(cfg),
	}, nil
}

// Debug returns a new debug-client. It's used to get verbose logging on specific operations.
//
//	client.Debug().
//		Pkg.
//		Query().
//		Count(ctx)
func (c *Client) Debug() *Client {
	if c.debug {
		return c
	}
	cfg := c.config
	cfg.driver = dialect.Debug(c.driver, c.log)
	client := &Client{config: cfg}
	client.init()
	return client
}

// Close closes the database connection and prevents new queries from starting.
func (c *Client) Close() error {
	return c.driver.Close()
}

// Use adds the mutation hooks to all the entity clients.
// In order to add hooks to a specific client, call: `client.Node.Use(...)`.
func (c *Client) Use(hooks ...Hook) {
	c.Pkg.Use(hooks...)
}

// PkgClient is a client for the Pkg schema.
type PkgClient struct {
	config
}

// NewPkgClient returns a client for the Pkg from the given config.
func NewPkgClient(c config) *PkgClient {
	return &PkgClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `pkg.Hooks(f(g(h())))`.
func (c *PkgClient) Use(hooks ...Hook) {
	c.hooks.Pkg = append(c.hooks.Pkg, hooks...)
}

// Create returns a builder for creating a Pkg entity.
func (c *PkgClient) Create() *PkgCreate {
	mutation := newPkgMutation(c.config, OpCreate)
	return &PkgCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Pkg entities.
func (c *PkgClient) CreateBulk(builders ...*PkgCreate) *PkgCreateBulk {
	return &PkgCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Pkg.
func (c *PkgClient) Update() *PkgUpdate {
	mutation := newPkgMutation(c.config, OpUpdate)
	return &PkgUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *PkgClient) UpdateOne(pk *Pkg) *PkgUpdateOne {
	mutation := newPkgMutation(c.config, OpUpdateOne, withPkg(pk))
	return &PkgUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *PkgClient) UpdateOneID(id int) *PkgUpdateOne {
	mutation := newPkgMutation(c.config, OpUpdateOne, withPkgID(id))
	return &PkgUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Pkg.
func (c *PkgClient) Delete() *PkgDelete {
	mutation := newPkgMutation(c.config, OpDelete)
	return &PkgDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *PkgClient) DeleteOne(pk *Pkg) *PkgDeleteOne {
	return c.DeleteOneID(pk.ID)
}

// DeleteOne returns a builder for deleting the given entity by its id.
func (c *PkgClient) DeleteOneID(id int) *PkgDeleteOne {
	builder := c.Delete().Where(pkg.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &PkgDeleteOne{builder}
}

// Query returns a query builder for Pkg.
func (c *PkgClient) Query() *PkgQuery {
	return &PkgQuery{
		config: c.config,
	}
}

// Get returns a Pkg entity by its id.
func (c *PkgClient) Get(ctx context.Context, id int) (*Pkg, error) {
	return c.Query().Where(pkg.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *PkgClient) GetX(ctx context.Context, id int) *Pkg {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// Hooks returns the client hooks.
func (c *PkgClient) Hooks() []Hook {
	return c.hooks.Pkg
}
