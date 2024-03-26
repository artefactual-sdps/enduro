// Code generated by ent, DO NOT EDIT.

package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db/predicate"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db/preservationaction"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db/preservationtask"
	"github.com/google/uuid"
)

// PreservationTaskUpdate is the builder for updating PreservationTask entities.
type PreservationTaskUpdate struct {
	config
	hooks    []Hook
	mutation *PreservationTaskMutation
}

// Where appends a list predicates to the PreservationTaskUpdate builder.
func (ptu *PreservationTaskUpdate) Where(ps ...predicate.PreservationTask) *PreservationTaskUpdate {
	ptu.mutation.Where(ps...)
	return ptu
}

// SetTaskID sets the "task_id" field.
func (ptu *PreservationTaskUpdate) SetTaskID(u uuid.UUID) *PreservationTaskUpdate {
	ptu.mutation.SetTaskID(u)
	return ptu
}

// SetNillableTaskID sets the "task_id" field if the given value is not nil.
func (ptu *PreservationTaskUpdate) SetNillableTaskID(u *uuid.UUID) *PreservationTaskUpdate {
	if u != nil {
		ptu.SetTaskID(*u)
	}
	return ptu
}

// SetName sets the "name" field.
func (ptu *PreservationTaskUpdate) SetName(s string) *PreservationTaskUpdate {
	ptu.mutation.SetName(s)
	return ptu
}

// SetNillableName sets the "name" field if the given value is not nil.
func (ptu *PreservationTaskUpdate) SetNillableName(s *string) *PreservationTaskUpdate {
	if s != nil {
		ptu.SetName(*s)
	}
	return ptu
}

// SetStatus sets the "status" field.
func (ptu *PreservationTaskUpdate) SetStatus(i int8) *PreservationTaskUpdate {
	ptu.mutation.ResetStatus()
	ptu.mutation.SetStatus(i)
	return ptu
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (ptu *PreservationTaskUpdate) SetNillableStatus(i *int8) *PreservationTaskUpdate {
	if i != nil {
		ptu.SetStatus(*i)
	}
	return ptu
}

// AddStatus adds i to the "status" field.
func (ptu *PreservationTaskUpdate) AddStatus(i int8) *PreservationTaskUpdate {
	ptu.mutation.AddStatus(i)
	return ptu
}

// SetStartedAt sets the "started_at" field.
func (ptu *PreservationTaskUpdate) SetStartedAt(t time.Time) *PreservationTaskUpdate {
	ptu.mutation.SetStartedAt(t)
	return ptu
}

// SetNillableStartedAt sets the "started_at" field if the given value is not nil.
func (ptu *PreservationTaskUpdate) SetNillableStartedAt(t *time.Time) *PreservationTaskUpdate {
	if t != nil {
		ptu.SetStartedAt(*t)
	}
	return ptu
}

// ClearStartedAt clears the value of the "started_at" field.
func (ptu *PreservationTaskUpdate) ClearStartedAt() *PreservationTaskUpdate {
	ptu.mutation.ClearStartedAt()
	return ptu
}

// SetCompletedAt sets the "completed_at" field.
func (ptu *PreservationTaskUpdate) SetCompletedAt(t time.Time) *PreservationTaskUpdate {
	ptu.mutation.SetCompletedAt(t)
	return ptu
}

// SetNillableCompletedAt sets the "completed_at" field if the given value is not nil.
func (ptu *PreservationTaskUpdate) SetNillableCompletedAt(t *time.Time) *PreservationTaskUpdate {
	if t != nil {
		ptu.SetCompletedAt(*t)
	}
	return ptu
}

// ClearCompletedAt clears the value of the "completed_at" field.
func (ptu *PreservationTaskUpdate) ClearCompletedAt() *PreservationTaskUpdate {
	ptu.mutation.ClearCompletedAt()
	return ptu
}

// SetNote sets the "note" field.
func (ptu *PreservationTaskUpdate) SetNote(s string) *PreservationTaskUpdate {
	ptu.mutation.SetNote(s)
	return ptu
}

// SetNillableNote sets the "note" field if the given value is not nil.
func (ptu *PreservationTaskUpdate) SetNillableNote(s *string) *PreservationTaskUpdate {
	if s != nil {
		ptu.SetNote(*s)
	}
	return ptu
}

// SetPreservationActionID sets the "preservation_action_id" field.
func (ptu *PreservationTaskUpdate) SetPreservationActionID(i int) *PreservationTaskUpdate {
	ptu.mutation.SetPreservationActionID(i)
	return ptu
}

// SetNillablePreservationActionID sets the "preservation_action_id" field if the given value is not nil.
func (ptu *PreservationTaskUpdate) SetNillablePreservationActionID(i *int) *PreservationTaskUpdate {
	if i != nil {
		ptu.SetPreservationActionID(*i)
	}
	return ptu
}

// SetActionID sets the "action" edge to the PreservationAction entity by ID.
func (ptu *PreservationTaskUpdate) SetActionID(id int) *PreservationTaskUpdate {
	ptu.mutation.SetActionID(id)
	return ptu
}

// SetAction sets the "action" edge to the PreservationAction entity.
func (ptu *PreservationTaskUpdate) SetAction(p *PreservationAction) *PreservationTaskUpdate {
	return ptu.SetActionID(p.ID)
}

// Mutation returns the PreservationTaskMutation object of the builder.
func (ptu *PreservationTaskUpdate) Mutation() *PreservationTaskMutation {
	return ptu.mutation
}

// ClearAction clears the "action" edge to the PreservationAction entity.
func (ptu *PreservationTaskUpdate) ClearAction() *PreservationTaskUpdate {
	ptu.mutation.ClearAction()
	return ptu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (ptu *PreservationTaskUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, ptu.sqlSave, ptu.mutation, ptu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (ptu *PreservationTaskUpdate) SaveX(ctx context.Context) int {
	affected, err := ptu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (ptu *PreservationTaskUpdate) Exec(ctx context.Context) error {
	_, err := ptu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ptu *PreservationTaskUpdate) ExecX(ctx context.Context) {
	if err := ptu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (ptu *PreservationTaskUpdate) check() error {
	if v, ok := ptu.mutation.PreservationActionID(); ok {
		if err := preservationtask.PreservationActionIDValidator(v); err != nil {
			return &ValidationError{Name: "preservation_action_id", err: fmt.Errorf(`db: validator failed for field "PreservationTask.preservation_action_id": %w`, err)}
		}
	}
	if _, ok := ptu.mutation.ActionID(); ptu.mutation.ActionCleared() && !ok {
		return errors.New(`db: clearing a required unique edge "PreservationTask.action"`)
	}
	return nil
}

func (ptu *PreservationTaskUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := ptu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(preservationtask.Table, preservationtask.Columns, sqlgraph.NewFieldSpec(preservationtask.FieldID, field.TypeInt))
	if ps := ptu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := ptu.mutation.TaskID(); ok {
		_spec.SetField(preservationtask.FieldTaskID, field.TypeUUID, value)
	}
	if value, ok := ptu.mutation.Name(); ok {
		_spec.SetField(preservationtask.FieldName, field.TypeString, value)
	}
	if value, ok := ptu.mutation.Status(); ok {
		_spec.SetField(preservationtask.FieldStatus, field.TypeInt8, value)
	}
	if value, ok := ptu.mutation.AddedStatus(); ok {
		_spec.AddField(preservationtask.FieldStatus, field.TypeInt8, value)
	}
	if value, ok := ptu.mutation.StartedAt(); ok {
		_spec.SetField(preservationtask.FieldStartedAt, field.TypeTime, value)
	}
	if ptu.mutation.StartedAtCleared() {
		_spec.ClearField(preservationtask.FieldStartedAt, field.TypeTime)
	}
	if value, ok := ptu.mutation.CompletedAt(); ok {
		_spec.SetField(preservationtask.FieldCompletedAt, field.TypeTime, value)
	}
	if ptu.mutation.CompletedAtCleared() {
		_spec.ClearField(preservationtask.FieldCompletedAt, field.TypeTime)
	}
	if value, ok := ptu.mutation.Note(); ok {
		_spec.SetField(preservationtask.FieldNote, field.TypeString, value)
	}
	if ptu.mutation.ActionCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   preservationtask.ActionTable,
			Columns: []string{preservationtask.ActionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(preservationaction.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ptu.mutation.ActionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   preservationtask.ActionTable,
			Columns: []string{preservationtask.ActionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(preservationaction.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, ptu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{preservationtask.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	ptu.mutation.done = true
	return n, nil
}

// PreservationTaskUpdateOne is the builder for updating a single PreservationTask entity.
type PreservationTaskUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *PreservationTaskMutation
}

// SetTaskID sets the "task_id" field.
func (ptuo *PreservationTaskUpdateOne) SetTaskID(u uuid.UUID) *PreservationTaskUpdateOne {
	ptuo.mutation.SetTaskID(u)
	return ptuo
}

// SetNillableTaskID sets the "task_id" field if the given value is not nil.
func (ptuo *PreservationTaskUpdateOne) SetNillableTaskID(u *uuid.UUID) *PreservationTaskUpdateOne {
	if u != nil {
		ptuo.SetTaskID(*u)
	}
	return ptuo
}

// SetName sets the "name" field.
func (ptuo *PreservationTaskUpdateOne) SetName(s string) *PreservationTaskUpdateOne {
	ptuo.mutation.SetName(s)
	return ptuo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (ptuo *PreservationTaskUpdateOne) SetNillableName(s *string) *PreservationTaskUpdateOne {
	if s != nil {
		ptuo.SetName(*s)
	}
	return ptuo
}

// SetStatus sets the "status" field.
func (ptuo *PreservationTaskUpdateOne) SetStatus(i int8) *PreservationTaskUpdateOne {
	ptuo.mutation.ResetStatus()
	ptuo.mutation.SetStatus(i)
	return ptuo
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (ptuo *PreservationTaskUpdateOne) SetNillableStatus(i *int8) *PreservationTaskUpdateOne {
	if i != nil {
		ptuo.SetStatus(*i)
	}
	return ptuo
}

// AddStatus adds i to the "status" field.
func (ptuo *PreservationTaskUpdateOne) AddStatus(i int8) *PreservationTaskUpdateOne {
	ptuo.mutation.AddStatus(i)
	return ptuo
}

// SetStartedAt sets the "started_at" field.
func (ptuo *PreservationTaskUpdateOne) SetStartedAt(t time.Time) *PreservationTaskUpdateOne {
	ptuo.mutation.SetStartedAt(t)
	return ptuo
}

// SetNillableStartedAt sets the "started_at" field if the given value is not nil.
func (ptuo *PreservationTaskUpdateOne) SetNillableStartedAt(t *time.Time) *PreservationTaskUpdateOne {
	if t != nil {
		ptuo.SetStartedAt(*t)
	}
	return ptuo
}

// ClearStartedAt clears the value of the "started_at" field.
func (ptuo *PreservationTaskUpdateOne) ClearStartedAt() *PreservationTaskUpdateOne {
	ptuo.mutation.ClearStartedAt()
	return ptuo
}

// SetCompletedAt sets the "completed_at" field.
func (ptuo *PreservationTaskUpdateOne) SetCompletedAt(t time.Time) *PreservationTaskUpdateOne {
	ptuo.mutation.SetCompletedAt(t)
	return ptuo
}

// SetNillableCompletedAt sets the "completed_at" field if the given value is not nil.
func (ptuo *PreservationTaskUpdateOne) SetNillableCompletedAt(t *time.Time) *PreservationTaskUpdateOne {
	if t != nil {
		ptuo.SetCompletedAt(*t)
	}
	return ptuo
}

// ClearCompletedAt clears the value of the "completed_at" field.
func (ptuo *PreservationTaskUpdateOne) ClearCompletedAt() *PreservationTaskUpdateOne {
	ptuo.mutation.ClearCompletedAt()
	return ptuo
}

// SetNote sets the "note" field.
func (ptuo *PreservationTaskUpdateOne) SetNote(s string) *PreservationTaskUpdateOne {
	ptuo.mutation.SetNote(s)
	return ptuo
}

// SetNillableNote sets the "note" field if the given value is not nil.
func (ptuo *PreservationTaskUpdateOne) SetNillableNote(s *string) *PreservationTaskUpdateOne {
	if s != nil {
		ptuo.SetNote(*s)
	}
	return ptuo
}

// SetPreservationActionID sets the "preservation_action_id" field.
func (ptuo *PreservationTaskUpdateOne) SetPreservationActionID(i int) *PreservationTaskUpdateOne {
	ptuo.mutation.SetPreservationActionID(i)
	return ptuo
}

// SetNillablePreservationActionID sets the "preservation_action_id" field if the given value is not nil.
func (ptuo *PreservationTaskUpdateOne) SetNillablePreservationActionID(i *int) *PreservationTaskUpdateOne {
	if i != nil {
		ptuo.SetPreservationActionID(*i)
	}
	return ptuo
}

// SetActionID sets the "action" edge to the PreservationAction entity by ID.
func (ptuo *PreservationTaskUpdateOne) SetActionID(id int) *PreservationTaskUpdateOne {
	ptuo.mutation.SetActionID(id)
	return ptuo
}

// SetAction sets the "action" edge to the PreservationAction entity.
func (ptuo *PreservationTaskUpdateOne) SetAction(p *PreservationAction) *PreservationTaskUpdateOne {
	return ptuo.SetActionID(p.ID)
}

// Mutation returns the PreservationTaskMutation object of the builder.
func (ptuo *PreservationTaskUpdateOne) Mutation() *PreservationTaskMutation {
	return ptuo.mutation
}

// ClearAction clears the "action" edge to the PreservationAction entity.
func (ptuo *PreservationTaskUpdateOne) ClearAction() *PreservationTaskUpdateOne {
	ptuo.mutation.ClearAction()
	return ptuo
}

// Where appends a list predicates to the PreservationTaskUpdate builder.
func (ptuo *PreservationTaskUpdateOne) Where(ps ...predicate.PreservationTask) *PreservationTaskUpdateOne {
	ptuo.mutation.Where(ps...)
	return ptuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (ptuo *PreservationTaskUpdateOne) Select(field string, fields ...string) *PreservationTaskUpdateOne {
	ptuo.fields = append([]string{field}, fields...)
	return ptuo
}

// Save executes the query and returns the updated PreservationTask entity.
func (ptuo *PreservationTaskUpdateOne) Save(ctx context.Context) (*PreservationTask, error) {
	return withHooks(ctx, ptuo.sqlSave, ptuo.mutation, ptuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (ptuo *PreservationTaskUpdateOne) SaveX(ctx context.Context) *PreservationTask {
	node, err := ptuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (ptuo *PreservationTaskUpdateOne) Exec(ctx context.Context) error {
	_, err := ptuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ptuo *PreservationTaskUpdateOne) ExecX(ctx context.Context) {
	if err := ptuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (ptuo *PreservationTaskUpdateOne) check() error {
	if v, ok := ptuo.mutation.PreservationActionID(); ok {
		if err := preservationtask.PreservationActionIDValidator(v); err != nil {
			return &ValidationError{Name: "preservation_action_id", err: fmt.Errorf(`db: validator failed for field "PreservationTask.preservation_action_id": %w`, err)}
		}
	}
	if _, ok := ptuo.mutation.ActionID(); ptuo.mutation.ActionCleared() && !ok {
		return errors.New(`db: clearing a required unique edge "PreservationTask.action"`)
	}
	return nil
}

func (ptuo *PreservationTaskUpdateOne) sqlSave(ctx context.Context) (_node *PreservationTask, err error) {
	if err := ptuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(preservationtask.Table, preservationtask.Columns, sqlgraph.NewFieldSpec(preservationtask.FieldID, field.TypeInt))
	id, ok := ptuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`db: missing "PreservationTask.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := ptuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, preservationtask.FieldID)
		for _, f := range fields {
			if !preservationtask.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("db: invalid field %q for query", f)}
			}
			if f != preservationtask.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := ptuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := ptuo.mutation.TaskID(); ok {
		_spec.SetField(preservationtask.FieldTaskID, field.TypeUUID, value)
	}
	if value, ok := ptuo.mutation.Name(); ok {
		_spec.SetField(preservationtask.FieldName, field.TypeString, value)
	}
	if value, ok := ptuo.mutation.Status(); ok {
		_spec.SetField(preservationtask.FieldStatus, field.TypeInt8, value)
	}
	if value, ok := ptuo.mutation.AddedStatus(); ok {
		_spec.AddField(preservationtask.FieldStatus, field.TypeInt8, value)
	}
	if value, ok := ptuo.mutation.StartedAt(); ok {
		_spec.SetField(preservationtask.FieldStartedAt, field.TypeTime, value)
	}
	if ptuo.mutation.StartedAtCleared() {
		_spec.ClearField(preservationtask.FieldStartedAt, field.TypeTime)
	}
	if value, ok := ptuo.mutation.CompletedAt(); ok {
		_spec.SetField(preservationtask.FieldCompletedAt, field.TypeTime, value)
	}
	if ptuo.mutation.CompletedAtCleared() {
		_spec.ClearField(preservationtask.FieldCompletedAt, field.TypeTime)
	}
	if value, ok := ptuo.mutation.Note(); ok {
		_spec.SetField(preservationtask.FieldNote, field.TypeString, value)
	}
	if ptuo.mutation.ActionCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   preservationtask.ActionTable,
			Columns: []string{preservationtask.ActionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(preservationaction.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ptuo.mutation.ActionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   preservationtask.ActionTable,
			Columns: []string{preservationtask.ActionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(preservationaction.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &PreservationTask{config: ptuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, ptuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{preservationtask.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	ptuo.mutation.done = true
	return _node, nil
}
