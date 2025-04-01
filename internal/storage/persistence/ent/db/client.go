// Code generated by ent, DO NOT EDIT.

package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/migrate"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/aip"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/deletionrequest"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/location"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/task"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/workflow"
)

// Client is the client that holds all ent builders.
type Client struct {
	config
	// Schema is the client for creating, migrating and dropping schema.
	Schema *migrate.Schema
	// AIP is the client for interacting with the AIP builders.
	AIP *AIPClient
	// DeletionRequest is the client for interacting with the DeletionRequest builders.
	DeletionRequest *DeletionRequestClient
	// Location is the client for interacting with the Location builders.
	Location *LocationClient
	// Task is the client for interacting with the Task builders.
	Task *TaskClient
	// Workflow is the client for interacting with the Workflow builders.
	Workflow *WorkflowClient
}

// NewClient creates a new client configured with the given options.
func NewClient(opts ...Option) *Client {
	client := &Client{config: newConfig(opts...)}
	client.init()
	return client
}

func (c *Client) init() {
	c.Schema = migrate.NewSchema(c.driver)
	c.AIP = NewAIPClient(c.config)
	c.DeletionRequest = NewDeletionRequestClient(c.config)
	c.Location = NewLocationClient(c.config)
	c.Task = NewTaskClient(c.config)
	c.Workflow = NewWorkflowClient(c.config)
}

type (
	// config is the configuration for the client and its builder.
	config struct {
		// driver used for executing database requests.
		driver dialect.Driver
		// debug enable a debug logging.
		debug bool
		// log used for logging on debug mode.
		log func(...any)
		// hooks to execute on mutations.
		hooks *hooks
		// interceptors to execute on queries.
		inters *inters
	}
	// Option function to configure the client.
	Option func(*config)
)

// newConfig creates a new config for the client.
func newConfig(opts ...Option) config {
	cfg := config{log: log.Println, hooks: &hooks{}, inters: &inters{}}
	cfg.options(opts...)
	return cfg
}

// options applies the options on the config object.
func (c *config) options(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
	if c.debug {
		c.driver = dialect.Debug(c.driver, c.log)
	}
}

// Debug enables debug logging on the ent.Driver.
func Debug() Option {
	return func(c *config) {
		c.debug = true
	}
}

// Log sets the logging function for debug mode.
func Log(fn func(...any)) Option {
	return func(c *config) {
		c.log = fn
	}
}

// Driver configures the client driver.
func Driver(driver dialect.Driver) Option {
	return func(c *config) {
		c.driver = driver
	}
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

// ErrTxStarted is returned when trying to start a new transaction from a transactional client.
var ErrTxStarted = errors.New("db: cannot start a transaction within a transaction")

// Tx returns a new transactional client. The provided context
// is used until the transaction is committed or rolled back.
func (c *Client) Tx(ctx context.Context) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, ErrTxStarted
	}
	tx, err := newTx(ctx, c.driver)
	if err != nil {
		return nil, fmt.Errorf("db: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = tx
	return &Tx{
		ctx:             ctx,
		config:          cfg,
		AIP:             NewAIPClient(cfg),
		DeletionRequest: NewDeletionRequestClient(cfg),
		Location:        NewLocationClient(cfg),
		Task:            NewTaskClient(cfg),
		Workflow:        NewWorkflowClient(cfg),
	}, nil
}

// BeginTx returns a transactional client with specified options.
func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, errors.New("ent: cannot start a transaction within a transaction")
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
		ctx:             ctx,
		config:          cfg,
		AIP:             NewAIPClient(cfg),
		DeletionRequest: NewDeletionRequestClient(cfg),
		Location:        NewLocationClient(cfg),
		Task:            NewTaskClient(cfg),
		Workflow:        NewWorkflowClient(cfg),
	}, nil
}

// Debug returns a new debug-client. It's used to get verbose logging on specific operations.
//
//	client.Debug().
//		AIP.
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
	c.AIP.Use(hooks...)
	c.DeletionRequest.Use(hooks...)
	c.Location.Use(hooks...)
	c.Task.Use(hooks...)
	c.Workflow.Use(hooks...)
}

// Intercept adds the query interceptors to all the entity clients.
// In order to add interceptors to a specific client, call: `client.Node.Intercept(...)`.
func (c *Client) Intercept(interceptors ...Interceptor) {
	c.AIP.Intercept(interceptors...)
	c.DeletionRequest.Intercept(interceptors...)
	c.Location.Intercept(interceptors...)
	c.Task.Intercept(interceptors...)
	c.Workflow.Intercept(interceptors...)
}

// Mutate implements the ent.Mutator interface.
func (c *Client) Mutate(ctx context.Context, m Mutation) (Value, error) {
	switch m := m.(type) {
	case *AIPMutation:
		return c.AIP.mutate(ctx, m)
	case *DeletionRequestMutation:
		return c.DeletionRequest.mutate(ctx, m)
	case *LocationMutation:
		return c.Location.mutate(ctx, m)
	case *TaskMutation:
		return c.Task.mutate(ctx, m)
	case *WorkflowMutation:
		return c.Workflow.mutate(ctx, m)
	default:
		return nil, fmt.Errorf("db: unknown mutation type %T", m)
	}
}

// AIPClient is a client for the AIP schema.
type AIPClient struct {
	config
}

// NewAIPClient returns a client for the AIP from the given config.
func NewAIPClient(c config) *AIPClient {
	return &AIPClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `aip.Hooks(f(g(h())))`.
func (c *AIPClient) Use(hooks ...Hook) {
	c.hooks.AIP = append(c.hooks.AIP, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `aip.Intercept(f(g(h())))`.
func (c *AIPClient) Intercept(interceptors ...Interceptor) {
	c.inters.AIP = append(c.inters.AIP, interceptors...)
}

// Create returns a builder for creating a AIP entity.
func (c *AIPClient) Create() *AIPCreate {
	mutation := newAIPMutation(c.config, OpCreate)
	return &AIPCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of AIP entities.
func (c *AIPClient) CreateBulk(builders ...*AIPCreate) *AIPCreateBulk {
	return &AIPCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *AIPClient) MapCreateBulk(slice any, setFunc func(*AIPCreate, int)) *AIPCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &AIPCreateBulk{err: fmt.Errorf("calling to AIPClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*AIPCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &AIPCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for AIP.
func (c *AIPClient) Update() *AIPUpdate {
	mutation := newAIPMutation(c.config, OpUpdate)
	return &AIPUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *AIPClient) UpdateOne(a *AIP) *AIPUpdateOne {
	mutation := newAIPMutation(c.config, OpUpdateOne, withAIP(a))
	return &AIPUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *AIPClient) UpdateOneID(id int) *AIPUpdateOne {
	mutation := newAIPMutation(c.config, OpUpdateOne, withAIPID(id))
	return &AIPUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for AIP.
func (c *AIPClient) Delete() *AIPDelete {
	mutation := newAIPMutation(c.config, OpDelete)
	return &AIPDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *AIPClient) DeleteOne(a *AIP) *AIPDeleteOne {
	return c.DeleteOneID(a.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *AIPClient) DeleteOneID(id int) *AIPDeleteOne {
	builder := c.Delete().Where(aip.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &AIPDeleteOne{builder}
}

// Query returns a query builder for AIP.
func (c *AIPClient) Query() *AIPQuery {
	return &AIPQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeAIP},
		inters: c.Interceptors(),
	}
}

// Get returns a AIP entity by its id.
func (c *AIPClient) Get(ctx context.Context, id int) (*AIP, error) {
	return c.Query().Where(aip.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *AIPClient) GetX(ctx context.Context, id int) *AIP {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryLocation queries the location edge of a AIP.
func (c *AIPClient) QueryLocation(a *AIP) *LocationQuery {
	query := (&LocationClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := a.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(aip.Table, aip.FieldID, id),
			sqlgraph.To(location.Table, location.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, aip.LocationTable, aip.LocationColumn),
		)
		fromV = sqlgraph.Neighbors(a.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// QueryWorkflows queries the workflows edge of a AIP.
func (c *AIPClient) QueryWorkflows(a *AIP) *WorkflowQuery {
	query := (&WorkflowClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := a.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(aip.Table, aip.FieldID, id),
			sqlgraph.To(workflow.Table, workflow.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, aip.WorkflowsTable, aip.WorkflowsColumn),
		)
		fromV = sqlgraph.Neighbors(a.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// QueryDeletionRequests queries the deletion_requests edge of a AIP.
func (c *AIPClient) QueryDeletionRequests(a *AIP) *DeletionRequestQuery {
	query := (&DeletionRequestClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := a.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(aip.Table, aip.FieldID, id),
			sqlgraph.To(deletionrequest.Table, deletionrequest.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, aip.DeletionRequestsTable, aip.DeletionRequestsColumn),
		)
		fromV = sqlgraph.Neighbors(a.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *AIPClient) Hooks() []Hook {
	return c.hooks.AIP
}

// Interceptors returns the client interceptors.
func (c *AIPClient) Interceptors() []Interceptor {
	return c.inters.AIP
}

func (c *AIPClient) mutate(ctx context.Context, m *AIPMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&AIPCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&AIPUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&AIPUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&AIPDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("db: unknown AIP mutation op: %q", m.Op())
	}
}

// DeletionRequestClient is a client for the DeletionRequest schema.
type DeletionRequestClient struct {
	config
}

// NewDeletionRequestClient returns a client for the DeletionRequest from the given config.
func NewDeletionRequestClient(c config) *DeletionRequestClient {
	return &DeletionRequestClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `deletionrequest.Hooks(f(g(h())))`.
func (c *DeletionRequestClient) Use(hooks ...Hook) {
	c.hooks.DeletionRequest = append(c.hooks.DeletionRequest, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `deletionrequest.Intercept(f(g(h())))`.
func (c *DeletionRequestClient) Intercept(interceptors ...Interceptor) {
	c.inters.DeletionRequest = append(c.inters.DeletionRequest, interceptors...)
}

// Create returns a builder for creating a DeletionRequest entity.
func (c *DeletionRequestClient) Create() *DeletionRequestCreate {
	mutation := newDeletionRequestMutation(c.config, OpCreate)
	return &DeletionRequestCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of DeletionRequest entities.
func (c *DeletionRequestClient) CreateBulk(builders ...*DeletionRequestCreate) *DeletionRequestCreateBulk {
	return &DeletionRequestCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *DeletionRequestClient) MapCreateBulk(slice any, setFunc func(*DeletionRequestCreate, int)) *DeletionRequestCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &DeletionRequestCreateBulk{err: fmt.Errorf("calling to DeletionRequestClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*DeletionRequestCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &DeletionRequestCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for DeletionRequest.
func (c *DeletionRequestClient) Update() *DeletionRequestUpdate {
	mutation := newDeletionRequestMutation(c.config, OpUpdate)
	return &DeletionRequestUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *DeletionRequestClient) UpdateOne(dr *DeletionRequest) *DeletionRequestUpdateOne {
	mutation := newDeletionRequestMutation(c.config, OpUpdateOne, withDeletionRequest(dr))
	return &DeletionRequestUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *DeletionRequestClient) UpdateOneID(id int) *DeletionRequestUpdateOne {
	mutation := newDeletionRequestMutation(c.config, OpUpdateOne, withDeletionRequestID(id))
	return &DeletionRequestUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for DeletionRequest.
func (c *DeletionRequestClient) Delete() *DeletionRequestDelete {
	mutation := newDeletionRequestMutation(c.config, OpDelete)
	return &DeletionRequestDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *DeletionRequestClient) DeleteOne(dr *DeletionRequest) *DeletionRequestDeleteOne {
	return c.DeleteOneID(dr.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *DeletionRequestClient) DeleteOneID(id int) *DeletionRequestDeleteOne {
	builder := c.Delete().Where(deletionrequest.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &DeletionRequestDeleteOne{builder}
}

// Query returns a query builder for DeletionRequest.
func (c *DeletionRequestClient) Query() *DeletionRequestQuery {
	return &DeletionRequestQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeDeletionRequest},
		inters: c.Interceptors(),
	}
}

// Get returns a DeletionRequest entity by its id.
func (c *DeletionRequestClient) Get(ctx context.Context, id int) (*DeletionRequest, error) {
	return c.Query().Where(deletionrequest.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *DeletionRequestClient) GetX(ctx context.Context, id int) *DeletionRequest {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryAip queries the aip edge of a DeletionRequest.
func (c *DeletionRequestClient) QueryAip(dr *DeletionRequest) *AIPQuery {
	query := (&AIPClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := dr.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(deletionrequest.Table, deletionrequest.FieldID, id),
			sqlgraph.To(aip.Table, aip.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, deletionrequest.AipTable, deletionrequest.AipColumn),
		)
		fromV = sqlgraph.Neighbors(dr.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// QueryWorkflow queries the workflow edge of a DeletionRequest.
func (c *DeletionRequestClient) QueryWorkflow(dr *DeletionRequest) *WorkflowQuery {
	query := (&WorkflowClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := dr.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(deletionrequest.Table, deletionrequest.FieldID, id),
			sqlgraph.To(workflow.Table, workflow.FieldID),
			sqlgraph.Edge(sqlgraph.O2O, true, deletionrequest.WorkflowTable, deletionrequest.WorkflowColumn),
		)
		fromV = sqlgraph.Neighbors(dr.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *DeletionRequestClient) Hooks() []Hook {
	return c.hooks.DeletionRequest
}

// Interceptors returns the client interceptors.
func (c *DeletionRequestClient) Interceptors() []Interceptor {
	return c.inters.DeletionRequest
}

func (c *DeletionRequestClient) mutate(ctx context.Context, m *DeletionRequestMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&DeletionRequestCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&DeletionRequestUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&DeletionRequestUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&DeletionRequestDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("db: unknown DeletionRequest mutation op: %q", m.Op())
	}
}

// LocationClient is a client for the Location schema.
type LocationClient struct {
	config
}

// NewLocationClient returns a client for the Location from the given config.
func NewLocationClient(c config) *LocationClient {
	return &LocationClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `location.Hooks(f(g(h())))`.
func (c *LocationClient) Use(hooks ...Hook) {
	c.hooks.Location = append(c.hooks.Location, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `location.Intercept(f(g(h())))`.
func (c *LocationClient) Intercept(interceptors ...Interceptor) {
	c.inters.Location = append(c.inters.Location, interceptors...)
}

// Create returns a builder for creating a Location entity.
func (c *LocationClient) Create() *LocationCreate {
	mutation := newLocationMutation(c.config, OpCreate)
	return &LocationCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Location entities.
func (c *LocationClient) CreateBulk(builders ...*LocationCreate) *LocationCreateBulk {
	return &LocationCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *LocationClient) MapCreateBulk(slice any, setFunc func(*LocationCreate, int)) *LocationCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &LocationCreateBulk{err: fmt.Errorf("calling to LocationClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*LocationCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &LocationCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Location.
func (c *LocationClient) Update() *LocationUpdate {
	mutation := newLocationMutation(c.config, OpUpdate)
	return &LocationUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *LocationClient) UpdateOne(l *Location) *LocationUpdateOne {
	mutation := newLocationMutation(c.config, OpUpdateOne, withLocation(l))
	return &LocationUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *LocationClient) UpdateOneID(id int) *LocationUpdateOne {
	mutation := newLocationMutation(c.config, OpUpdateOne, withLocationID(id))
	return &LocationUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Location.
func (c *LocationClient) Delete() *LocationDelete {
	mutation := newLocationMutation(c.config, OpDelete)
	return &LocationDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *LocationClient) DeleteOne(l *Location) *LocationDeleteOne {
	return c.DeleteOneID(l.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *LocationClient) DeleteOneID(id int) *LocationDeleteOne {
	builder := c.Delete().Where(location.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &LocationDeleteOne{builder}
}

// Query returns a query builder for Location.
func (c *LocationClient) Query() *LocationQuery {
	return &LocationQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeLocation},
		inters: c.Interceptors(),
	}
}

// Get returns a Location entity by its id.
func (c *LocationClient) Get(ctx context.Context, id int) (*Location, error) {
	return c.Query().Where(location.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *LocationClient) GetX(ctx context.Context, id int) *Location {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryAips queries the aips edge of a Location.
func (c *LocationClient) QueryAips(l *Location) *AIPQuery {
	query := (&AIPClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := l.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(location.Table, location.FieldID, id),
			sqlgraph.To(aip.Table, aip.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, location.AipsTable, location.AipsColumn),
		)
		fromV = sqlgraph.Neighbors(l.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *LocationClient) Hooks() []Hook {
	return c.hooks.Location
}

// Interceptors returns the client interceptors.
func (c *LocationClient) Interceptors() []Interceptor {
	return c.inters.Location
}

func (c *LocationClient) mutate(ctx context.Context, m *LocationMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&LocationCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&LocationUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&LocationUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&LocationDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("db: unknown Location mutation op: %q", m.Op())
	}
}

// TaskClient is a client for the Task schema.
type TaskClient struct {
	config
}

// NewTaskClient returns a client for the Task from the given config.
func NewTaskClient(c config) *TaskClient {
	return &TaskClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `task.Hooks(f(g(h())))`.
func (c *TaskClient) Use(hooks ...Hook) {
	c.hooks.Task = append(c.hooks.Task, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `task.Intercept(f(g(h())))`.
func (c *TaskClient) Intercept(interceptors ...Interceptor) {
	c.inters.Task = append(c.inters.Task, interceptors...)
}

// Create returns a builder for creating a Task entity.
func (c *TaskClient) Create() *TaskCreate {
	mutation := newTaskMutation(c.config, OpCreate)
	return &TaskCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Task entities.
func (c *TaskClient) CreateBulk(builders ...*TaskCreate) *TaskCreateBulk {
	return &TaskCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *TaskClient) MapCreateBulk(slice any, setFunc func(*TaskCreate, int)) *TaskCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &TaskCreateBulk{err: fmt.Errorf("calling to TaskClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*TaskCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &TaskCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Task.
func (c *TaskClient) Update() *TaskUpdate {
	mutation := newTaskMutation(c.config, OpUpdate)
	return &TaskUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *TaskClient) UpdateOne(t *Task) *TaskUpdateOne {
	mutation := newTaskMutation(c.config, OpUpdateOne, withTask(t))
	return &TaskUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *TaskClient) UpdateOneID(id int) *TaskUpdateOne {
	mutation := newTaskMutation(c.config, OpUpdateOne, withTaskID(id))
	return &TaskUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Task.
func (c *TaskClient) Delete() *TaskDelete {
	mutation := newTaskMutation(c.config, OpDelete)
	return &TaskDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *TaskClient) DeleteOne(t *Task) *TaskDeleteOne {
	return c.DeleteOneID(t.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *TaskClient) DeleteOneID(id int) *TaskDeleteOne {
	builder := c.Delete().Where(task.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &TaskDeleteOne{builder}
}

// Query returns a query builder for Task.
func (c *TaskClient) Query() *TaskQuery {
	return &TaskQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeTask},
		inters: c.Interceptors(),
	}
}

// Get returns a Task entity by its id.
func (c *TaskClient) Get(ctx context.Context, id int) (*Task, error) {
	return c.Query().Where(task.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *TaskClient) GetX(ctx context.Context, id int) *Task {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryWorkflow queries the workflow edge of a Task.
func (c *TaskClient) QueryWorkflow(t *Task) *WorkflowQuery {
	query := (&WorkflowClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := t.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(task.Table, task.FieldID, id),
			sqlgraph.To(workflow.Table, workflow.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, task.WorkflowTable, task.WorkflowColumn),
		)
		fromV = sqlgraph.Neighbors(t.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *TaskClient) Hooks() []Hook {
	return c.hooks.Task
}

// Interceptors returns the client interceptors.
func (c *TaskClient) Interceptors() []Interceptor {
	return c.inters.Task
}

func (c *TaskClient) mutate(ctx context.Context, m *TaskMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&TaskCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&TaskUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&TaskUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&TaskDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("db: unknown Task mutation op: %q", m.Op())
	}
}

// WorkflowClient is a client for the Workflow schema.
type WorkflowClient struct {
	config
}

// NewWorkflowClient returns a client for the Workflow from the given config.
func NewWorkflowClient(c config) *WorkflowClient {
	return &WorkflowClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `workflow.Hooks(f(g(h())))`.
func (c *WorkflowClient) Use(hooks ...Hook) {
	c.hooks.Workflow = append(c.hooks.Workflow, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `workflow.Intercept(f(g(h())))`.
func (c *WorkflowClient) Intercept(interceptors ...Interceptor) {
	c.inters.Workflow = append(c.inters.Workflow, interceptors...)
}

// Create returns a builder for creating a Workflow entity.
func (c *WorkflowClient) Create() *WorkflowCreate {
	mutation := newWorkflowMutation(c.config, OpCreate)
	return &WorkflowCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Workflow entities.
func (c *WorkflowClient) CreateBulk(builders ...*WorkflowCreate) *WorkflowCreateBulk {
	return &WorkflowCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *WorkflowClient) MapCreateBulk(slice any, setFunc func(*WorkflowCreate, int)) *WorkflowCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &WorkflowCreateBulk{err: fmt.Errorf("calling to WorkflowClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*WorkflowCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &WorkflowCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Workflow.
func (c *WorkflowClient) Update() *WorkflowUpdate {
	mutation := newWorkflowMutation(c.config, OpUpdate)
	return &WorkflowUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *WorkflowClient) UpdateOne(w *Workflow) *WorkflowUpdateOne {
	mutation := newWorkflowMutation(c.config, OpUpdateOne, withWorkflow(w))
	return &WorkflowUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *WorkflowClient) UpdateOneID(id int) *WorkflowUpdateOne {
	mutation := newWorkflowMutation(c.config, OpUpdateOne, withWorkflowID(id))
	return &WorkflowUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Workflow.
func (c *WorkflowClient) Delete() *WorkflowDelete {
	mutation := newWorkflowMutation(c.config, OpDelete)
	return &WorkflowDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *WorkflowClient) DeleteOne(w *Workflow) *WorkflowDeleteOne {
	return c.DeleteOneID(w.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *WorkflowClient) DeleteOneID(id int) *WorkflowDeleteOne {
	builder := c.Delete().Where(workflow.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &WorkflowDeleteOne{builder}
}

// Query returns a query builder for Workflow.
func (c *WorkflowClient) Query() *WorkflowQuery {
	return &WorkflowQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeWorkflow},
		inters: c.Interceptors(),
	}
}

// Get returns a Workflow entity by its id.
func (c *WorkflowClient) Get(ctx context.Context, id int) (*Workflow, error) {
	return c.Query().Where(workflow.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *WorkflowClient) GetX(ctx context.Context, id int) *Workflow {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryAip queries the aip edge of a Workflow.
func (c *WorkflowClient) QueryAip(w *Workflow) *AIPQuery {
	query := (&AIPClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := w.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(workflow.Table, workflow.FieldID, id),
			sqlgraph.To(aip.Table, aip.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, workflow.AipTable, workflow.AipColumn),
		)
		fromV = sqlgraph.Neighbors(w.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// QueryTasks queries the tasks edge of a Workflow.
func (c *WorkflowClient) QueryTasks(w *Workflow) *TaskQuery {
	query := (&TaskClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := w.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(workflow.Table, workflow.FieldID, id),
			sqlgraph.To(task.Table, task.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, workflow.TasksTable, workflow.TasksColumn),
		)
		fromV = sqlgraph.Neighbors(w.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// QueryDeletionRequest queries the deletion_request edge of a Workflow.
func (c *WorkflowClient) QueryDeletionRequest(w *Workflow) *DeletionRequestQuery {
	query := (&DeletionRequestClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := w.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(workflow.Table, workflow.FieldID, id),
			sqlgraph.To(deletionrequest.Table, deletionrequest.FieldID),
			sqlgraph.Edge(sqlgraph.O2O, false, workflow.DeletionRequestTable, workflow.DeletionRequestColumn),
		)
		fromV = sqlgraph.Neighbors(w.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *WorkflowClient) Hooks() []Hook {
	return c.hooks.Workflow
}

// Interceptors returns the client interceptors.
func (c *WorkflowClient) Interceptors() []Interceptor {
	return c.inters.Workflow
}

func (c *WorkflowClient) mutate(ctx context.Context, m *WorkflowMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&WorkflowCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&WorkflowUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&WorkflowUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&WorkflowDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("db: unknown Workflow mutation op: %q", m.Op())
	}
}

// hooks and interceptors per client, for fast access.
type (
	hooks struct {
		AIP, DeletionRequest, Location, Task, Workflow []ent.Hook
	}
	inters struct {
		AIP, DeletionRequest, Location, Task, Workflow []ent.Interceptor
	}
)
