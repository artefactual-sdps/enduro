// Code generated by ent, DO NOT EDIT.

package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/aip"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/location"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
	"github.com/google/uuid"
)

// LocationCreate is the builder for creating a Location entity.
type LocationCreate struct {
	config
	mutation *LocationMutation
	hooks    []Hook
}

// SetName sets the "name" field.
func (lc *LocationCreate) SetName(s string) *LocationCreate {
	lc.mutation.SetName(s)
	return lc
}

// SetDescription sets the "description" field.
func (lc *LocationCreate) SetDescription(s string) *LocationCreate {
	lc.mutation.SetDescription(s)
	return lc
}

// SetSource sets the "source" field.
func (lc *LocationCreate) SetSource(es enums.LocationSource) *LocationCreate {
	lc.mutation.SetSource(es)
	return lc
}

// SetPurpose sets the "purpose" field.
func (lc *LocationCreate) SetPurpose(ep enums.LocationPurpose) *LocationCreate {
	lc.mutation.SetPurpose(ep)
	return lc
}

// SetUUID sets the "uuid" field.
func (lc *LocationCreate) SetUUID(u uuid.UUID) *LocationCreate {
	lc.mutation.SetUUID(u)
	return lc
}

// SetConfig sets the "config" field.
func (lc *LocationCreate) SetConfig(tc types.LocationConfig) *LocationCreate {
	lc.mutation.SetConfig(tc)
	return lc
}

// SetCreatedAt sets the "created_at" field.
func (lc *LocationCreate) SetCreatedAt(t time.Time) *LocationCreate {
	lc.mutation.SetCreatedAt(t)
	return lc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (lc *LocationCreate) SetNillableCreatedAt(t *time.Time) *LocationCreate {
	if t != nil {
		lc.SetCreatedAt(*t)
	}
	return lc
}

// AddAipIDs adds the "aips" edge to the AIP entity by IDs.
func (lc *LocationCreate) AddAipIDs(ids ...int) *LocationCreate {
	lc.mutation.AddAipIDs(ids...)
	return lc
}

// AddAips adds the "aips" edges to the AIP entity.
func (lc *LocationCreate) AddAips(a ...*AIP) *LocationCreate {
	ids := make([]int, len(a))
	for i := range a {
		ids[i] = a[i].ID
	}
	return lc.AddAipIDs(ids...)
}

// Mutation returns the LocationMutation object of the builder.
func (lc *LocationCreate) Mutation() *LocationMutation {
	return lc.mutation
}

// Save creates the Location in the database.
func (lc *LocationCreate) Save(ctx context.Context) (*Location, error) {
	lc.defaults()
	return withHooks(ctx, lc.sqlSave, lc.mutation, lc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (lc *LocationCreate) SaveX(ctx context.Context) *Location {
	v, err := lc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (lc *LocationCreate) Exec(ctx context.Context) error {
	_, err := lc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (lc *LocationCreate) ExecX(ctx context.Context) {
	if err := lc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (lc *LocationCreate) defaults() {
	if _, ok := lc.mutation.CreatedAt(); !ok {
		v := location.DefaultCreatedAt()
		lc.mutation.SetCreatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (lc *LocationCreate) check() error {
	if _, ok := lc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`db: missing required field "Location.name"`)}
	}
	if _, ok := lc.mutation.Description(); !ok {
		return &ValidationError{Name: "description", err: errors.New(`db: missing required field "Location.description"`)}
	}
	if _, ok := lc.mutation.Source(); !ok {
		return &ValidationError{Name: "source", err: errors.New(`db: missing required field "Location.source"`)}
	}
	if v, ok := lc.mutation.Source(); ok {
		if err := location.SourceValidator(v); err != nil {
			return &ValidationError{Name: "source", err: fmt.Errorf(`db: validator failed for field "Location.source": %w`, err)}
		}
	}
	if _, ok := lc.mutation.Purpose(); !ok {
		return &ValidationError{Name: "purpose", err: errors.New(`db: missing required field "Location.purpose"`)}
	}
	if v, ok := lc.mutation.Purpose(); ok {
		if err := location.PurposeValidator(v); err != nil {
			return &ValidationError{Name: "purpose", err: fmt.Errorf(`db: validator failed for field "Location.purpose": %w`, err)}
		}
	}
	if _, ok := lc.mutation.UUID(); !ok {
		return &ValidationError{Name: "uuid", err: errors.New(`db: missing required field "Location.uuid"`)}
	}
	if _, ok := lc.mutation.Config(); !ok {
		return &ValidationError{Name: "config", err: errors.New(`db: missing required field "Location.config"`)}
	}
	if _, ok := lc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`db: missing required field "Location.created_at"`)}
	}
	return nil
}

func (lc *LocationCreate) sqlSave(ctx context.Context) (*Location, error) {
	if err := lc.check(); err != nil {
		return nil, err
	}
	_node, _spec := lc.createSpec()
	if err := sqlgraph.CreateNode(ctx, lc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	lc.mutation.id = &_node.ID
	lc.mutation.done = true
	return _node, nil
}

func (lc *LocationCreate) createSpec() (*Location, *sqlgraph.CreateSpec) {
	var (
		_node = &Location{config: lc.config}
		_spec = sqlgraph.NewCreateSpec(location.Table, sqlgraph.NewFieldSpec(location.FieldID, field.TypeInt))
	)
	if value, ok := lc.mutation.Name(); ok {
		_spec.SetField(location.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := lc.mutation.Description(); ok {
		_spec.SetField(location.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	if value, ok := lc.mutation.Source(); ok {
		_spec.SetField(location.FieldSource, field.TypeEnum, value)
		_node.Source = value
	}
	if value, ok := lc.mutation.Purpose(); ok {
		_spec.SetField(location.FieldPurpose, field.TypeEnum, value)
		_node.Purpose = value
	}
	if value, ok := lc.mutation.UUID(); ok {
		_spec.SetField(location.FieldUUID, field.TypeUUID, value)
		_node.UUID = value
	}
	if value, ok := lc.mutation.Config(); ok {
		_spec.SetField(location.FieldConfig, field.TypeJSON, value)
		_node.Config = value
	}
	if value, ok := lc.mutation.CreatedAt(); ok {
		_spec.SetField(location.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if nodes := lc.mutation.AipsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   location.AipsTable,
			Columns: []string{location.AipsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(aip.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// LocationCreateBulk is the builder for creating many Location entities in bulk.
type LocationCreateBulk struct {
	config
	err      error
	builders []*LocationCreate
}

// Save creates the Location entities in the database.
func (lcb *LocationCreateBulk) Save(ctx context.Context) ([]*Location, error) {
	if lcb.err != nil {
		return nil, lcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(lcb.builders))
	nodes := make([]*Location, len(lcb.builders))
	mutators := make([]Mutator, len(lcb.builders))
	for i := range lcb.builders {
		func(i int, root context.Context) {
			builder := lcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*LocationMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, lcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, lcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				if specs[i].ID.Value != nil {
					id := specs[i].ID.Value.(int64)
					nodes[i].ID = int(id)
				}
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, lcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (lcb *LocationCreateBulk) SaveX(ctx context.Context) []*Location {
	v, err := lcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (lcb *LocationCreateBulk) Exec(ctx context.Context) error {
	_, err := lcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (lcb *LocationCreateBulk) ExecX(ctx context.Context) {
	if err := lcb.Exec(ctx); err != nil {
		panic(err)
	}
}
