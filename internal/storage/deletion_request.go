package storage

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
)

func checkClaims(claims *auth.Claims) error {
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

	return nil
}

func (s *serviceImpl) RequestAipDeletion(ctx context.Context, payload *goastorage.RequestAipDeletionPayload) error {
	// Authentication must be enabled for now.
	claims := auth.UserClaimsFromContext(ctx)
	if err := checkClaims(claims); err != nil {
		return err
	}

	aipID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("invalid UUID"))
	}
	if payload.Reason == "" {
		return goastorage.MakeNotValid(errors.New("invalid reason"))
	}

	aip, err := s.ReadAip(ctx, aipID)
	if err != nil {
		return err
	}
	if aip.Status != enums.AIPStatusStored.String() {
		return goastorage.MakeNotValid(errors.New("AIP is not stored"))
	}

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
	if err := checkClaims(claims); err != nil {
		return err
	}

	aipID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("invalid UUID"))
	}

	aip, err := s.ReadAip(ctx, aipID)
	if err != nil {
		return err
	}
	if aip.Status != enums.AIPStatusPending.String() {
		return goastorage.MakeNotValid(errors.New("AIP is not awaiting user review"))
	}

	dr, err := s.ReadAipPendingDeletionRequest(ctx, aipID)
	if err != nil {
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	if dr.RequesterISS == claims.ISS && dr.RequesterSub == claims.Sub {
		return goastorage.MakeNotValid(errors.New("requester cannot review their own request"))
	}

	status := enums.DeletionRequestStatusRejected
	if payload.Approved {
		status = enums.DeletionRequestStatusApproved
	}

	signal := DeletionDecisionSignal{
		Status:    status,
		UserEmail: claims.Email,
		UserSub:   claims.Sub,
		UserISS:   claims.ISS,
	}
	err = s.tc.SignalWorkflow(ctx, StorageDeleteWorkflowID(aipID), "", DeletionDecisionSignalName, signal)
	if err != nil {
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

func (s *serviceImpl) CancelAipDeletion(
	ctx context.Context,
	payload *goastorage.CancelAipDeletionPayload,
) error {
	// Authentication must be enabled for now.
	claims := auth.UserClaimsFromContext(ctx)
	if err := checkClaims(claims); err != nil {
		return err
	}

	aipID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("invalid UUID"))
	}

	dr, err := s.storagePersistence.ReadAipPendingDeletionRequest(ctx, aipID)
	if err != nil {
		return err
	}

	// Check that the user is authorized to cancel the deletion request.
	if claims.ISS != dr.RequesterISS || claims.Sub != dr.RequesterSub {
		return ErrForbidden
	}

	// If the check flag is set, do not cancel the deletion request.
	if payload.Check != nil && *payload.Check {
		return nil
	}

	err = s.tc.SignalWorkflow(
		ctx,
		StorageDeleteWorkflowID(aipID),
		"",
		DeletionDecisionSignalName,
		DeletionDecisionSignal{
			Status:    enums.DeletionRequestStatusCanceled,
			UserEmail: claims.Email,
			UserSub:   claims.Sub,
			UserISS:   claims.ISS,
		},
	)
	if err != nil {
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}
