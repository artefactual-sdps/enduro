package storage

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"gocloud.dev/blob"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

type Location interface {
	UUID() uuid.UUID
	OpenBucket(ctx context.Context) (*blob.Bucket, error)
	// OpenSS or Open Storage Service is like Open Bucket but instead of a bucket it returns the
	// http response from the external storage service it has requested content from.
	OpenSS(ctx context.Context) (*http.Response, error)
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

type NotExternLocation error

func (l *locationImpl) OpenSS(ctx context.Context) (*http.Response, error) {
	// Open ss takes the location url on the config and resolves the location via a new request to that url
	// not sure if I can even get this, it is a bit confusing to understand how to grab the config
	// url := l
	// resp, err, := http.NewRequestWithContext(ctx, l, nil)
	// return resp, nil

	var foo bool
	if !foo {
		var err NotExternLocation
		return nil, err
	}
	panic("not implemented")
}
