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
	if claims.Iss == "" {
		return goastorage.MakeNotValid(errors.New("iss claim is required"))
	}

	return nil
}

func (s *serviceImpl) AipDeletionAuto(ctx context.Context, payload *goastorage.AipDeletionAutoPayload) error {
	if payload == nil {
		payload = &goastorage.AipDeletionAutoPayload{}
	}

	// Authentication can be disabled for auto-approve.
	claims := auth.UserClaimsFromContext(ctx)
	if claims != nil {
		if err := checkClaims(claims); err != nil {
			return err
		}
	} else {
		claims = &auth.Claims{
			Email: "unauthenticated@enduro.invalid",
			Sub:   "unauthenticated",
			Iss:   "unauthenticated",
		}
	}

	skipReport := payload.SkipReport != nil && *payload.SkipReport
	return s.requestAIPDeletion(ctx, payload.UUID, payload.Reason, claims, true, skipReport)
}

func (s *serviceImpl) RequestAipDeletion(ctx context.Context, payload *goastorage.RequestAipDeletionPayload) error {
	if payload == nil {
		payload = &goastorage.RequestAipDeletionPayload{}
	}

	// Authentication must be enabled for now.
	claims := auth.UserClaimsFromContext(ctx)
	if err := checkClaims(claims); err != nil {
		return err
	}

	return s.requestAIPDeletion(ctx, payload.UUID, payload.Reason, claims, false, false)
}

func (s *serviceImpl) requestAIPDeletion(
	ctx context.Context,
	aipUUID string,
	reason string,
	claims *auth.Claims,
	autoApprove bool,
	skipReport bool,
) error {
	aipID, err := uuid.Parse(aipUUID)
	if err != nil {
		return goastorage.MakeNotValid(errors.New("invalid UUID"))
	}
	if reason == "" {
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
		AIPID:       aipID,
		Reason:      reason,
		TaskQueue:   s.config.TaskQueue,
		UserEmail:   claims.Email,
		UserSub:     claims.Sub,
		UserIss:     claims.Iss,
		AutoApprove: autoApprove,
		SkipReport:  skipReport,
	})
	if err != nil {
		s.logger.Error(err, "error initializing delete workflow")
		return ErrInternalError
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

	// Ensure there is a pending deletion request for the AIP.
	drs, err := s.ListDeletionRequests(ctx, &persistence.DeletionRequestFilter{
		AIPUUID: &aipID,
		Status:  ref.New(enums.DeletionRequestStatusPending),
	})
	if err != nil || len(drs) == 0 {
		s.logger.Error(err, "deletion request not found", "aip_id", aipID)
		return ErrInternalError
	}

	if drs[0].RequesterIss == claims.Iss && drs[0].RequesterSub == claims.Sub {
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
		UserIss:   claims.Iss,
	}
	err = s.tc.SignalWorkflow(ctx, StorageDeleteWorkflowID(aipID), "", DeletionDecisionSignalName, signal)
	if err != nil {
		s.logger.Error(err, "error signaling delete workflow")
		return ErrInternalError
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
	if claims.Iss != drs[0].RequesterIss || claims.Sub != drs[0].RequesterSub {
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
			UserIss:   claims.Iss,
		},
	)
	if err != nil {
		s.logger.Error(err, "error signaling delete workflow")
		return ErrInternalError
	}

	return nil
}
