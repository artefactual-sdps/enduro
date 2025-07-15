package ingest

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"gocloud.dev/gcerrors"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/persistence"
)

func (w *goaWrapper) readSIP(ctx context.Context, id string) (*datatypes.SIP, error) {
	// Validate the payload UUID.
	sipUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, goaingest.MakeNotValid(errors.New("invalid UUID"))
	}

	// Read the persisted SIP.
	sip, err := w.perSvc.ReadSIP(ctx, sipUUID)
	if err != nil {
		if errors.Is(err, persistence.ErrNotFound) {
			return nil, &goaingest.SIPNotFound{UUID: id, Message: "SIP not found"}
		} else {
			return nil, goaingest.MakeInternalError(errors.New("error reading SIP"))
		}
	}

	return sip, nil
}

func (w *goaWrapper) DownloadSipRequest(
	ctx context.Context,
	payload *goaingest.DownloadSipRequestPayload,
) (*goaingest.DownloadSipRequestResult, error) {
	sip, err := w.readSIP(ctx, payload.UUID)
	if err != nil {
		return nil, err
	}

	// Check that failed as and failed key values are set.
	if sip.FailedAs == "" || sip.FailedKey == "" {
		return nil, goaingest.MakeNotValid(errors.New("SIP has no failed values"))
	}

	// Check if the failed SIP/PIP exists in the internal bucket.
	exists, err := w.internalStorage.Exists(ctx, sip.FailedKey)
	if err != nil {
		return nil, goaingest.MakeInternalError(errors.New("error checking SIP/PIP file"))
	}

	if !exists {
		return nil, &goaingest.SIPNotFound{
			UUID:    payload.UUID,
			Message: "Failed SIP/PIP file not found in the internal storage",
		}
	}

	// Request a ticket.
	ticket, err := w.ticketProvider.Request(ctx, nil)
	if err != nil {
		return nil, goaingest.MakeInternalError(errors.New("ticket request failed"))
	}

	// A ticket is not provided when authentication is disabled.
	// Do not set the ticket cookie in that case.
	res := &goaingest.DownloadSipRequestResult{}
	if ticket != "" {
		res.Ticket = &ticket
	}

	return res, nil
}

func (w *goaWrapper) DownloadSip(
	ctx context.Context,
	payload *goaingest.DownloadSipPayload,
) (*goaingest.DownloadSipResult, io.ReadCloser, error) {
	// Verify the ticket.
	if err := w.ticketProvider.Check(ctx, payload.Ticket, nil); err != nil {
		return nil, nil, ErrUnauthorized
	}

	sip, err := w.readSIP(ctx, payload.UUID)
	if err != nil {
		return nil, nil, err
	}

	// Check that failed as and failed key values are set.
	if sip.FailedAs == "" || sip.FailedKey == "" {
		return nil, nil, goaingest.MakeNotValid(errors.New("SIP has no failed values"))
	}

	// Get a reader from the internal storage for the SIP failed key.
	reader, err := w.internalStorage.NewReader(ctx, sip.FailedKey, nil)
	if err != nil {
		if gcerrors.Code(err) == gcerrors.NotFound {
			return nil, nil, &goaingest.SIPNotFound{
				UUID:    payload.UUID,
				Message: "Failed SIP/PIP file not found in the internal storage",
			}
		} else {
			return nil, nil, goaingest.MakeInternalError(errors.New("error reading SIP file"))
		}
	}

	return &goaingest.DownloadSipResult{
		ContentType:        reader.ContentType(),
		ContentLength:      reader.Size(),
		ContentDisposition: fmt.Sprintf("attachment; filename=\"%s\"", sip.FailedKey),
	}, reader, nil
}
