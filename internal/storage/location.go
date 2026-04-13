package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/bucket"
	"gocloud.dev/blob"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

type Location interface {
	UUID() uuid.UUID
	OpenBucket(ctx context.Context) (*blob.Bucket, error)
}

type locationImpl struct {
	id           uuid.UUID
	config       *types.LocationConfig
	bucketConfig *bucket.Config
}

var _ Location = (*locationImpl)(nil)

func NewInternalLocation(ctx context.Context, config *bucket.Config) (Location, error) {
	// Open the bucket to validate the configuration.
	b, err := bucket.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("NewInternalLocation: %v", err)
	}
	b.Close()

	return &locationImpl{
		id:           uuid.Nil,
		bucketConfig: config,
	}, nil
}

func NewLocation(location *goastorage.Location) (Location, error) {
	config, err := ConvertGoaLocationConfigToLocationConfig(location.Config)
	if err != nil {
		return nil, err
	}

	return &locationImpl{
		id:     location.UUID,
		config: &config,
	}, nil
}

func (l *locationImpl) UUID() uuid.UUID {
	return l.id
}

func (l *locationImpl) OpenBucket(ctx context.Context) (*blob.Bucket, error) {
	var b *blob.Bucket
	var err error
	if l.bucketConfig != nil {
		b, err = bucket.NewWithConfig(ctx, l.bucketConfig)
	} else if l.config != nil {
		b, err = l.config.Value.OpenBucket(ctx)
	} else {
		err = errors.New("no configuration available to open bucket")
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

func ConvertGoaLocationConfigToLocationConfig(goaConfig any) (types.LocationConfig, error) {
	var config types.LocationConfig
	switch c := goaConfig.(type) {
	case *goastorage.URLConfig:
		config.Value = c.ConvertToURLConfig()
	case *goastorage.S3Config:
		config.Value = c.ConvertToS3Config()
	case *goastorage.SFTPConfig:
		config.Value = c.ConvertToSFTPConfig()
	case *goastorage.AMSSConfig:
		config.Value = c.ConvertToAMSSConfig()
	default:
		return types.LocationConfig{}, fmt.Errorf("unsupported config type: %T", c)
	}

	if !config.Value.Valid() {
		return types.LocationConfig{}, errors.New("invalid configuration")
	}

	return config, nil
}
