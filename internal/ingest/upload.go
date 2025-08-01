package ingest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"strings"

	"github.com/google/uuid"
	"github.com/mholt/archiver/v4"
	"gocloud.dev/blob"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/persistence"
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

	// Identify file format to add extension in the object key.
	// TODO: Use github.com/mholt/archives. We still use the archived github.com/mholt/archiver/v4
	// in some activities, and using both causes a panic registering the same compressors.
	format, stream, err := archiver.Identify(part.FileName(), part)
	if err != nil {
		return nil, goaingest.MakeInvalidMultipartRequest(errors.New("unable to identify format"))
	}

	// Remove the format extension from the filename if it is included.
	ext := format.Name()
	name := strings.TrimSuffix(part.FileName(), ext)
	sipUUID := uuid.Must(uuid.NewRandomFromReader(w.rander))
	objectKey := fmt.Sprintf("%s%s-%s%s", SIPPrefix, name, sipUUID.String(), ext)
	wr, err := w.internalStorage.NewWriter(ctx, objectKey, &blob.WriterOptions{})
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

	if err := w.initSIP(
		ctx,
		sipUUID,
		part.FileName(),
		objectKey,
		ext,
		enums.WorkflowTypeCreateAip,
	); err != nil {
		// Delete SIP from internal bucket.
		err := errors.Join(err, w.internalStorage.Delete(ctx, objectKey))
		w.logger.Error(err, "Error initializing SIP ingest workflow after upload.")
		return nil, err
	}

	return &goaingest.UploadSipResult{UUID: sipUUID.String()}, nil
}

func (w *goaWrapper) initSIP(
	ctx context.Context,
	id uuid.UUID,
	name string,
	key string,
	extension string,
	wType enums.WorkflowType,
) error {
	user, err := w.findOrCreateUser(ctx, w.perSvc)
	if err != nil {
		return err
	}

	s := &datatypes.SIP{
		UUID:   id,
		Name:   name,
		Status: enums.SIPStatusQueued,
	}

	// If user is nil, it means authentication is not enabled.
	if user != nil {
		s.Uploader = &datatypes.Uploader{
			UUID:  user.UUID,
			Email: user.Email,
			Name:  user.Name,
		}
	}

	if err := w.perSvc.CreateSIP(ctx, s); err != nil {
		return err
	}

	req := ProcessingWorkflowRequest{
		SIPUUID:   id,
		SIPName:   name,
		Type:      wType,
		Key:       key,
		Extension: extension,
	}
	if err := InitProcessingWorkflow(ctx, w.tc, w.taskQueue, &req); err != nil {
		// Delete SIP from persistence.
		return errors.Join(err, w.perSvc.DeleteSIP(ctx, s.ID))
	}

	event.PublishIngestEvent(ctx, w.evsvc, sipToCreatedEvent(s))

	return nil
}

func (w *goaWrapper) findOrCreateUser(
	ctx context.Context,
	perSvc persistence.Service,
) (*datatypes.User, error) {
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

	user, err := perSvc.ReadOIDCUser(ctx, claims.Iss, claims.Sub)
	if err != nil {
		if errors.Is(err, persistence.ErrNotFound) {
			// User does not exist, create a new one.
			user, err := w.createUser(ctx, perSvc, claims)
			if err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
			return user, nil
		}

		return nil, err
	}

	return user, nil
}

func (w *goaWrapper) createUser(
	ctx context.Context,
	perSvc persistence.Service,
	claims *auth.Claims,
) (*datatypes.User, error) {
	u := &datatypes.User{
		UUID:    uuid.Must(uuid.NewRandomFromReader(w.rander)),
		Email:   claims.Email,
		Name:    claims.Name,
		OIDCIss: claims.Iss,
		OIDCSub: claims.Sub,
	}

	if err := perSvc.CreateUser(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return u, nil
}
