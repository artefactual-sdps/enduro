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

type Location interface {
	UUID() uuid.UUID
	OpenBucket(ctx context.Context) (*blob.Bucket, error)
}

type locationImpl struct {
	id     uuid.UUID
	config *types.LocationConfig
}

var _ Location = (*locationImpl)(nil)

func NewInternalLocation(config *LocationConfig) (Location, error) {
	var c types.LocationConfig

	if config.URL != "" {
		c.Value = types.URLConfig{
			URL: config.URL,
		}
	} else {
		c.Value = types.S3Config{
			Bucket:    config.Bucket,
			Region:    config.Region,
			Endpoint:  config.Endpoint,
			PathStyle: config.PathStyle,
			Profile:   config.Profile,
			Key:       config.Key,
			Secret:    config.Secret,
			Token:     config.Token,
		}
	}

	if !c.Value.Valid() {
		return nil, errors.New("invalid configuration")
	}

	return &locationImpl{
		id:     uuid.Nil,
		config: &c,
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
	b, err := l.config.Value.OpenBucket(ctx)
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
