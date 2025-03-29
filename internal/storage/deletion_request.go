package storage

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

func (s *serviceImpl) RequestAipDeletion(ctx context.Context, payload *goastorage.RequestAipDeletionPayload) error {
	// Authentication must be enabled for now.
	claims := auth.UserClaimsFromContext(ctx)
	if claims == nil {
		return goastorage.MakeNotValid(errors.New("authentication is required"))
	}
	if claims.Email == "" {
		return goastorage.MakeNotValid(errors.New("email claim is required"))
	}
	if claims.Sub == "" {
		return goastorage.MakeNotValid(errors.New("sub claim is required"))
	}
	if claims.ISS == "" {
		return goastorage.MakeNotValid(errors.New("iss claim is required"))
	}

	aipID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("invalid UUID"))
	}
	if payload.Reason == "" {
		return goastorage.MakeNotValid(errors.New("invalid reason"))
	}

	// TODO: Check AIP existence and status, same as in workflow.

	_, err = InitStorageDeleteWorkflow(ctx, s.tc, &StorageDeleteWorkflowRequest{
		AIPID:     aipID,
		Reason:    payload.Reason,
		TaskQueue: s.config.TaskQueue,
		UserEmail: claims.Email,
		UserSub:   claims.Sub,
		UserISS:   claims.ISS,
	})
	if err != nil {
		s.logger.Error(err, "error initializing delete workflow")
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

func (s *serviceImpl) ReviewAipDeletion(ctx context.Context, payload *goastorage.ReviewAipDeletionPayload) error {
	// Authentication must be enabled for now.
	claims := auth.UserClaimsFromContext(ctx)
	if claims == nil {
		return goastorage.MakeNotValid(errors.New("authentication is required"))
	}
	if claims.Email == "" {
		return goastorage.MakeNotValid(errors.New("email claim is required"))
	}
	if claims.Sub == "" {
		return goastorage.MakeNotValid(errors.New("sub claim is required"))
	}
	if claims.ISS == "" {
		return goastorage.MakeNotValid(errors.New("iss claim is required"))
	}

	aipID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("invalid UUID"))
	}

	// TODO: Check AIP existence and status, and DeletionRequest.

	signal := DeletionReviewedSignal{
		Approved:  payload.Approved,
		UserEmail: claims.Email,
		UserSub:   claims.Sub,
		UserISS:   claims.ISS,
	}
	err = s.tc.SignalWorkflow(ctx, StorageDeleteWorkflowID(aipID), "", DeletionReviewedSignalName, signal)
	if err != nil {
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}
