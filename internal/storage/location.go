package storage

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

var LocationFactory = func(cfg LocationConfig) (Location, error) {
	return NewLocation(cfg)
}

type Location interface {
	Name() string
	Bucket() *blob.Bucket
	SetBucket(*blob.Bucket)
}

type locationImpl struct {
	name   string
	config LocationConfig
	bucket *blob.Bucket
}

func NewLocation(config LocationConfig) (*locationImpl, error) {
	l := &locationImpl{
		name:   config.Name,
		config: config,
	}

	if b, err := l.openBucket(); err != nil {
		return nil, err
	} else {
		l.bucket = b
	}

	return l, nil
}

func (l *locationImpl) Name() string {
	return l.name
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
