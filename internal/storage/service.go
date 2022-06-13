package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-logr/logr"
	temporalsdk_client "go.temporal.io/sdk/client"

	goastorage "github.com/artefactual-labs/enduro/internal/api/gen/storage"
)

var urlExpirationTime = 15 * time.Minute

type Service interface {
	Submit(context.Context, *goastorage.SubmitPayload) (res *goastorage.SubmitResult, err error)
	Update(context.Context, *goastorage.UpdatePayload) (res *goastorage.UpdateResult, err error)
}

type storageImpl struct {
	logger logr.Logger
	db     *sql.DB
	tc     temporalsdk_client.Client
	config Config
	s3     *s3.S3
}

var _ Service = (*storageImpl)(nil)

func NewService(logger logr.Logger, db *sql.DB, tc temporalsdk_client.Client, config Config) (*storageImpl, error) {
	s := &storageImpl{
		logger: logger,
		db:     db,
		tc:     tc,
		config: config,
	}

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

	s.s3 = s3.New(sess)

	return s, nil
}

func (s *storageImpl) Submit(ctx context.Context, payload *goastorage.SubmitPayload) (*goastorage.SubmitResult, error) {
	if payload.Key == "" {
		return nil, goastorage.MakeNotValid(errors.New("error starting workflow - key is empty"))
	}

	workflowReq := &StorageWorkflowRequest{}
	exec, err := InitStorageWorkflow(ctx, s.tc, workflowReq)
	if err != nil {
		return nil, err
	}

	req, _ := s.s3.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String("aips"),
		Key:    &payload.Key,
	})
	url, err := req.Presign(urlExpirationTime)
	if err != nil {
		return nil, err
	}

	result := &goastorage.SubmitResult{
		URL:        url,
		WorkflowID: exec.GetID(),
	}
	return result, nil
}

func (s *storageImpl) Update(ctx context.Context, payload *goastorage.UpdatePayload) (*goastorage.UpdateResult, error) {
	signal := StorageWorkflowSignal{}

	err := s.tc.SignalWorkflow(context.Background(), payload.WorkflowID, "", StorageWorkflowSignalName, signal)
	if err != nil {
		return nil, err
	}

	result := &goastorage.UpdateResult{OK: true}
	return result, nil
}
