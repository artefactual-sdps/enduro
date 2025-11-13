package ingest

import (
	"context"
	"errors"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
)

var ErrNotImplemented error = goaingest.MakeNotImplemented(errors.New("not implemented"))

func (w *goaWrapper) AddBatch(
	ctx context.Context,
	payload *goaingest.AddBatchPayload,
) (*goaingest.AddBatchResult, error) {
	return nil, ErrNotImplemented
}

func (w *goaWrapper) ListBatches(
	ctx context.Context,
	payload *goaingest.ListBatchesPayload,
) (*goaingest.Batches, error) {
	return nil, ErrNotImplemented
}

func (w *goaWrapper) ShowBatch(
	ctx context.Context,
	payload *goaingest.ShowBatchPayload,
) (*goaingest.Batch, error) {
	return nil, ErrNotImplemented
}
