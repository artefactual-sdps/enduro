// Code generated by ent, DO NOT EDIT.

package db

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/location"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/pkg"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/predicate"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
	"github.com/google/uuid"
)

// PkgUpdate is the builder for updating Pkg entities.
type PkgUpdate struct {
	config
	hooks    []Hook
	mutation *PkgMutation
}

// Where appends a list predicates to the PkgUpdate builder.
func (pu *PkgUpdate) Where(ps ...predicate.Pkg) *PkgUpdate {
	pu.mutation.Where(ps...)
	return pu
}

// SetName sets the "name" field.
func (pu *PkgUpdate) SetName(s string) *PkgUpdate {
	pu.mutation.SetName(s)
	return pu
}

// SetNillableName sets the "name" field if the given value is not nil.
func (pu *PkgUpdate) SetNillableName(s *string) *PkgUpdate {
	if s != nil {
		pu.SetName(*s)
	}
	return pu
}

// SetAipID sets the "aip_id" field.
func (pu *PkgUpdate) SetAipID(u uuid.UUID) *PkgUpdate {
	pu.mutation.SetAipID(u)
	return pu
}

// SetNillableAipID sets the "aip_id" field if the given value is not nil.
func (pu *PkgUpdate) SetNillableAipID(u *uuid.UUID) *PkgUpdate {
	if u != nil {
		pu.SetAipID(*u)
	}
	return pu
}

// SetLocationID sets the "location_id" field.
func (pu *PkgUpdate) SetLocationID(i int) *PkgUpdate {
	pu.mutation.SetLocationID(i)
	return pu
}

// SetNillableLocationID sets the "location_id" field if the given value is not nil.
func (pu *PkgUpdate) SetNillableLocationID(i *int) *PkgUpdate {
	if i != nil {
		pu.SetLocationID(*i)
	}
	return pu
}

// ClearLocationID clears the value of the "location_id" field.
func (pu *PkgUpdate) ClearLocationID() *PkgUpdate {
	pu.mutation.ClearLocationID()
	return pu
}

// SetStatus sets the "status" field.
func (pu *PkgUpdate) SetStatus(ts types.PackageStatus) *PkgUpdate {
	pu.mutation.SetStatus(ts)
	return pu
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (pu *PkgUpdate) SetNillableStatus(ts *types.PackageStatus) *PkgUpdate {
	if ts != nil {
		pu.SetStatus(*ts)
	}
	return pu
}

// SetObjectKey sets the "object_key" field.
func (pu *PkgUpdate) SetObjectKey(u uuid.UUID) *PkgUpdate {
	pu.mutation.SetObjectKey(u)
	return pu
}

// SetNillableObjectKey sets the "object_key" field if the given value is not nil.
func (pu *PkgUpdate) SetNillableObjectKey(u *uuid.UUID) *PkgUpdate {
	if u != nil {
		pu.SetObjectKey(*u)
	}
	return pu
}

// SetLocation sets the "location" edge to the Location entity.
func (pu *PkgUpdate) SetLocation(l *Location) *PkgUpdate {
	return pu.SetLocationID(l.ID)
}

// Mutation returns the PkgMutation object of the builder.
func (pu *PkgUpdate) Mutation() *PkgMutation {
	return pu.mutation
}

// ClearLocation clears the "location" edge to the Location entity.
func (pu *PkgUpdate) ClearLocation() *PkgUpdate {
	pu.mutation.ClearLocation()
	return pu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (pu *PkgUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, pu.sqlSave, pu.mutation, pu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (pu *PkgUpdate) SaveX(ctx context.Context) int {
	affected, err := pu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (pu *PkgUpdate) Exec(ctx context.Context) error {
	_, err := pu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pu *PkgUpdate) ExecX(ctx context.Context) {
	if err := pu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (pu *PkgUpdate) check() error {
	if v, ok := pu.mutation.Status(); ok {
		if err := pkg.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`db: validator failed for field "Pkg.status": %w`, err)}
		}
	}
	return nil
}

func (pu *PkgUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := pu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(pkg.Table, pkg.Columns, sqlgraph.NewFieldSpec(pkg.FieldID, field.TypeInt))
	if ps := pu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := pu.mutation.Name(); ok {
		_spec.SetField(pkg.FieldName, field.TypeString, value)
	}
	if value, ok := pu.mutation.AipID(); ok {
		_spec.SetField(pkg.FieldAipID, field.TypeUUID, value)
	}
	if value, ok := pu.mutation.Status(); ok {
		_spec.SetField(pkg.FieldStatus, field.TypeEnum, value)
	}
	if value, ok := pu.mutation.ObjectKey(); ok {
		_spec.SetField(pkg.FieldObjectKey, field.TypeUUID, value)
	}
	if pu.mutation.LocationCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := pu.mutation.LocationIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, pu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{pkg.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	pu.mutation.done = true
	return n, nil
}

// PkgUpdateOne is the builder for updating a single Pkg entity.
type PkgUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *PkgMutation
}

// SetName sets the "name" field.
func (puo *PkgUpdateOne) SetName(s string) *PkgUpdateOne {
	puo.mutation.SetName(s)
	return puo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (puo *PkgUpdateOne) SetNillableName(s *string) *PkgUpdateOne {
	if s != nil {
		puo.SetName(*s)
	}
	return puo
}

// SetAipID sets the "aip_id" field.
func (puo *PkgUpdateOne) SetAipID(u uuid.UUID) *PkgUpdateOne {
	puo.mutation.SetAipID(u)
	return puo
}

// SetNillableAipID sets the "aip_id" field if the given value is not nil.
func (puo *PkgUpdateOne) SetNillableAipID(u *uuid.UUID) *PkgUpdateOne {
	if u != nil {
		puo.SetAipID(*u)
	}
	return puo
}

// SetLocationID sets the "location_id" field.
func (puo *PkgUpdateOne) SetLocationID(i int) *PkgUpdateOne {
	puo.mutation.SetLocationID(i)
	return puo
}

// SetNillableLocationID sets the "location_id" field if the given value is not nil.
func (puo *PkgUpdateOne) SetNillableLocationID(i *int) *PkgUpdateOne {
	if i != nil {
		puo.SetLocationID(*i)
	}
	return puo
}

// ClearLocationID clears the value of the "location_id" field.
func (puo *PkgUpdateOne) ClearLocationID() *PkgUpdateOne {
	puo.mutation.ClearLocationID()
	return puo
}

// SetStatus sets the "status" field.
func (puo *PkgUpdateOne) SetStatus(ts types.PackageStatus) *PkgUpdateOne {
	puo.mutation.SetStatus(ts)
	return puo
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (puo *PkgUpdateOne) SetNillableStatus(ts *types.PackageStatus) *PkgUpdateOne {
	if ts != nil {
		puo.SetStatus(*ts)
	}
	return puo
}

// SetObjectKey sets the "object_key" field.
func (puo *PkgUpdateOne) SetObjectKey(u uuid.UUID) *PkgUpdateOne {
	puo.mutation.SetObjectKey(u)
	return puo
}

// SetNillableObjectKey sets the "object_key" field if the given value is not nil.
func (puo *PkgUpdateOne) SetNillableObjectKey(u *uuid.UUID) *PkgUpdateOne {
	if u != nil {
		puo.SetObjectKey(*u)
	}
	return puo
}

// SetLocation sets the "location" edge to the Location entity.
func (puo *PkgUpdateOne) SetLocation(l *Location) *PkgUpdateOne {
	return puo.SetLocationID(l.ID)
}

// Mutation returns the PkgMutation object of the builder.
func (puo *PkgUpdateOne) Mutation() *PkgMutation {
	return puo.mutation
}

// ClearLocation clears the "location" edge to the Location entity.
func (puo *PkgUpdateOne) ClearLocation() *PkgUpdateOne {
	puo.mutation.ClearLocation()
	return puo
}

// Where appends a list predicates to the PkgUpdate builder.
func (puo *PkgUpdateOne) Where(ps ...predicate.Pkg) *PkgUpdateOne {
	puo.mutation.Where(ps...)
	return puo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (puo *PkgUpdateOne) Select(field string, fields ...string) *PkgUpdateOne {
	puo.fields = append([]string{field}, fields...)
	return puo
}

// Save executes the query and returns the updated Pkg entity.
func (puo *PkgUpdateOne) Save(ctx context.Context) (*Pkg, error) {
	return withHooks(ctx, puo.sqlSave, puo.mutation, puo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (puo *PkgUpdateOne) SaveX(ctx context.Context) *Pkg {
	node, err := puo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (puo *PkgUpdateOne) Exec(ctx context.Context) error {
	_, err := puo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (puo *PkgUpdateOne) ExecX(ctx context.Context) {
	if err := puo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (puo *PkgUpdateOne) check() error {
	if v, ok := puo.mutation.Status(); ok {
		if err := pkg.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`db: validator failed for field "Pkg.status": %w`, err)}
		}
	}
	return nil
}

func (puo *PkgUpdateOne) sqlSave(ctx context.Context) (_node *Pkg, err error) {
	if err := puo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(pkg.Table, pkg.Columns, sqlgraph.NewFieldSpec(pkg.FieldID, field.TypeInt))
	id, ok := puo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`db: missing "Pkg.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := puo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, pkg.FieldID)
		for _, f := range fields {
			if !pkg.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("db: invalid field %q for query", f)}
			}
			if f != pkg.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := puo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := puo.mutation.Name(); ok {
		_spec.SetField(pkg.FieldName, field.TypeString, value)
	}
	if value, ok := puo.mutation.AipID(); ok {
		_spec.SetField(pkg.FieldAipID, field.TypeUUID, value)
	}
	if value, ok := puo.mutation.Status(); ok {
		_spec.SetField(pkg.FieldStatus, field.TypeEnum, value)
	}
	if value, ok := puo.mutation.ObjectKey(); ok {
		_spec.SetField(pkg.FieldObjectKey, field.TypeUUID, value)
	}
	if puo.mutation.LocationCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := puo.mutation.LocationIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &Pkg{config: puo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, puo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{pkg.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	puo.mutation.done = true
	return _node, nil
}
