package storage

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	temporalsdk_client "go.temporal.io/sdk/client"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"

	goastorage "github.com/artefactual-labs/enduro/internal/api/gen/storage"
)

var urlExpirationTime = 15 * time.Minute

type Service interface {
	Submit(context.Context, *goastorage.SubmitPayload) (res *goastorage.SubmitResult, err error)
	Update(context.Context, *goastorage.UpdatePayload) (res *goastorage.UpdateResult, err error)
}

type serviceImpl struct {
	logger logr.Logger
	db     *sql.DB
	tc     temporalsdk_client.Client
	config Config
	bucket *blob.Bucket
}

var _ Service = (*serviceImpl)(nil)

func NewService(logger logr.Logger, db *sql.DB, tc temporalsdk_client.Client, config Config) (*serviceImpl, error) {
	s := &serviceImpl{
		logger: logger,
		db:     db,
		tc:     tc,
		config: config,
	}

	var err error
	s.bucket, err = s.openBucket(&config)
	if err != nil {
		return nil, fmt.Errorf("error opening bucket: %v", err)
	}

	return s, nil
}

func (s *serviceImpl) openBucket(config *Config) (*blob.Bucket, error) {
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
	return s3blob.OpenBucket(context.Background(), sess, s.config.Bucket, nil)
}

func (s *serviceImpl) Submit(ctx context.Context, payload *goastorage.SubmitPayload) (*goastorage.SubmitResult, error) {
	workflowReq := &StorageWorkflowRequest{}
	exec, err := InitStorageWorkflow(ctx, s.tc, workflowReq)
	if err != nil {
		return nil, err
	}

	url, err := s.bucket.SignedURL(ctx, uuid.New().String(), &blob.SignedURLOptions{Expiry: urlExpirationTime, Method: http.MethodPut})
	if err != nil {
		return nil, err
	}

	result := &goastorage.SubmitResult{
		URL:        url,
		WorkflowID: exec.GetID(),
	}
	return result, nil
}

func (s *serviceImpl) Update(ctx context.Context, payload *goastorage.UpdatePayload) (*goastorage.UpdateResult, error) {
	signal := StorageWorkflowSignal{}

	err := s.tc.SignalWorkflow(context.Background(), payload.WorkflowID, "", StorageWorkflowSignalName, signal)
	if err != nil {
		return nil, err
	}

	result := &goastorage.UpdateResult{OK: true}
	return result, nil
}

func SetBucket(s *serviceImpl, b *blob.Bucket) {
	s.bucket = b
}
