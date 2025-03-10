// Code generated by ent, DO NOT EDIT.

package db

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/aip"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/location"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/predicate"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
	"github.com/google/uuid"
)

// LocationUpdate is the builder for updating Location entities.
type LocationUpdate struct {
	config
	hooks    []Hook
	mutation *LocationMutation
}

// Where appends a list predicates to the LocationUpdate builder.
func (lu *LocationUpdate) Where(ps ...predicate.Location) *LocationUpdate {
	lu.mutation.Where(ps...)
	return lu
}

// SetName sets the "name" field.
func (lu *LocationUpdate) SetName(s string) *LocationUpdate {
	lu.mutation.SetName(s)
	return lu
}

// SetNillableName sets the "name" field if the given value is not nil.
func (lu *LocationUpdate) SetNillableName(s *string) *LocationUpdate {
	if s != nil {
		lu.SetName(*s)
	}
	return lu
}

// SetDescription sets the "description" field.
func (lu *LocationUpdate) SetDescription(s string) *LocationUpdate {
	lu.mutation.SetDescription(s)
	return lu
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (lu *LocationUpdate) SetNillableDescription(s *string) *LocationUpdate {
	if s != nil {
		lu.SetDescription(*s)
	}
	return lu
}

// SetSource sets the "source" field.
func (lu *LocationUpdate) SetSource(es enums.LocationSource) *LocationUpdate {
	lu.mutation.SetSource(es)
	return lu
}

// SetNillableSource sets the "source" field if the given value is not nil.
func (lu *LocationUpdate) SetNillableSource(es *enums.LocationSource) *LocationUpdate {
	if es != nil {
		lu.SetSource(*es)
	}
	return lu
}

// SetPurpose sets the "purpose" field.
func (lu *LocationUpdate) SetPurpose(ep enums.LocationPurpose) *LocationUpdate {
	lu.mutation.SetPurpose(ep)
	return lu
}

// SetNillablePurpose sets the "purpose" field if the given value is not nil.
func (lu *LocationUpdate) SetNillablePurpose(ep *enums.LocationPurpose) *LocationUpdate {
	if ep != nil {
		lu.SetPurpose(*ep)
	}
	return lu
}

// SetUUID sets the "uuid" field.
func (lu *LocationUpdate) SetUUID(u uuid.UUID) *LocationUpdate {
	lu.mutation.SetUUID(u)
	return lu
}

// SetNillableUUID sets the "uuid" field if the given value is not nil.
func (lu *LocationUpdate) SetNillableUUID(u *uuid.UUID) *LocationUpdate {
	if u != nil {
		lu.SetUUID(*u)
	}
	return lu
}

// SetConfig sets the "config" field.
func (lu *LocationUpdate) SetConfig(tc types.LocationConfig) *LocationUpdate {
	lu.mutation.SetConfig(tc)
	return lu
}

// SetNillableConfig sets the "config" field if the given value is not nil.
func (lu *LocationUpdate) SetNillableConfig(tc *types.LocationConfig) *LocationUpdate {
	if tc != nil {
		lu.SetConfig(*tc)
	}
	return lu
}

// AddAipIDs adds the "aips" edge to the AIP entity by IDs.
func (lu *LocationUpdate) AddAipIDs(ids ...int) *LocationUpdate {
	lu.mutation.AddAipIDs(ids...)
	return lu
}

// AddAips adds the "aips" edges to the AIP entity.
func (lu *LocationUpdate) AddAips(a ...*AIP) *LocationUpdate {
	ids := make([]int, len(a))
	for i := range a {
		ids[i] = a[i].ID
	}
	return lu.AddAipIDs(ids...)
}

// Mutation returns the LocationMutation object of the builder.
func (lu *LocationUpdate) Mutation() *LocationMutation {
	return lu.mutation
}

// ClearAips clears all "aips" edges to the AIP entity.
func (lu *LocationUpdate) ClearAips() *LocationUpdate {
	lu.mutation.ClearAips()
	return lu
}

// RemoveAipIDs removes the "aips" edge to AIP entities by IDs.
func (lu *LocationUpdate) RemoveAipIDs(ids ...int) *LocationUpdate {
	lu.mutation.RemoveAipIDs(ids...)
	return lu
}

// RemoveAips removes "aips" edges to AIP entities.
func (lu *LocationUpdate) RemoveAips(a ...*AIP) *LocationUpdate {
	ids := make([]int, len(a))
	for i := range a {
		ids[i] = a[i].ID
	}
	return lu.RemoveAipIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (lu *LocationUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, lu.sqlSave, lu.mutation, lu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (lu *LocationUpdate) SaveX(ctx context.Context) int {
	affected, err := lu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (lu *LocationUpdate) Exec(ctx context.Context) error {
	_, err := lu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (lu *LocationUpdate) ExecX(ctx context.Context) {
	if err := lu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (lu *LocationUpdate) check() error {
	if v, ok := lu.mutation.Source(); ok {
		if err := location.SourceValidator(v); err != nil {
			return &ValidationError{Name: "source", err: fmt.Errorf(`db: validator failed for field "Location.source": %w`, err)}
		}
	}
	if v, ok := lu.mutation.Purpose(); ok {
		if err := location.PurposeValidator(v); err != nil {
			return &ValidationError{Name: "purpose", err: fmt.Errorf(`db: validator failed for field "Location.purpose": %w`, err)}
		}
	}
	return nil
}

func (lu *LocationUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := lu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(location.Table, location.Columns, sqlgraph.NewFieldSpec(location.FieldID, field.TypeInt))
	if ps := lu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := lu.mutation.Name(); ok {
		_spec.SetField(location.FieldName, field.TypeString, value)
	}
	if value, ok := lu.mutation.Description(); ok {
		_spec.SetField(location.FieldDescription, field.TypeString, value)
	}
	if value, ok := lu.mutation.Source(); ok {
		_spec.SetField(location.FieldSource, field.TypeEnum, value)
	}
	if value, ok := lu.mutation.Purpose(); ok {
		_spec.SetField(location.FieldPurpose, field.TypeEnum, value)
	}
	if value, ok := lu.mutation.UUID(); ok {
		_spec.SetField(location.FieldUUID, field.TypeUUID, value)
	}
	if value, ok := lu.mutation.Config(); ok {
		_spec.SetField(location.FieldConfig, field.TypeJSON, value)
	}
	if lu.mutation.AipsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := lu.mutation.RemovedAipsIDs(); len(nodes) > 0 && !lu.mutation.AipsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := lu.mutation.AipsIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, lu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{location.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	lu.mutation.done = true
	return n, nil
}

// LocationUpdateOne is the builder for updating a single Location entity.
type LocationUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *LocationMutation
}

// SetName sets the "name" field.
func (luo *LocationUpdateOne) SetName(s string) *LocationUpdateOne {
	luo.mutation.SetName(s)
	return luo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (luo *LocationUpdateOne) SetNillableName(s *string) *LocationUpdateOne {
	if s != nil {
		luo.SetName(*s)
	}
	return luo
}

// SetDescription sets the "description" field.
func (luo *LocationUpdateOne) SetDescription(s string) *LocationUpdateOne {
	luo.mutation.SetDescription(s)
	return luo
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (luo *LocationUpdateOne) SetNillableDescription(s *string) *LocationUpdateOne {
	if s != nil {
		luo.SetDescription(*s)
	}
	return luo
}

// SetSource sets the "source" field.
func (luo *LocationUpdateOne) SetSource(es enums.LocationSource) *LocationUpdateOne {
	luo.mutation.SetSource(es)
	return luo
}

// SetNillableSource sets the "source" field if the given value is not nil.
func (luo *LocationUpdateOne) SetNillableSource(es *enums.LocationSource) *LocationUpdateOne {
	if es != nil {
		luo.SetSource(*es)
	}
	return luo
}

// SetPurpose sets the "purpose" field.
func (luo *LocationUpdateOne) SetPurpose(ep enums.LocationPurpose) *LocationUpdateOne {
	luo.mutation.SetPurpose(ep)
	return luo
}

// SetNillablePurpose sets the "purpose" field if the given value is not nil.
func (luo *LocationUpdateOne) SetNillablePurpose(ep *enums.LocationPurpose) *LocationUpdateOne {
	if ep != nil {
		luo.SetPurpose(*ep)
	}
	return luo
}

// SetUUID sets the "uuid" field.
func (luo *LocationUpdateOne) SetUUID(u uuid.UUID) *LocationUpdateOne {
	luo.mutation.SetUUID(u)
	return luo
}

// SetNillableUUID sets the "uuid" field if the given value is not nil.
func (luo *LocationUpdateOne) SetNillableUUID(u *uuid.UUID) *LocationUpdateOne {
	if u != nil {
		luo.SetUUID(*u)
	}
	return luo
}

// SetConfig sets the "config" field.
func (luo *LocationUpdateOne) SetConfig(tc types.LocationConfig) *LocationUpdateOne {
	luo.mutation.SetConfig(tc)
	return luo
}

// SetNillableConfig sets the "config" field if the given value is not nil.
func (luo *LocationUpdateOne) SetNillableConfig(tc *types.LocationConfig) *LocationUpdateOne {
	if tc != nil {
		luo.SetConfig(*tc)
	}
	return luo
}

// AddAipIDs adds the "aips" edge to the AIP entity by IDs.
func (luo *LocationUpdateOne) AddAipIDs(ids ...int) *LocationUpdateOne {
	luo.mutation.AddAipIDs(ids...)
	return luo
}

// AddAips adds the "aips" edges to the AIP entity.
func (luo *LocationUpdateOne) AddAips(a ...*AIP) *LocationUpdateOne {
	ids := make([]int, len(a))
	for i := range a {
		ids[i] = a[i].ID
	}
	return luo.AddAipIDs(ids...)
}

// Mutation returns the LocationMutation object of the builder.
func (luo *LocationUpdateOne) Mutation() *LocationMutation {
	return luo.mutation
}

// ClearAips clears all "aips" edges to the AIP entity.
func (luo *LocationUpdateOne) ClearAips() *LocationUpdateOne {
	luo.mutation.ClearAips()
	return luo
}

// RemoveAipIDs removes the "aips" edge to AIP entities by IDs.
func (luo *LocationUpdateOne) RemoveAipIDs(ids ...int) *LocationUpdateOne {
	luo.mutation.RemoveAipIDs(ids...)
	return luo
}

// RemoveAips removes "aips" edges to AIP entities.
func (luo *LocationUpdateOne) RemoveAips(a ...*AIP) *LocationUpdateOne {
	ids := make([]int, len(a))
	for i := range a {
		ids[i] = a[i].ID
	}
	return luo.RemoveAipIDs(ids...)
}

// Where appends a list predicates to the LocationUpdate builder.
func (luo *LocationUpdateOne) Where(ps ...predicate.Location) *LocationUpdateOne {
	luo.mutation.Where(ps...)
	return luo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (luo *LocationUpdateOne) Select(field string, fields ...string) *LocationUpdateOne {
	luo.fields = append([]string{field}, fields...)
	return luo
}

// Save executes the query and returns the updated Location entity.
func (luo *LocationUpdateOne) Save(ctx context.Context) (*Location, error) {
	return withHooks(ctx, luo.sqlSave, luo.mutation, luo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (luo *LocationUpdateOne) SaveX(ctx context.Context) *Location {
	node, err := luo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (luo *LocationUpdateOne) Exec(ctx context.Context) error {
	_, err := luo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (luo *LocationUpdateOne) ExecX(ctx context.Context) {
	if err := luo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (luo *LocationUpdateOne) check() error {
	if v, ok := luo.mutation.Source(); ok {
		if err := location.SourceValidator(v); err != nil {
			return &ValidationError{Name: "source", err: fmt.Errorf(`db: validator failed for field "Location.source": %w`, err)}
		}
	}
	if v, ok := luo.mutation.Purpose(); ok {
		if err := location.PurposeValidator(v); err != nil {
			return &ValidationError{Name: "purpose", err: fmt.Errorf(`db: validator failed for field "Location.purpose": %w`, err)}
		}
	}
	return nil
}

func (luo *LocationUpdateOne) sqlSave(ctx context.Context) (_node *Location, err error) {
	if err := luo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(location.Table, location.Columns, sqlgraph.NewFieldSpec(location.FieldID, field.TypeInt))
	id, ok := luo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`db: missing "Location.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := luo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, location.FieldID)
		for _, f := range fields {
			if !location.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("db: invalid field %q for query", f)}
			}
			if f != location.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := luo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := luo.mutation.Name(); ok {
		_spec.SetField(location.FieldName, field.TypeString, value)
	}
	if value, ok := luo.mutation.Description(); ok {
		_spec.SetField(location.FieldDescription, field.TypeString, value)
	}
	if value, ok := luo.mutation.Source(); ok {
		_spec.SetField(location.FieldSource, field.TypeEnum, value)
	}
	if value, ok := luo.mutation.Purpose(); ok {
		_spec.SetField(location.FieldPurpose, field.TypeEnum, value)
	}
	if value, ok := luo.mutation.UUID(); ok {
		_spec.SetField(location.FieldUUID, field.TypeUUID, value)
	}
	if value, ok := luo.mutation.Config(); ok {
		_spec.SetField(location.FieldConfig, field.TypeJSON, value)
	}
	if luo.mutation.AipsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := luo.mutation.RemovedAipsIDs(); len(nodes) > 0 && !luo.mutation.AipsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := luo.mutation.AipsIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &Location{config: luo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, luo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{location.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	luo.mutation.done = true
	return _node, nil
}
