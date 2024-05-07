package client

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/location"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/pkg"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

var ErrUnexpectedUpdateResults = errors.New("update operation had unexpected results")

type Client struct {
	c *db.Client
}

var _ persistence.Storage = (*Client)(nil)

func NewClient(c *db.Client) *Client {
	return &Client{c: c}
}

func (c *Client) CreatePackage(ctx context.Context, goapkg *goastorage.Package) (*goastorage.Package, error) {
	q := c.c.Pkg.Create()

	q.SetName(goapkg.Name)
	q.SetAipID(goapkg.AipID)
	q.SetObjectKey(goapkg.ObjectKey)
	q.SetStatus(types.NewPackageStatus(goapkg.Status))

	if goapkg.LocationID != nil {
		id, err := c.c.Location.Query().
			Where(location.UUID(*goapkg.LocationID)).
			OnlyID(ctx)
		if err != nil {
			if db.IsNotFound(err) {
				return nil, &goastorage.LocationNotFound{
					UUID: *goapkg.LocationID, Message: "location not found",
				}
			} else {
				return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
			}
		}
		q.SetLocationID(id)
	}

	pkg, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	return pkgAsGoa(ctx, pkg), nil
}

func (c *Client) ListPackages(ctx context.Context) (goastorage.PackageCollection, error) {
	pkgs := []*goastorage.Package{}

	res, err := c.c.Pkg.Query().All(ctx)
	for _, item := range res {
		pkgs = append(pkgs, pkgAsGoa(ctx, item))
	}

	return pkgs, err
}

func (c *Client) ReadPackage(ctx context.Context, aipID uuid.UUID) (*goastorage.Package, error) {
	pkg, err := c.c.Pkg.Query().
		Where(
			pkg.AipID(aipID),
		).
		Only(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, &goastorage.PackageNotFound{AipID: aipID, Message: "package not found"}
		} else {
			return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
		}
	}

	return pkgAsGoa(ctx, pkg), nil
}

func (c *Client) UpdatePackageStatus(ctx context.Context, aipID uuid.UUID, status types.PackageStatus) error {
	n, err := c.c.Pkg.Update().
		Where(
			pkg.AipID(aipID),
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

func (c *Client) UpdatePackageLocationID(ctx context.Context, aipID, locationID uuid.UUID) error {
	l, err := c.c.Location.Query().
		Where(
			location.UUID(locationID),
		).
		Only(ctx)
	if err != nil {
		return err
	}

	n, err := c.c.Pkg.Update().
		Where(
			pkg.AipID(aipID),
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

func pkgAsGoa(ctx context.Context, pkg *db.Pkg) *goastorage.Package {
	p := &goastorage.Package{
		Name:      pkg.Name,
		AipID:     pkg.AipID,
		Status:    pkg.Status.String(),
		ObjectKey: pkg.ObjectKey,
		CreatedAt: pkg.CreatedAt.Format(time.RFC3339),
	}

	// TODO: should we use UUID as the foreign key?
	l, err := pkg.QueryLocation().Only(ctx)
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
	q := c.c.Location.Create()

	q.SetName(location.Name)
	q.SetDescription(ref.DerefZero(location.Description))
	q.SetSource(types.NewLocationSource(location.Source))
	q.SetPurpose(types.NewLocationPurpose(location.Purpose))
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
	case *types.URLConfig:
		l.Config = &goastorage.URLConfig{
			URL: c.URL,
		}
	}

	return l
}

func (c *Client) LocationPackages(ctx context.Context, locationID uuid.UUID) (goastorage.PackageCollection, error) {
	res, err := c.c.Location.Query().Where(location.UUID(locationID)).QueryPackages().All(ctx)
	if err != nil {
		return nil, err
	}

	packages := []*goastorage.Package{}
	for _, item := range res {
		packages = append(packages, pkgAsGoa(ctx, item))
	}

	return packages, nil
}
