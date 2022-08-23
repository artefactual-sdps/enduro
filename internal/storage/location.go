package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/google/uuid"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
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

	var config *types.S3Config
	switch c := location.Config.(type) {
	case *goastorage.S3Config:
		config = c.ConvertToS3Config()
	default:
		return nil, fmt.Errorf("unsupported config type: %T", c)
	}

	if !config.Valid() {
		return nil, errors.New("invalid configuration")
	}

	l.config = LocationConfig{
		Region:    config.Region,
		Endpoint:  config.Endpoint,
		PathStyle: config.PathStyle,
		Profile:   config.Profile,
		Key:       config.Key,
		Secret:    config.Secret,
		Token:     config.Token,
		Bucket:    config.Bucket,
	}

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
