package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"gocloud.dev/blob"
	"gocloud.dev/gcerrors"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/auditlog"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
)

func deletionReportReader(
	ctx context.Context,
	s *serviceImpl,
	aipID string,
) (r *blob.Reader, key string, err error) {
	id, err := uuid.Parse(aipID)
	if err != nil {
		return nil, "", goastorage.MakeNotValid(errors.New("invalid UUID"))
	}

	aip, err := s.ReadAip(ctx, id)
	if err != nil {
		return nil, "", err
	}

	if aip.Status != enums.AIPStatusDeleted.String() || aip.DeletionReportKey == nil || *aip.DeletionReportKey == "" {
		return nil, "", goastorage.MakeNotValid(errors.New("deletion report is not available for download"))
	}

	loc, err := s.Location(ctx, uuid.Nil)
	if err != nil {
		return nil, "", err
	}

	b, err := loc.OpenBucket(ctx)
	if err != nil {
		return nil, "", err
	}

	r, err = b.NewReader(ctx, *aip.DeletionReportKey, nil)
	if err != nil {
		if gcerrors.Code(err) == gcerrors.NotFound {
			return nil, "", goastorage.MakeNotFound(errors.New("deletion report not found"))
		} else {
			return nil, "", goastorage.MakeInternalError(errors.New("error reading deletion report"))
		}
	}
	return r, *aip.DeletionReportKey, nil
}

func (s *serviceImpl) AipDeletionReportRequest(
	ctx context.Context,
	payload *goastorage.AipDeletionReportRequestPayload,
) (*goastorage.AipDeletionReportRequestResult, error) {
	// Check that the deletion report exists in the location bucket and can be
	// read.
	r, _, err := deletionReportReader(ctx, s, payload.UUID)
	if err != nil {
		return nil, err
	}
	r.Close()

	// Request a ticket.
	ticket, err := s.ticketProvider.Request(ctx, nil)
	if err != nil {
		return nil, goastorage.MakeInternalError(errors.New("ticket request failed"))
	}

	res := &goastorage.AipDeletionReportRequestResult{}
	var userEmail string

	// A ticket is not provided when authentication is disabled.
	// Do not set the ticket cookie in that case.
	if ticket != "" {
		res.Ticket = &ticket

		claims := auth.UserClaimsFromContext(ctx)
		if claims != nil && claims.Email != "" {
			userEmail = claims.Email
		}
	}

	s.auditLogger.Log(ctx, &auditlog.Event{
		Level:      auditlog.LevelInfo,
		Msg:        "AIP deletion report download requested",
		Type:       "aip.download_deletion_report",
		ResourceID: payload.UUID,
		User:       userEmail,
	})

	return res, nil
}

func (s *serviceImpl) AipDeletionReport(
	ctx context.Context,
	payload *goastorage.AipDeletionReportPayload,
) (*goastorage.AipDeletionReportResult, io.ReadCloser, error) {
	// Verify the ticket.
	if err := s.ticketProvider.Check(ctx, payload.Ticket, nil); err != nil {
		return nil, nil, ErrUnauthorized
	}

	r, key, err := deletionReportReader(ctx, s, payload.UUID)
	if err != nil {
		return nil, nil, err
	}

	filename, ok := strings.CutPrefix(ReportPrefix, key)
	if !ok {
		filename = fmt.Sprintf("aip_deletion_report_%s.pdf", payload.UUID)
	}

	return &goastorage.AipDeletionReportResult{
		ContentType:        r.ContentType(),
		ContentLength:      r.Size(),
		ContentDisposition: fmt.Sprintf("attachment; filename=\"%s\"", filename),
	}, r, nil
}
