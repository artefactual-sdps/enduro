package storage

import (
	"context"
	"errors"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

func (s *serviceImpl) MonitorRequest(
	ctx context.Context,
	payload *goastorage.MonitorRequestPayload,
) (*goastorage.MonitorRequestResult, error) {
	return nil, goastorage.MakeNotImplemented(errors.New("not implemented"))
}

// Monitor storage activity. It implements goastorage.Service.
func (s *serviceImpl) Monitor(
	ctx context.Context,
	payload *goastorage.MonitorPayload,
	stream goastorage.MonitorServerStream,
) error {
	return goastorage.MakeNotImplemented(errors.New("not implemented"))
}
