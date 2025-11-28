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
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

func deletionReportReader(
	ctx context.Context,
	s *serviceImpl,
	dr *types.DeletionRequest,
) (*blob.Reader, error) {
	loc, err := s.Location(ctx, uuid.Nil)
	if err != nil {
		return nil, err
	}

	b, err := loc.OpenBucket(ctx)
	if err != nil {
		return nil, err
	}

	return b.NewReader(ctx, dr.ReportKey, nil)
}

func (s *serviceImpl) DownloadDeletionReportRequest(
	ctx context.Context,
	payload *goastorage.DownloadDeletionReportRequestPayload,
) (*goastorage.DownloadDeletionReportRequestResult, error) {
	id, err := uuid.Parse(payload.UUID)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("invalid UUID"))
	}

	dr, err := s.ReadDeletionRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	if dr.Status != enums.DeletionRequestStatusApproved || dr.ReportKey == "" {
		return nil, goastorage.MakeNotValid(errors.New("deletion report is not available for download"))
	}

	// Check that the deletion report exists in the location bucket.
	r, err := deletionReportReader(ctx, s, dr)
	if err != nil {
		if gcerrors.Code(err) == gcerrors.NotFound {
			return nil, goastorage.MakeNotFound(errors.New("deletion report not found"))
		} else {
			return nil, goastorage.MakeInternalError(errors.New("error reading deletion report"))
		}
	}
	r.Close()

	// Request a ticket.
	ticket, err := s.ticketProvider.Request(ctx, nil)
	if err != nil {
		return nil, goastorage.MakeInternalError(errors.New("ticket request failed"))
	}

	res := &goastorage.DownloadDeletionReportRequestResult{}
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
		Msg:        "Deletion report download requested",
		Type:       "deletion_report.download",
		ResourceID: dr.UUID.String(),
		User:       userEmail,
	})

	return res, nil
}

func (s *serviceImpl) DownloadDeletionReport(
	ctx context.Context,
	payload *goastorage.DownloadDeletionReportPayload,
) (*goastorage.DownloadDeletionReportResult, io.ReadCloser, error) {
	// Verify the ticket.
	if err := s.ticketProvider.Check(ctx, payload.Ticket, nil); err != nil {
		return nil, nil, ErrUnauthorized
	}

	id, err := uuid.Parse(payload.UUID)
	if err != nil {
		return nil, nil, goastorage.MakeNotValid(errors.New("invalid UUID"))
	}

	dr, err := s.ReadDeletionRequest(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	if dr.Status != enums.DeletionRequestStatusApproved || dr.ReportKey == "" {
		return nil, nil, goastorage.MakeNotValid(errors.New("deletion report is not available for download"))
	}

	r, err := deletionReportReader(ctx, s, dr)
	if err != nil {
		return nil, nil, goastorage.MakeInternalError(errors.New("error reading deletion report"))
	}

	filename, ok := strings.CutPrefix(ReportPrefix+"/", dr.ReportKey)
	if !ok {
		filename = fmt.Sprintf("aip_deletion_report_%s.pdf", id)
	}

	return &goastorage.DownloadDeletionReportResult{
		ContentType:        r.ContentType(),
		ContentLength:      r.Size(),
		ContentDisposition: fmt.Sprintf("attachment; filename=\"%s\"", filename),
	}, r, nil
}
