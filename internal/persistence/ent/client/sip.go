package entclient

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/entfilter"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db/sip"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db/user"
)

// CreateSIP creates and persists a new SIP using the values from s
// then returns the updated SIP.
//
// The input SIP "ID" and "CreatedAt" values are ignored; the stored SIP
// "ID" is generated by the persistence implementation and "CreatedAt" is always
// set to the current time.
func (c *client) CreateSIP(ctx context.Context, s *datatypes.SIP) error {
	// Validate required fields.
	if s.UUID == uuid.Nil {
		return newRequiredFieldError("UUID")
	}
	if s.Name == "" {
		return newRequiredFieldError("Name")
	}

	tx, err := c.ent.BeginTx(ctx, nil)
	if err != nil {
		return newDBErrorWithDetails(err, "create SIP")
	}

	q := tx.SIP.Create().
		SetUUID(s.UUID).
		SetName(s.Name).
		SetStatus(s.Status).
		SetCreatedAt(time.Now())

	// Add optional fields.
	if s.AIPID.Valid {
		q.SetAipID(s.AIPID.UUID)
	}
	if s.StartedAt.Valid {
		q.SetStartedAt(s.StartedAt.Time)
	}
	if s.CompletedAt.Valid {
		q.SetCompletedAt(s.CompletedAt.Time)
	}

	// If Uploader.UUID is set, find the user and link it to the SIP.
	var uploader *datatypes.Uploader
	if s.Uploader != nil && s.Uploader.UUID != uuid.Nil {
		u, err := tx.User.Query().Where(user.UUID(s.Uploader.UUID)).Only(ctx)
		if err != nil {
			return rollback(tx, newDBErrorWithDetails(err, "create SIP"))
		}
		q.SetUser(u)

		uploader = &datatypes.Uploader{
			UUID:  u.UUID,
			Email: u.Email,
			Name:  u.Name,
		}
	}

	// Save the SIP.
	dbs, err := q.Save(ctx)
	if err != nil {
		return rollback(tx, newDBErrorWithDetails(err, "create SIP"))
	}
	if err = tx.Commit(); err != nil {
		return rollback(tx, newDBError(err))
	}

	// Update SIP with DB data, to get generated values (e.g. ID).
	*s = *convertSIP(dbs)

	// Manually set the uploader data because the dbs result doesn't include the
	// user edge.
	if uploader != nil {
		s.Uploader = uploader
	}

	return nil
}

// UpdateSIP updates the persisted SIP identified by id using the
// updater function, then returns the updated SIP.
//
// The SIP "ID", "UUID", "CreatedAt", and "UploaderID" values can not be updated
// with this method.
func (c *client) UpdateSIP(
	ctx context.Context,
	id uuid.UUID,
	updater persistence.SIPUpdater,
) (*datatypes.SIP, error) {
	tx, err := c.ent.BeginTx(ctx, nil)
	if err != nil {
		return nil, newDBError(err)
	}

	// Get the current SIP data from the database.
	dbs, err := tx.SIP.Query().
		Where(sip.UUID(id)).
		WithUser(func(q *db.UserQuery) {
			q.Select(user.FieldUUID)
			q.Select(user.FieldEmail)
			q.Select(user.FieldName)
		}).
		Only(ctx)
	if err != nil {
		return nil, rollback(tx, newDBError(err))
	}

	// Keep database ID in case it's changed by the updater.
	dbID := dbs.ID

	// Get the uploader data so we can set it on the returned SIP later.
	var uploader *datatypes.Uploader
	if dbs.Edges.User != nil {
		uploader = &datatypes.Uploader{
			UUID:  dbs.Edges.User.UUID,
			Email: dbs.Edges.User.Email,
			Name:  dbs.Edges.User.Name,
		}
	}

	// Get an updated datatypes.SIP from the updater function.
	up, err := updater(convertSIP(dbs))
	if err != nil {
		return nil, rollback(tx, newUpdaterError(err))
	}

	// Save the updated SIP data to the database.
	q := tx.SIP.UpdateOneID(dbID).SetName(up.Name)

	// Validate columns.
	if up.Status.IsValid() {
		q.SetStatus(up.Status)
	}

	// Set optional column values.
	if up.AIPID.Valid {
		q.SetAipID(up.AIPID.UUID)
	}
	if up.StartedAt.Valid {
		q.SetStartedAt(up.StartedAt.Time)
	}
	if up.CompletedAt.Valid {
		q.SetCompletedAt(up.CompletedAt.Time)
	}
	if up.FailedAs.IsValid() {
		q.SetFailedAs(up.FailedAs)
	}
	if up.FailedKey != "" {
		q.SetFailedKey(up.FailedKey)
	}

	// Save changes.
	dbs, err = q.Save(ctx)
	if err != nil {
		return nil, rollback(tx, newDBError(err))
	}
	if err = tx.Commit(); err != nil {
		return nil, rollback(tx, newDBError(err))
	}

	s := convertSIP(dbs)

	// Set the uploader data on the returned SIP.
	if uploader != nil {
		s.Uploader = uploader
	}

	return s, nil
}

// DeleteSIP deletes the persisted SIP identified by id.
func (c *client) DeleteSIP(ctx context.Context, id int) error {
	if err := c.ent.SIP.DeleteOneID(id).Exec(ctx); err != nil {
		return newDBErrorWithDetails(err, "delete SIP")
	}

	return nil
}

// ReadSIP returns the SIP identified by id.
func (c *client) ReadSIP(ctx context.Context, id uuid.UUID) (*datatypes.SIP, error) {
	s, err := c.ent.SIP.Query().
		Where(sip.UUID(id)).
		WithUser(func(q *db.UserQuery) {
			q.Select(user.FieldUUID)
			q.Select(user.FieldEmail)
			q.Select(user.FieldName)
		}).
		Only(ctx)
	if err != nil {
		return nil, newDBError(err)
	}

	return convertSIP(s), nil
}

// ListSIPs returns a slice of SIPs filtered according to f.
func (c *client) ListSIPs(ctx context.Context, f *persistence.SIPFilter) (
	[]*datatypes.SIP, *persistence.Page, error,
) {
	res := []*datatypes.SIP{}

	if f == nil {
		f = &persistence.SIPFilter{}
	}

	q := c.ent.SIP.Query().WithUser(func(q *db.UserQuery) {
		q.Select(user.FieldUUID)
		q.Select(user.FieldEmail)
		q.Select(user.FieldName)
	})

	if f.UploaderID != nil {
		q.Where(sip.HasUserWith(user.UUID(*f.UploaderID)))
	}

	page, whole := filterSIPs(q, f)

	r, err := page.All(ctx)
	if err != nil {
		return nil, nil, newDBError(err)
	}

	for _, i := range r {
		res = append(res, convertSIP(i))
	}

	total, err := whole.Count(ctx)
	if err != nil {
		return nil, nil, newDBError(err)
	}

	pp := &persistence.Page{
		Limit:  f.Limit,
		Offset: f.Offset,
		Total:  total,
	}

	return res, pp, err
}

// filterSIPs applies the SIP filter f to the query q.
func filterSIPs(q *db.SIPQuery, f *persistence.SIPFilter) (page, whole *db.SIPQuery) {
	qf := entfilter.NewFilter(q, entfilter.SortableFields{
		sip.FieldID: {Name: "ID", Default: true},
	})
	qf.Contains(sip.FieldName, f.Name)
	qf.Equals(sip.FieldAipID, f.AIPID)
	qf.Equals(sip.FieldStatus, f.Status)
	qf.AddDateRange(sip.FieldCreatedAt, f.CreatedAt)
	qf.OrderBy(f.Sort)
	qf.Page(f.Limit, f.Offset)

	// Update the SIPFilter values with the actual values set on the query.
	// E.g. calling `qf.Page(0,0)` will set the query limit equal to the default
	// page size.
	f.Limit = qf.Limit
	f.Offset = qf.Offset

	return qf.Apply()
}
