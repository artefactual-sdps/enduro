package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/entfilter"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/aip"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/location"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
	"github.com/artefactual-sdps/enduro/internal/timerange"
)

var ErrUnexpectedUpdateResults = errors.New("update operation had unexpected results")

type Client struct {
	c *db.Client
}

var _ persistence.Storage = (*Client)(nil)

func NewClient(c *db.Client) *Client {
	return &Client{c: c}
}

func (c *Client) CreateAIP(ctx context.Context, goaaip *goastorage.AIP) (*goastorage.AIP, error) {
	status, err := enums.ParseAIPStatusWithDefault(goaaip.Status)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("status: invalid value"))
	}

	q := c.c.AIP.Create().
		SetName(goaaip.Name).
		SetAipID(goaaip.UUID).
		SetObjectKey(goaaip.ObjectKey).
		SetStatus(status)

	if goaaip.LocationID != nil {
		id, err := c.c.Location.Query().
			Where(location.UUID(*goaaip.LocationID)).
			OnlyID(ctx)
		if err != nil {
			if db.IsNotFound(err) {
				return nil, &goastorage.LocationNotFound{
					UUID: *goaaip.LocationID, Message: "location not found",
				}
			} else {
				return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
			}
		}
		q.SetLocationID(id)
	}

	a, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	return aipAsGoa(ctx, a), nil
}

func (c *Client) ListAIPs(ctx context.Context, payload *goastorage.ListAipsPayload) (*goastorage.AIPs, error) {
	if payload == nil {
		payload = &goastorage.ListAipsPayload{}
	}

	createdAt, err := timerange.Parse(payload.EarliestCreatedTime, payload.LatestCreatedTime)
	if err != nil {
		return nil, goastorage.MakeNotValid(fmt.Errorf("created at: %v", err))
	}

	var status *enums.AIPStatus
	if payload.Status != nil {
		s, err := enums.ParseAIPStatus(*payload.Status)
		if err != nil {
			return nil, goastorage.MakeNotValid(errors.New("status: invalid value"))
		}
		status = &s
	}

	qf := entfilter.NewFilter(c.c.AIP.Query(), entfilter.SortableFields{
		aip.FieldID: {Name: "ID", Default: true},
	})
	qf.Contains(aip.FieldName, payload.Name)
	qf.Equals(aip.FieldStatus, status)
	qf.AddDateRange(aip.FieldCreatedAt, createdAt)
	qf.OrderBy(entfilter.NewSort().AddCol("id", true))
	qf.Page(ref.DerefZero(payload.Limit), ref.DerefZero(payload.Offset))
	page, whole := qf.Apply()

	res, err := page.All(ctx)
	if err != nil {
		return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}
	total, err := whole.Count(ctx)
	if err != nil {
		return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	aips := []*goastorage.AIP{}
	for _, item := range res {
		aips = append(aips, aipAsGoa(ctx, item))
	}

	r := &goastorage.AIPs{
		Items: aips,
		Page: &goastorage.EnduroPage{
			Limit:  qf.Limit,
			Offset: qf.Offset,
			Total:  total,
		},
	}

	return r, err
}

func (c *Client) ReadAIP(ctx context.Context, aipID uuid.UUID) (*goastorage.AIP, error) {
	a, err := c.c.AIP.Query().
		Where(
			aip.AipID(aipID),
		).
		Only(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, &goastorage.AIPNotFound{UUID: aipID, Message: "AIP not found"}
		} else {
			return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
		}
	}

	return aipAsGoa(ctx, a), nil
}

func (c *Client) UpdateAIPStatus(ctx context.Context, aipID uuid.UUID, status enums.AIPStatus) error {
	n, err := c.c.AIP.Update().
		Where(
			aip.AipID(aipID),
		).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		return err
	}

	if n != 1 {
		return ErrUnexpectedUpdateResults
	}

	return nil
}

func (c *Client) UpdateAIPLocationID(ctx context.Context, aipID, locationID uuid.UUID) error {
	l, err := c.c.Location.Query().
		Where(
			location.UUID(locationID),
		).
		Only(ctx)
	if err != nil {
		return err
	}

	n, err := c.c.AIP.Update().
		Where(
			aip.AipID(aipID),
		).
		SetLocation(l).
		Save(ctx)
	if err != nil {
		return err
	}

	if n != 1 {
		return ErrUnexpectedUpdateResults
	}

	return nil
}

func aipAsGoa(ctx context.Context, a *db.AIP) *goastorage.AIP {
	p := &goastorage.AIP{
		Name:      a.Name,
		UUID:      a.AipID,
		Status:    a.Status.String(),
		ObjectKey: a.ObjectKey,
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
	}

	// TODO: should we use UUID as the foreign key?
	l, err := a.QueryLocation().Only(ctx)
	if err == nil {
		p.LocationID = &l.UUID
	}

	return p
}

func (c *Client) CreateLocation(
	ctx context.Context,
	location *goastorage.Location,
	config *types.LocationConfig,
) (*goastorage.Location, error) {
	purpose, err := enums.ParseLocationPurposeWithDefault(location.Purpose)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("purpose: invalid value"))
	}
	source, err := enums.ParseLocationSourceWithDefault(location.Source)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("source: invalid value"))
	}

	q := c.c.Location.Create()

	q.SetName(location.Name)
	q.SetDescription(ref.DerefZero(location.Description))
	q.SetSource(source)
	q.SetPurpose(purpose)
	q.SetUUID(location.UUID)

	q.SetConfig(ref.DerefZero(config))

	l, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	return locationAsGoa(l), nil
}

func (c *Client) ListLocations(ctx context.Context) (goastorage.LocationCollection, error) {
	locations := []*goastorage.Location{}

	res, err := c.c.Location.Query().All(ctx)
	for _, item := range res {
		locations = append(locations, locationAsGoa(item))
	}

	return locations, err
}

func (c *Client) ReadLocation(ctx context.Context, locationID uuid.UUID) (*goastorage.Location, error) {
	l, err := c.c.Location.Query().
		Where(
			location.UUID(locationID),
		).
		Only(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, &goastorage.LocationNotFound{UUID: locationID, Message: "location not found"}
		} else {
			return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
		}
	}

	return locationAsGoa(l), nil
}

func locationAsGoa(loc *db.Location) *goastorage.Location {
	l := &goastorage.Location{
		Name:        loc.Name,
		Description: &loc.Description,
		Source:      loc.Source.String(),
		Purpose:     loc.Purpose.String(),
		UUID:        loc.UUID,
		CreatedAt:   loc.CreatedAt.Format(time.RFC3339),
	}

	switch c := loc.Config.Value.(type) {
	case *types.S3Config:
		l.Config = &goastorage.S3Config{
			Bucket:    c.Bucket,
			Region:    c.Region,
			Endpoint:  &c.Endpoint,
			PathStyle: &c.PathStyle,
			Profile:   &c.Profile,
			Key:       &c.Key,
			Secret:    &c.Secret,
			Token:     &c.Token,
		}
	case *types.SFTPConfig:
		l.Config = &goastorage.SFTPConfig{
			Address:   c.Address,
			Username:  c.Username,
			Password:  c.Password,
			Directory: c.Directory,
		}
	case *types.AMSSConfig:
		l.Config = &goastorage.AMSSConfig{
			APIKey:   c.APIKey,
			URL:      c.URL,
			Username: c.Username,
		}
	case *types.URLConfig:
		l.Config = &goastorage.URLConfig{
			URL: c.URL,
		}

	}

	return l
}

func (c *Client) LocationAIPs(ctx context.Context, locationID uuid.UUID) (goastorage.AIPCollection, error) {
	res, err := c.c.Location.Query().Where(location.UUID(locationID)).QueryAips().All(ctx)
	if err != nil {
		return nil, err
	}

	aips := []*goastorage.AIP{}
	for _, item := range res {
		aips = append(aips, aipAsGoa(ctx, item))
	}

	return aips, nil
}
