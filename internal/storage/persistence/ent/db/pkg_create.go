// Code generated by ent, DO NOT EDIT.

package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/location"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/pkg"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
	"github.com/google/uuid"
)

// PkgCreate is the builder for creating a Pkg entity.
type PkgCreate struct {
	config
	mutation *PkgMutation
	hooks    []Hook
}

// SetName sets the "name" field.
func (pc *PkgCreate) SetName(s string) *PkgCreate {
	pc.mutation.SetName(s)
	return pc
}

// SetAipID sets the "aip_id" field.
func (pc *PkgCreate) SetAipID(u uuid.UUID) *PkgCreate {
	pc.mutation.SetAipID(u)
	return pc
}

// SetLocationID sets the "location_id" field.
func (pc *PkgCreate) SetLocationID(i int) *PkgCreate {
	pc.mutation.SetLocationID(i)
	return pc
}

// SetNillableLocationID sets the "location_id" field if the given value is not nil.
func (pc *PkgCreate) SetNillableLocationID(i *int) *PkgCreate {
	if i != nil {
		pc.SetLocationID(*i)
	}
	return pc
}

// SetStatus sets the "status" field.
func (pc *PkgCreate) SetStatus(ts types.PackageStatus) *PkgCreate {
	pc.mutation.SetStatus(ts)
	return pc
}

// SetObjectKey sets the "object_key" field.
func (pc *PkgCreate) SetObjectKey(u uuid.UUID) *PkgCreate {
	pc.mutation.SetObjectKey(u)
	return pc
}

// SetCreatedAt sets the "created_at" field.
func (pc *PkgCreate) SetCreatedAt(t time.Time) *PkgCreate {
	pc.mutation.SetCreatedAt(t)
	return pc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (pc *PkgCreate) SetNillableCreatedAt(t *time.Time) *PkgCreate {
	if t != nil {
		pc.SetCreatedAt(*t)
	}
	return pc
}

// SetLocation sets the "location" edge to the Location entity.
func (pc *PkgCreate) SetLocation(l *Location) *PkgCreate {
	return pc.SetLocationID(l.ID)
}

// Mutation returns the PkgMutation object of the builder.
func (pc *PkgCreate) Mutation() *PkgMutation {
	return pc.mutation
}

// Save creates the Pkg in the database.
func (pc *PkgCreate) Save(ctx context.Context) (*Pkg, error) {
	pc.defaults()
	return withHooks(ctx, pc.sqlSave, pc.mutation, pc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (pc *PkgCreate) SaveX(ctx context.Context) *Pkg {
	v, err := pc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (pc *PkgCreate) Exec(ctx context.Context) error {
	_, err := pc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pc *PkgCreate) ExecX(ctx context.Context) {
	if err := pc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (pc *PkgCreate) defaults() {
	if _, ok := pc.mutation.CreatedAt(); !ok {
		v := pkg.DefaultCreatedAt()
		pc.mutation.SetCreatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (pc *PkgCreate) check() error {
	if _, ok := pc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`db: missing required field "Pkg.name"`)}
	}
	if _, ok := pc.mutation.AipID(); !ok {
		return &ValidationError{Name: "aip_id", err: errors.New(`db: missing required field "Pkg.aip_id"`)}
	}
	if _, ok := pc.mutation.Status(); !ok {
		return &ValidationError{Name: "status", err: errors.New(`db: missing required field "Pkg.status"`)}
	}
	if v, ok := pc.mutation.Status(); ok {
		if err := pkg.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`db: validator failed for field "Pkg.status": %w`, err)}
		}
	}
	if _, ok := pc.mutation.ObjectKey(); !ok {
		return &ValidationError{Name: "object_key", err: errors.New(`db: missing required field "Pkg.object_key"`)}
	}
	if _, ok := pc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`db: missing required field "Pkg.created_at"`)}
	}
	return nil
}

func (pc *PkgCreate) sqlSave(ctx context.Context) (*Pkg, error) {
	if err := pc.check(); err != nil {
		return nil, err
	}
	_node, _spec := pc.createSpec()
	if err := sqlgraph.CreateNode(ctx, pc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	pc.mutation.id = &_node.ID
	pc.mutation.done = true
	return _node, nil
}

func (pc *PkgCreate) createSpec() (*Pkg, *sqlgraph.CreateSpec) {
	var (
		_node = &Pkg{config: pc.config}
		_spec = sqlgraph.NewCreateSpec(pkg.Table, sqlgraph.NewFieldSpec(pkg.FieldID, field.TypeInt))
	)
	if value, ok := pc.mutation.Name(); ok {
		_spec.SetField(pkg.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := pc.mutation.AipID(); ok {
		_spec.SetField(pkg.FieldAipID, field.TypeUUID, value)
		_node.AipID = value
	}
	if value, ok := pc.mutation.Status(); ok {
		_spec.SetField(pkg.FieldStatus, field.TypeEnum, value)
		_node.Status = value
	}
	if value, ok := pc.mutation.ObjectKey(); ok {
		_spec.SetField(pkg.FieldObjectKey, field.TypeUUID, value)
		_node.ObjectKey = value
	}
	if value, ok := pc.mutation.CreatedAt(); ok {
		_spec.SetField(pkg.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if nodes := pc.mutation.LocationIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   pkg.LocationTable,
			Columns: []string{pkg.LocationColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(location.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.LocationID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// PkgCreateBulk is the builder for creating many Pkg entities in bulk.
type PkgCreateBulk struct {
	config
	err      error
	builders []*PkgCreate
}

// Save creates the Pkg entities in the database.
func (pcb *PkgCreateBulk) Save(ctx context.Context) ([]*Pkg, error) {
	if pcb.err != nil {
		return nil, pcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(pcb.builders))
	nodes := make([]*Pkg, len(pcb.builders))
	mutators := make([]Mutator, len(pcb.builders))
	for i := range pcb.builders {
		func(i int, root context.Context) {
			builder := pcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*PkgMutation)
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
					_, err = mutators[i+1].Mutate(root, pcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, pcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, pcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (pcb *PkgCreateBulk) SaveX(ctx context.Context) []*Pkg {
	v, err := pcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (pcb *PkgCreateBulk) Exec(ctx context.Context) error {
	_, err := pcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pcb *PkgCreateBulk) ExecX(ctx context.Context) {
	if err := pcb.Exec(ctx); err != nil {
		panic(err)
	}
}
