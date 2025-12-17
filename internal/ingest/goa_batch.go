package ingest

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence"
)

var ErrNotImplemented error = goaingest.MakeNotImplemented(errors.New("not implemented"))

func (svc *ingestImpl) AddBatch(
	ctx context.Context,
	payload *goaingest.AddBatchPayload,
) (*goaingest.AddBatchResult, error) {
	if payload == nil {
		return nil, goaingest.MakeNotValid(errors.New("missing payload"))
	}

	sourceID, err := uuid.Parse(payload.SourceID)
	if err != nil {
		return nil, goaingest.MakeNotValid(errors.New("invalid SourceID"))
	}

	if len(payload.Keys) == 0 {
		return nil, goaingest.MakeNotValid(errors.New("empty Keys"))
	}

	claims, err := checkClaims(ctx)
	if err != nil {
		return nil, goaingest.MakeNotValid(err)
	}

	// TODO: Discuss identifier generation strategy.
	bUUID := uuid.Must(uuid.NewRandomFromReader(svc.rander))
	identifier := fmt.Sprintf("Batch-%s", bUUID.String())
	if payload.Identifier != nil && *payload.Identifier != "" {
		identifier = *payload.Identifier
	}
	b := &datatypes.Batch{
		UUID:       bUUID,
		Status:     enums.BatchStatusQueued,
		Identifier: identifier,
		SIPSCount:  len(payload.Keys),
	}

	if claims != nil {
		b.Uploader = &datatypes.User{
			UUID:    uuid.Must(uuid.NewRandomFromReader(svc.rander)),
			Email:   claims.Email,
			Name:    claims.Name,
			OIDCIss: claims.Iss,
			OIDCSub: claims.Sub,
		}
	}

	if err := svc.perSvc.CreateBatch(ctx, b); err != nil {
		svc.logger.Error(err, "AddBatch")
		return nil, ErrInternalError
	}

	req := BatchWorkflowRequest{
		Batch:           *b,
		SIPSourceID:     sourceID,
		Keys:            payload.Keys,
		RetentionPeriod: svc.sipSource.RetentionPeriod(),
	}
	if err := InitBatchWorkflow(ctx, svc.tc, svc.taskQueue, &req); err != nil {
		// Delete Batch from persistence.
		err = errors.Join(err, svc.perSvc.DeleteBatch(ctx, b.UUID))
		svc.logger.Error(err, "AddBatch")
		return nil, ErrInternalError
	}

	PublishEvent(ctx, svc.evsvc, batchToCreatedEvent(b))
	svc.auditLogger.Log(ctx, batchIngestAuditEvent(b))

	return &goaingest.AddBatchResult{UUID: bUUID.String()}, nil
}

func (svc *ingestImpl) ListBatches(
	ctx context.Context,
	payload *goaingest.ListBatchesPayload,
) (*goaingest.Batches, error) {
	return nil, ErrNotImplemented
}

func (svc *ingestImpl) ShowBatch(
	ctx context.Context,
	payload *goaingest.ShowBatchPayload,
) (*goaingest.Batch, error) {
	batchUUID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return nil, goaingest.MakeNotValid(errors.New("invalid UUID"))
	}

	b, err := svc.perSvc.ReadBatch(ctx, batchUUID)
	if err == persistence.ErrNotFound {
		return nil, &goaingest.BatchNotFound{UUID: payload.UUID, Message: "Batch not found"}
	} else if err != nil {
		svc.logger.Error(err, "ShowBatch")
		return nil, ErrInternalError
	}

	return b.Goa(), nil
}
