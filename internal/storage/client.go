package storage

import (
	"context"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

type Client interface {
	Submit(context.Context, *goastorage.SubmitPayload) (*goastorage.SubmitResult, error)
	Create(context.Context, *goastorage.CreatePayload) (*goastorage.Package, error)
	Update(context.Context, *goastorage.UpdatePayload) error
	Download(context.Context, *goastorage.DownloadPayload) ([]byte, error)
	Locations(context.Context, *goastorage.LocationsPayload) (goastorage.LocationCollection, error)
	AddLocation(context.Context, *goastorage.AddLocationPayload) (*goastorage.AddLocationResult, error)
	Move(context.Context, *goastorage.MovePayload) error
	MoveStatus(context.Context, *goastorage.MoveStatusPayload) (*goastorage.MoveStatusResult, error)
	Reject(context.Context, *goastorage.RejectPayload) error
	Show(context.Context, *goastorage.ShowPayload) (*goastorage.Package, error)
	ShowLocation(context.Context, *goastorage.ShowLocationPayload) (*goastorage.Location, error)
	LocationPackages(context.Context, *goastorage.LocationPackagesPayload) (goastorage.PackageCollection, error)
}

var _ Client = (*goastorage.Client)(nil)
