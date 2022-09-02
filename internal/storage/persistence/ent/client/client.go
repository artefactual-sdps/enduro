package client

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/ref"
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

func (c *Client) CreatePackage(ctx context.Context, goapkg *goastorage.StoragePackage) (*goastorage.StoredStoragePackage, error) {
	q := c.c.Pkg.Create()

	q.SetName(goapkg.Name)

	q.SetAipID(goapkg.AipID)

	var objectKey uuid.UUID
	if goapkg.ObjectKey != nil {
		objectKey = *goapkg.ObjectKey
	}
	q.SetObjectKey(objectKey)

	q.SetStatus(types.NewPackageStatus(goapkg.Status))

	pkg, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	return pkgAsGoa(ctx, pkg), nil
}

func (c *Client) ListPackages(ctx context.Context) ([]*goastorage.StoredStoragePackage, error) {
	pkgs := []*goastorage.StoredStoragePackage{}

	res, err := c.c.Pkg.Query().All(ctx)
	for _, item := range res {
		pkgs = append(pkgs, pkgAsGoa(ctx, item))
	}

	return pkgs, err
}

func (c *Client) ReadPackage(ctx context.Context, AIPID uuid.UUID) (*goastorage.StoredStoragePackage, error) {
	pkg, err := c.c.Pkg.Query().
		Where(
			pkg.AipID(AIPID),
		).
		Only(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, &goastorage.StoragePackageNotfound{AipID: AIPID, Message: "package not found"}
		} else if err != nil {
			return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
		}
	}

	return pkgAsGoa(ctx, pkg), nil
}

func (c *Client) UpdatePackageStatus(ctx context.Context, status types.PackageStatus, AIPID uuid.UUID) error {
	n, err := c.c.Pkg.Update().
		Where(
			pkg.AipID(AIPID),
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

func (c *Client) UpdatePackageLocationID(ctx context.Context, locationID uuid.UUID, aipID uuid.UUID) error {
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

func pkgAsGoa(ctx context.Context, pkg *db.Pkg) *goastorage.StoredStoragePackage {
	p := &goastorage.StoredStoragePackage{
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

func (c *Client) CreateLocation(ctx context.Context, location *goastorage.Location, config *types.LocationConfig) (*goastorage.StoredLocation, error) {
	q := c.c.Location.Create()

	q.SetName(location.Name)
	q.SetDescription(ref.DerefZero(location.Description))
	q.SetSource(types.NewLocationSource(location.Source))
	q.SetPurpose(types.NewLocationPurpose(location.Purpose))

	if location.UUID != nil {
		q.SetUUID(*location.UUID)
	}

	q.SetConfig(ref.DerefZero(config))

	l, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	return locationAsGoa(l), nil
}

func (c *Client) ListLocations(ctx context.Context) (goastorage.StoredLocationCollection, error) {
	locations := []*goastorage.StoredLocation{}

	res, err := c.c.Location.Query().All(ctx)
	for _, item := range res {
		locations = append(locations, locationAsGoa(item))
	}

	return locations, err
}

func (c *Client) ReadLocation(ctx context.Context, locationID uuid.UUID) (*goastorage.StoredLocation, error) {
	l, err := c.c.Location.Query().
		Where(
			location.UUID(locationID),
		).
		Only(ctx)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, &goastorage.StorageLocationNotfound{UUID: locationID, Message: "location not found"}
		} else if err != nil {
			return nil, goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
		}
	}

	return locationAsGoa(l), nil
}

func locationAsGoa(loc *db.Location) *goastorage.StoredLocation {
	l := &goastorage.StoredLocation{
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
	}

	return l
}

func (c *Client) LocationPackages(ctx context.Context, locationID uuid.UUID) (goastorage.StoredStoragePackageCollection, error) {
	res, err := c.c.Location.Query().Where(location.UUID(locationID)).QueryPackages().All(ctx)
	if err != nil {
		return nil, err
	}

	packages := []*goastorage.StoredStoragePackage{}
	for _, item := range res {
		packages = append(packages, pkgAsGoa(ctx, item))
	}

	return packages, nil
}
