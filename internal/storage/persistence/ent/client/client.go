package client

import (
	"context"
	"errors"

	"github.com/google/uuid"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/ref"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/location"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/pkg"
	"github.com/artefactual-sdps/enduro/internal/storage/purpose"
	"github.com/artefactual-sdps/enduro/internal/storage/source"
	"github.com/artefactual-sdps/enduro/internal/storage/status"
)

var ErrUnexpectedUpdateResults = errors.New("update operation had unexpected results")

type Client struct {
	c *db.Client
}

var _ persistence.Storage = (*Client)(nil)

func NewClient(c *db.Client) *Client {
	return &Client{c: c}
}

func (c *Client) CreatePackage(ctx context.Context, name string, AIPID uuid.UUID, objectKey uuid.UUID) (*goastorage.StoredStoragePackage, error) {
	pkg, err := c.c.Pkg.Create().
		SetName(name).
		SetAipID(AIPID).
		SetObjectKey(objectKey).
		SetStatus(status.StatusUnspecified).
		Save(ctx)
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
		return nil, err
	}

	return pkgAsGoa(ctx, pkg), nil
}

func (c *Client) UpdatePackageStatus(ctx context.Context, status status.PackageStatus, AIPID uuid.UUID) error {
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

func (c *Client) UpdatePackageLocation(ctx context.Context, locationName string, aipID uuid.UUID) error {
	l, err := c.c.Location.Query().
		Where(
			// TODO: switch to look by UUID
			location.Name(locationName),
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
		ID:        uint(pkg.ID),
		Name:      pkg.Name,
		AipID:     pkg.AipID.String(),
		Status:    pkg.Status.String(),
		ObjectKey: pkg.ObjectKey.String(),
	}

	l, err := pkg.QueryLocation().Only(ctx)
	if err == nil {
		// TODO: switch to location UUID
		p.Location = &l.Name
	}

	return p
}

func (c *Client) CreateLocation(ctx context.Context, name string, description *string, source source.LocationSource, purpose purpose.LocationPurpose, UUID uuid.UUID) (*goastorage.StoredLocation, error) {
	var d string
	if description != nil {
		d = *description
	}
	l, err := c.c.Location.Create().
		SetName(name).
		SetDescription(d).
		SetSource(source).
		SetPurpose(purpose).
		SetUUID(UUID).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return locationAsGoa(ctx, l), nil
}

func (c *Client) ReadLocation(ctx context.Context, UUID uuid.UUID) (*goastorage.StoredLocation, error) {
	l, err := c.c.Location.Query().
		Where(
			location.UUID(UUID),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return locationAsGoa(ctx, l), nil
}

func locationAsGoa(ctx context.Context, loc *db.Location) *goastorage.StoredLocation {
	l := &goastorage.StoredLocation{
		ID:          uint(loc.ID),
		Name:        loc.Name,
		Description: &loc.Description,
		Source:      loc.Source.String(),
		Purpose:     loc.Purpose.String(),
		UUID:        ref.New(loc.UUID.String()),
	}

	return l
}
