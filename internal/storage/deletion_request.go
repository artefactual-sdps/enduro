package storage

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
)

type deletionActor struct {
	email         string
	sub           string
	iss           string
	authenticated bool
}

// checkClaims validates user identity for deletion operations. When API auth is
// disabled it returns an "unknown" actor.
func (s *serviceImpl) checkClaims(ctx context.Context) (deletionActor, error) {
	claims := auth.UserClaimsFromContext(ctx)
	if claims == nil {
		return deletionActor{
			email:         "unknown",
			sub:           "unknown",
			iss:           "unknown",
			authenticated: false,
		}, nil
	}
	if claims.Email == "" {
		return deletionActor{}, goastorage.MakeNotValid(errors.New("email claim is required"))
	}
	if claims.Sub == "" {
		return deletionActor{}, goastorage.MakeNotValid(errors.New("sub claim is required"))
	}
	if claims.Iss == "" {
		return deletionActor{}, goastorage.MakeNotValid(errors.New("iss claim is required"))
	}
	return deletionActor{
		email:         claims.Email,
		sub:           claims.Sub,
		iss:           claims.Iss,
		authenticated: true,
	}, nil
}

func (s *serviceImpl) RequestAipDeletion(ctx context.Context, payload *goastorage.RequestAipDeletionPayload) error {
	actor, err := s.checkClaims(ctx)
	if err != nil {
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
		UserEmail: actor.email,
		UserSub:   actor.sub,
		UserIss:   actor.iss,
	})
	if err != nil {
		s.logger.Error(err, "error initializing delete workflow")
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}

func (s *serviceImpl) ReviewAipDeletion(ctx context.Context, payload *goastorage.ReviewAipDeletionPayload) error {
	actor, err := s.checkClaims(ctx)
	if err != nil {
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

	// Ensure there is a pending deletion request for the AIP.
	drs, err := s.ListDeletionRequests(ctx, &persistence.DeletionRequestFilter{
		AIPUUID: &aipID,
		Status:  ref.New(enums.DeletionRequestStatusPending),
	})
	if err != nil || len(drs) == 0 {
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	if actor.authenticated && drs[0].RequesterIss == actor.iss && drs[0].RequesterSub == actor.sub {
		return goastorage.MakeNotValid(errors.New("requester cannot review their own request"))
	}

	status := enums.DeletionRequestStatusRejected
	if payload.Approved {
		status = enums.DeletionRequestStatusApproved
	}

	signal := DeletionDecisionSignal{
		Status:    status,
		UserEmail: actor.email,
		UserSub:   actor.sub,
		UserIss:   actor.iss,
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
	actor, err := s.checkClaims(ctx)
	if err != nil {
		return err
	}

	aipID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("invalid UUID"))
	}

	drs, err := s.ListDeletionRequests(ctx, &persistence.DeletionRequestFilter{
		AIPUUID: &aipID,
		Status:  ref.New(enums.DeletionRequestStatusPending),
	})
	if err != nil {
		return err
	}
	if len(drs) == 0 {
		return goastorage.MakeNotValid(errors.New("no valid deletion requests"))
	}

	// Check that the user is authorized to cancel the deletion request.
	if actor.authenticated && (actor.iss != drs[0].RequesterIss || actor.sub != drs[0].RequesterSub) {
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
			UserEmail: actor.email,
			UserSub:   actor.sub,
			UserIss:   actor.iss,
		},
	)
	if err != nil {
		return goastorage.MakeNotAvailable(errors.New("cannot perform operation"))
	}

	return nil
}
