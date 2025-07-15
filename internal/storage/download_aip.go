package storage

import (
	"context"
	"errors"
	"fmt"
	"io"

	"gocloud.dev/gcerrors"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/fsutil"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
)

func (s *serviceImpl) DownloadAipRequest(
	ctx context.Context,
	payload *goastorage.DownloadAipRequestPayload,
) (*goastorage.DownloadAipRequestResult, error) {
	aip, err := s.ShowAip(ctx, &goastorage.ShowAipPayload{UUID: payload.UUID})
	if err != nil {
		return nil, err
	}

	if aip.Status != enums.AIPStatusStored.String() && aip.Status != enums.AIPStatusPending.String() {
		return nil, goastorage.MakeNotValid(errors.New("AIP is not available for download"))
	}

	// Check if the failed AIP exists in the location bucket.
	reader, err := s.AipReader(ctx, aip)
	if err != nil {
		if gcerrors.Code(err) == gcerrors.NotFound {
			return nil, &goastorage.AIPNotFound{
				UUID:    aip.UUID,
				Message: "AIP file not found in the location bucket",
			}
		} else {
			return nil, goastorage.MakeInternalError(errors.New("error checking AIP file"))
		}
	}
	reader.Close()

	// Request a ticket.
	ticket, err := s.ticketProvider.Request(ctx, nil)
	if err != nil {
		return nil, goastorage.MakeInternalError(errors.New("ticket request failed"))
	}

	// A ticket is not provided when authentication is disabled.
	// Do not set the ticket cookie in that case.
	res := &goastorage.DownloadAipRequestResult{}
	if ticket != "" {
		res.Ticket = &ticket
	}

	return res, nil
}

func (s *serviceImpl) DownloadAip(
	ctx context.Context,
	payload *goastorage.DownloadAipPayload,
) (*goastorage.DownloadAipResult, io.ReadCloser, error) {
	// Verify the ticket.
	if err := s.ticketProvider.Check(ctx, payload.Ticket, nil); err != nil {
		return nil, nil, ErrUnauthorized
	}

	aip, err := s.ShowAip(ctx, &goastorage.ShowAipPayload{UUID: payload.UUID})
	if err != nil {
		return nil, nil, err
	}

	if aip.Status != enums.AIPStatusStored.String() && aip.Status != enums.AIPStatusPending.String() {
		return nil, nil, goastorage.MakeNotValid(errors.New("AIP is not available for download"))
	}

	// Get a reader from the location bucket for the AIP object key.
	reader, err := s.AipReader(ctx, aip)
	if err != nil {
		if gcerrors.Code(err) == gcerrors.NotFound {
			return nil, nil, &goastorage.AIPNotFound{
				UUID:    aip.UUID,
				Message: "AIP file not found in the location bucket",
			}
		} else {
			return nil, nil, goastorage.MakeInternalError(errors.New("error reading AIP file"))
		}
	}

	filename := fmt.Sprintf("%s-%s.7z", fsutil.BaseNoExt(aip.Name), aip.UUID)

	return &goastorage.DownloadAipResult{
		ContentType:        reader.ContentType(),
		ContentLength:      reader.Size(),
		ContentDisposition: fmt.Sprintf("attachment; filename=\"%s\"", filename),
	}, reader, nil
}
