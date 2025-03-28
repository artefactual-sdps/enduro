package storage

import (
	"context"
	"errors"

	"github.com/google/uuid"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

func (s *serviceImpl) RequestAipDeletion(ctx context.Context, payload *goastorage.RequestAipDeletionPayload) error {
	aipID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("invalid UUID"))
	}
	if payload.Reason == "" {
		return goastorage.MakeNotValid(errors.New("invalid reason"))
	}

	// TODO:
	// - Check AIP existence and status, same as in workflow.
	// - Get user details from context claim and include them in the request.

	_, err = InitStorageDeleteWorkflow(ctx, s.tc, &StorageDeleteWorkflowRequest{
		AIPID:     aipID,
		Reason:    payload.Reason,
		TaskQueue: s.config.TaskQueue,
		UserEmail: "",
		UserSub:   "",
		UserISS:   "",
	})
	if err != nil {
		s.logger.Error(err, "error initializing delete workflow")
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

func (s *serviceImpl) ReviewAipDeletion(ctx context.Context, payload *goastorage.ReviewAipDeletionPayload) error {
	aipID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("invalid UUID"))
	}

	// TODO:
	// - Check AIP existence and status, and DeletionRequest.
	// - Get user details from context claim and include them in the signal.

	signal := DeletionReviewedSignal{
		Approved:  payload.Approved,
		UserEmail: "",
		UserSub:   "",
		UserISS:   "",
	}
	err = s.tc.SignalWorkflow(ctx, StorageDeleteWorkflowID(aipID), "", DeletionReviewedSignalName, signal)
	if err != nil {
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}
