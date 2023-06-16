package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gocloud.dev/blob"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

type InternalLocationFactory func(config *LocationConfig) (Location, error)

var DefaultInternalLocationFactory = func(config *LocationConfig) (Location, error) {
	return NewInternalLocation(config)
}

type LocationFactory func(location *goastorage.Location) (Location, error)

var DefaultLocationFactory = func(location *goastorage.Location) (Location, error) {
	return NewLocation(location)
}

type Location interface {
	UUID() uuid.UUID
	Bucket() *blob.Bucket
	Close() error
}

type locationImpl struct {
	id     uuid.UUID
	bucket *blob.Bucket
}

func NewInternalLocation(config *LocationConfig) (*locationImpl, error) {
	var (
		b   *blob.Bucket
		err error
	)

	if config.URL != "" {
		b, err = openURLBucket(config)
	} else {
		b, err = openS3Bucket(config)
	}
	if err != nil {
		return nil, err
	}

	return &locationImpl{
		id:     uuid.Nil,
		bucket: b,
	}, nil
}

func openURLBucket(config *LocationConfig) (*blob.Bucket, error) {
	c := types.URLConfig{URL: config.URL}
	if !c.Valid() {
		return nil, errors.New("invalid configuration")
	}

	b, err := c.OpenBucket(context.Background())
	if err != nil {
		return nil, err
	}

	return b, nil
}

func openS3Bucket(config *LocationConfig) (*blob.Bucket, error) {
	s3Config := &types.S3Config{
		Region:    config.Region,
		Endpoint:  config.Endpoint,
		PathStyle: config.PathStyle,
		Profile:   config.Profile,
		Key:       config.Key,
		Secret:    config.Secret,
		Token:     config.Token,
		Bucket:    config.Bucket,
	}
	if !s3Config.Valid() {
		return nil, errors.New("invalid configuration")
	}

	b, err := s3Config.OpenBucket(context.Background())
	if err != nil {
		return nil, err
	}

	return b, nil
}

func NewLocation(location *goastorage.Location) (*locationImpl, error) {
	l := &locationImpl{
		id: location.UUID,
	}

	var config types.LocationConfig
	switch c := location.Config.(type) {
	case *goastorage.URLConfig:
		config.Value = c.ConvertToURLConfig()
	case *goastorage.S3Config:
		config.Value = c.ConvertToS3Config()
	case *goastorage.SFTPConfig:
		config.Value = c.ConvertToSFTPConfig()
	default:
		return nil, fmt.Errorf("unsupported config type: %T", c)
	}

	if !config.Value.Valid() {
		return nil, errors.New("invalid configuration")
	}

	if b, err := config.Value.OpenBucket(context.Background()); err != nil {
		return nil, err
	} else {
		l.bucket = b
	}

	return l, nil
}

func (l *locationImpl) UUID() uuid.UUID {
	return l.id
}

func (l *locationImpl) Bucket() *blob.Bucket {
	return l.bucket
}

func (l *locationImpl) Close() error {
	return l.bucket.Close()
}
