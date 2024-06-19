package upload

import (
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"time"

	"github.com/go-logr/logr"
	"go.artefactual.dev/tools/bucket"
	"goa.design/goa/v3/security"
	"gocloud.dev/blob"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goaupload "github.com/artefactual-sdps/enduro/internal/api/gen/upload"
)

const UPLOAD_MAX_SIZE = 102400000 // 100 MB

type Service interface {
	Upload(ctx context.Context, payload *goaupload.UploadPayload, req io.ReadCloser) error

	Bucket() *blob.Bucket
	Close() error
}

type serviceImpl struct {
	logger        logr.Logger
	config        Config
	bucket        *blob.Bucket
	uploadMaxSize int
	tokenVerifier auth.TokenVerifier
}

var _ Service = (*serviceImpl)(nil)

var (
	ErrUnauthorized error = goaupload.Unauthorized("Unauthorized")
	ErrForbidden    error = goaupload.Forbidden("Forbidden")
)

func NewService(
	logger logr.Logger,
	config Config,
	uploadMaxSize int,
	tokenVerifier auth.TokenVerifier,
) (s *serviceImpl, err error) {
	s = &serviceImpl{
		logger:        logger,
		config:        config,
		uploadMaxSize: uploadMaxSize,
		tokenVerifier: tokenVerifier,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	err = s.openBucket(ctx, config)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *serviceImpl) openBucket(ctx context.Context, config Config) error {
	if b, err := bucket.NewWithConfig(ctx, &bucket.Config{
		URL:       config.URL,
		Endpoint:  config.Endpoint,
		Bucket:    config.Bucket,
		AccessKey: config.Key,
		SecretKey: config.Secret,
		Token:     config.Token,
		Profile:   config.Profile,
		Region:    config.Region,
		PathStyle: config.PathStyle,
	}); err != nil {
		return err
	} else {
		s.bucket = b
	}

	return nil
}

func (s *serviceImpl) JWTAuth(
	ctx context.Context,
	token string,
	scheme *security.JWTScheme,
) (context.Context, error) {
	claims, err := s.tokenVerifier.Verify(ctx, token)
	if err != nil {
		if !errors.Is(err, auth.ErrUnauthorized) {
			s.logger.V(1).Info("failed to verify token", "err", err)
		}
		return ctx, ErrUnauthorized
	}

	if !claims.CheckAttributes(scheme.RequiredScopes) {
		return ctx, ErrForbidden
	}

	ctx = auth.WithUserClaims(ctx, claims)

	return ctx, nil
}

func (s *serviceImpl) Bucket() *blob.Bucket {
	return s.bucket
}

func (s *serviceImpl) Close() error {
	return s.bucket.Close()
}

func (s *serviceImpl) Upload(ctx context.Context, payload *goaupload.UploadPayload, req io.ReadCloser) error {
	defer req.Close()

	lr := io.LimitReader(req, int64(s.uploadMaxSize))

	_, params, err := mime.ParseMediaType(payload.ContentType)
	if err != nil {
		return goaupload.MakeInvalidMediaType(errors.New("invalid media type"))
	}
	mr := multipart.NewReader(lr, params["boundary"])

	part, err := mr.NextPart()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return goaupload.MakeInvalidMultipartRequest(errors.New("invalid multipart request"))
	}

	w, err := s.bucket.NewWriter(ctx, part.FileName(), &blob.WriterOptions{})
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(w, part)
	closeErr := w.Close()

	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}

	return nil
}
