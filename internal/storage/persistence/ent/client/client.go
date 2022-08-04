package client

import (
	"context"
	"errors"

	"github.com/google/uuid"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/pkg"
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

	return asGoa(pkg), nil
}

func (c *Client) ListPackages(ctx context.Context) ([]*goastorage.StoredStoragePackage, error) {
	pkgs := []*goastorage.StoredStoragePackage{}

	res, err := c.c.Pkg.Query().All(ctx)
	for _, item := range res {
		pkgs = append(pkgs, asGoa(item))
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

	return asGoa(pkg), nil
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

func (c *Client) UpdatePackageLocation(ctx context.Context, location string, aipID uuid.UUID) error {
	n, err := c.c.Pkg.Update().
		Where(
			pkg.AipID(aipID),
		).
		SetLocation(location).
		Save(ctx)
	if err != nil {
		return err
	}

	if n != 1 {
		return ErrUnexpectedUpdateResults
	}

	return nil
}

func asGoa(pkg *db.Pkg) *goastorage.StoredStoragePackage {
	p := &goastorage.StoredStoragePackage{
		ID:        uint(pkg.ID),
		Name:      pkg.Name,
		AipID:     pkg.AipID.String(),
		Status:    pkg.Status.String(),
		ObjectKey: pkg.ObjectKey.String(),
	}

	if pkg.Location != "" {
		p.Location = &pkg.Location
	}

	return p
}
