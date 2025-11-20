package ingest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mholt/archives"
	"gocloud.dev/blob"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
)

type UploadConfig struct {
	MaxSize int64

	// RetentionPeriod is the duration for which SIPs should be retained after
	// a successful ingest. If negative, SIPs will be retained indefinitely.
	RetentionPeriod time.Duration
}

func (svc *ingestImpl) UploadSip(
	ctx context.Context,
	payload *goaingest.UploadSipPayload,
	req io.ReadCloser,
) (*goaingest.UploadSipResult, error) {
	defer req.Close()

	// Check claims before processing the upload.
	claims, err := checkClaims(ctx)
	if err != nil {
		return nil, goaingest.MakeNotValid(err)
	}

	lr := io.LimitReader(req, int64(svc.uploadMaxSize))

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

	// Identify file format to add extension in the object key.
	format, stream, err := archives.Identify(ctx, part.FileName(), part)
	if err != nil {
		return nil, goaingest.MakeInvalidMultipartRequest(errors.New("unable to identify format"))
	}

	// Remove the format extension from the filename if it is included.
	ext := format.Extension()
	name := strings.TrimSuffix(part.FileName(), ext)
	sipUUID := uuid.Must(uuid.NewRandomFromReader(svc.rander))
	objectKey := fmt.Sprintf("%s%s-%s%s", SIPPrefix, name, sipUUID.String(), ext)
	wr, err := svc.internalStorage.NewWriter(ctx, objectKey, &blob.WriterOptions{})
	if err != nil {
		return nil, err
	}

	_, copyErr := io.Copy(wr, stream)
	closeErr := wr.Close()

	if copyErr != nil {
		return nil, copyErr
	}
	if closeErr != nil {
		return nil, closeErr
	}

	if err := svc.initSIP(
		ctx,
		sipUUID,
		part.FileName(),
		objectKey,
		ext,
		enums.WorkflowTypeCreateAip,
		claims,
	); err != nil {
		// Delete SIP from internal bucket.
		err := errors.Join(err, svc.internalStorage.Delete(ctx, objectKey))
		svc.logger.Error(err, "failed to init SIP ingest workflow after upload")
		return nil, err
	}

	return &goaingest.UploadSipResult{UUID: sipUUID.String()}, nil
}

func (svc *ingestImpl) initSIP(
	ctx context.Context,
	id uuid.UUID,
	name string,
	key string,
	extension string,
	wType enums.WorkflowType,
	claims *auth.Claims,
) error {
	s := &datatypes.SIP{
		UUID:   id,
		Name:   name,
		Status: enums.SIPStatusQueued,
	}

	// If claims is nil, it means authentication is not enabled.
	if claims != nil {
		s.Uploader = &datatypes.User{
			UUID:    uuid.Must(uuid.NewRandomFromReader(svc.rander)),
			Email:   claims.Email,
			Name:    claims.Name,
			OIDCIss: claims.Iss,
			OIDCSub: claims.Sub,
		}
	}

	if err := svc.perSvc.CreateSIP(ctx, s); err != nil {
		return err
	}

	req := ProcessingWorkflowRequest{
		SIPUUID:         id,
		SIPName:         name,
		Type:            wType,
		Key:             key,
		Extension:       extension,
		RetentionPeriod: svc.uploadRetentionPeriod,
	}
	if err := InitProcessingWorkflow(ctx, svc.tc, svc.taskQueue, &req); err != nil {
		// Delete SIP from persistence.
		return errors.Join(err, svc.perSvc.DeleteSIP(ctx, id))
	}

	PublishEvent(ctx, svc.evsvc, sipToCreatedEvent(s))
	svc.auditLogger.Log(ctx, sipIngestAuditEvent(s))

	return nil
}

func checkClaims(ctx context.Context) (*auth.Claims, error) {
	claims := auth.UserClaimsFromContext(ctx)
	if claims == nil {
		return nil, nil
	}
	if claims.Iss == "" {
		return nil, fmt.Errorf("invalid user claims: missing Iss")
	}
	if claims.Sub == "" {
		return nil, fmt.Errorf("invalid user claims: missing Sub")
	}
	return claims, nil
}
