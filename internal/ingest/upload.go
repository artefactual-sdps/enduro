package ingest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"

	"github.com/google/uuid"
	"gocloud.dev/blob"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/event"
)

type UploadConfig struct {
	MaxSize int64
}

func (w *goaWrapper) UploadSip(
	ctx context.Context,
	payload *goaingest.UploadSipPayload,
	req io.ReadCloser,
) (*goaingest.UploadSipResult, error) {
	defer req.Close()

	lr := io.LimitReader(req, int64(w.uploadMaxSize))

	_, params, err := mime.ParseMediaType(payload.ContentType)
	if err != nil {
		return nil, goaingest.MakeInvalidMediaType(errors.New("invalid media type"))
	}
	mr := multipart.NewReader(lr, params["boundary"])

	part, err := mr.NextPart()
	if err == io.EOF {
		return nil, goaingest.MakeInvalidMultipartRequest(errors.New("missing file part in upload"))
	}
	if err != nil {
		return nil, goaingest.MakeInvalidMultipartRequest(errors.New("invalid multipart request"))
	}

	sipUUID := uuid.Must(uuid.NewRandomFromReader(w.rander))
	objectKey := fmt.Sprintf("%s%s", SIPPrefix, sipUUID.String())
	wr, err := w.internalBucket.NewWriter(ctx, objectKey, &blob.WriterOptions{})
	if err != nil {
		return nil, err
	}

	_, copyErr := io.Copy(wr, part)
	closeErr := wr.Close()

	if copyErr != nil {
		return nil, copyErr
	}
	if closeErr != nil {
		return nil, closeErr
	}

	if err := w.initSIP(ctx, sipUUID, part.FileName(), objectKey, enums.WorkflowTypeCreateAip); err != nil {
		// Delete SIP from internal bucket.
		err := errors.Join(err, w.internalBucket.Delete(ctx, objectKey))
		w.logger.Error(err, "Error initializing SIP ingest workflow after upload.")
		return nil, err
	}

	return &goaingest.UploadSipResult{UUID: sipUUID.String()}, nil
}

func (w *goaWrapper) initSIP(ctx context.Context, id uuid.UUID, name, key string, wType enums.WorkflowType) error {
	s := &datatypes.SIP{
		UUID:   id,
		Name:   name,
		Status: enums.SIPStatusQueued,
	}
	if err := w.perSvc.CreateSIP(ctx, s); err != nil {
		return err
	}

	req := ProcessingWorkflowRequest{
		SIPUUID: id,
		SIPName: name,
		Type:    wType,
		Key:     key,
	}
	if err := InitProcessingWorkflow(ctx, w.tc, w.taskQueue, &req); err != nil {
		// Delete SIP from persistence.
		return errors.Join(err, w.perSvc.DeleteSIP(ctx, s.ID))
	}

	event.PublishEvent(ctx, w.evsvc, sipToCreatedEvent(s))

	return nil
}
