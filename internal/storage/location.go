package storage

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/google/uuid"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

var LocationFactory = func(cfg LocationConfig) (Location, error) {
	return NewInternalLocation(cfg)
}

type Location interface {
	UUID() uuid.UUID
	Bucket() *blob.Bucket
	SetBucket(*blob.Bucket)
}

type locationImpl struct {
	id     uuid.UUID
	config LocationConfig
	bucket *blob.Bucket
}

func NewInternalLocation(config LocationConfig) (*locationImpl, error) {
	l := &locationImpl{
		id:     uuid.Nil,
		config: config,
	}

	if b, err := l.openBucket(); err != nil {
		return nil, err
	} else {
		l.bucket = b
	}

	return l, nil
}

func NewLocation(location *goastorage.StoredLocation) (*locationImpl, error) {
	l := &locationImpl{
		id: *location.UUID,
	}

	// TODO: loading the S3Config, etc...

	return l, nil
}

func (l *locationImpl) UUID() uuid.UUID {
	return l.id
}

func (l *locationImpl) openBucket() (*blob.Bucket, error) {
	sessOpts := session.Options{}
	sessOpts.Config.WithRegion(l.config.Region)
	sessOpts.Config.WithEndpoint(l.config.Endpoint)
	sessOpts.Config.WithS3ForcePathStyle(l.config.PathStyle)
	sessOpts.Config.WithCredentials(
		credentials.NewStaticCredentials(
			l.config.Key, l.config.Secret, l.config.Token,
		),
	)
	sess, err := session.NewSessionWithOptions(sessOpts)
	if err != nil {
		return nil, err
	}
	return s3blob.OpenBucket(context.Background(), sess, l.config.Bucket, nil)
}

func (l *locationImpl) Bucket() *blob.Bucket {
	return l.bucket
}

func (l *locationImpl) SetBucket(b *blob.Bucket) {
	l.bucket = b
}
