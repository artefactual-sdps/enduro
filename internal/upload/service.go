package upload

import (
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-logr/logr"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"

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
}

var _ Service = (*serviceImpl)(nil)

func NewService(logger logr.Logger, config Config, uploadMaxSize int) (s *serviceImpl, err error) {
	s = &serviceImpl{
		logger:        logger,
		config:        config,
		uploadMaxSize: uploadMaxSize,
	}

	err = s.openBucket(config)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *serviceImpl) openBucket(config Config) error {
	var err error
	var b *blob.Bucket
	ctx := context.Background()

	if config.URL != "" {
		b, err = blob.OpenBucket(ctx, config.URL)
	} else {
		b, err = s.openS3Bucket(ctx)
	}
	if err != nil {
		return err
	}
	s.bucket = b

	return nil
}

func (s *serviceImpl) openS3Bucket(ctx context.Context) (*blob.Bucket, error) {
	sessOpts := session.Options{}
	sessOpts.Config.WithRegion(s.config.Region)
	sessOpts.Config.WithEndpoint(s.config.Endpoint)
	sessOpts.Config.WithS3ForcePathStyle(s.config.PathStyle)
	sessOpts.Config.WithCredentials(
		credentials.NewStaticCredentials(
			s.config.Key, s.config.Secret, s.config.Token,
		),
	)
	sess, err := session.NewSessionWithOptions(sessOpts)
	if err != nil {
		return nil, err
	}
	return s3blob.OpenBucket(ctx, sess, s.config.Bucket, nil)
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
