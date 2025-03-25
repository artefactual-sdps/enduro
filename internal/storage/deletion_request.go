package storage

import (
	"context"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

func (svc *serviceImpl) RequestAipDeletion(ctx context.Context, payload *goastorage.RequestAipDeletionPayload) error {
	return nil
}

func (svc *serviceImpl) ReviewAipDeletion(ctx context.Context, payload *goastorage.ReviewAipDeletionPayload) error {
	return nil
}
