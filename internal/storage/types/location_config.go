package types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/rukavina/sftpblob"
	"go.artefactual.dev/tools/bucket"
	"gocloud.dev/blob"

	"github.com/artefactual-sdps/enduro/internal/storage/ssblob"
)

type configVal interface {
	Valid() bool
	OpenBucket(context.Context) (*blob.Bucket, error)
}

type LocationConfig struct {
	Value configVal
}

type configTypes struct {
	S3         *S3Config   `json:"s3,omitempty"`
	SFTPConfig *SFTPConfig `json:"sftp,omitempty"`
	URLConfig  *URLConfig  `json:"url,omitempty"`
	SSConfig   *AMSSConfig `json:"amss,omitempty"`
}

func (c LocationConfig) MarshalJSON() ([]byte, error) {
	types := configTypes{}

	switch c := c.Value.(type) {
	case *S3Config:
		types.S3 = c
	case *SFTPConfig:
		types.SFTPConfig = c
	case *URLConfig:
		types.URLConfig = c
	case *AMSSConfig:
		types.SSConfig = c
	default:
		return nil, fmt.Errorf("unsupported config type: %T", c)
	}

	return json.Marshal(types)
}

func (c *LocationConfig) UnmarshalJSON(blob []byte) error {
	// Raise an error if the doc describes multiple configs.
	keys := map[string]json.RawMessage{}
	err := json.Unmarshal(blob, &keys)
	if err != nil {
		return errors.New("undefined configuration format")
	}
	if len(keys) > 1 {
		return errors.New("multiple config values have been assigned")
	}

	types := configTypes{}
	if err := json.Unmarshal(blob, &types); err != nil {
		return err
	}

	switch {
	case types.S3 != nil:
		c.Value = types.S3
	case types.SFTPConfig != nil:
		c.Value = types.SFTPConfig
	case types.URLConfig != nil:
		c.Value = types.URLConfig
	case types.SSConfig != nil:
		c.Value = types.SSConfig

	default:
		return errors.New("undefined configuration document")
	}

	return nil
}

type S3Config struct {
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	Endpoint  string `json:"endpoint,omitempty"`
	PathStyle bool   `json:"path_style,omitempty"`
	Profile   string `json:"profile,omitempty"`
	Key       string `json:"key,omitempty"`
	Secret    string `json:"secret,omitempty"`
	Token     string `json:"token,omitempty"`
}

func (c S3Config) Valid() bool {
	if c.Bucket == "" || c.Region == "" {
		return false
	}

	return true
}

func (c S3Config) OpenBucket(ctx context.Context) (*blob.Bucket, error) {
	return bucket.NewWithConfig(ctx, &bucket.Config{
		Endpoint:  c.Endpoint,
		Bucket:    c.Bucket,
		AccessKey: c.Key,
		SecretKey: c.Secret,
		Token:     c.Token,
		Profile:   c.Profile,
		Region:    c.Region,
		PathStyle: c.PathStyle,
	})
}

type SFTPConfig struct {
	Address   string `json:"address"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Directory string `json:"directory"`
}

func (c SFTPConfig) Valid() bool {
	if c.Address == "" || c.Username == "" || c.Password == "" {
		return false
	}

	return true
}

func (c SFTPConfig) OpenBucket(ctx context.Context) (*blob.Bucket, error) {
	sftpUrl := fmt.Sprintf("sftp://%s:%s@%s/%s", c.Username, c.Password, c.Address, c.Directory)

	url, err := url.Parse(sftpUrl)
	if err != nil {
		return nil, err
	}

	return sftpblob.OpenBucket(url, nil)
}

type URLConfig struct {
	URL string `json:"url"`
}

func (c URLConfig) Valid() bool {
	return c.URL != ""
}

func (c URLConfig) OpenBucket(ctx context.Context) (*blob.Bucket, error) {
	b, err := blob.OpenBucket(ctx, c.URL)
	if err != nil {
		return nil, fmt.Errorf("open bucket by URL: %v", err)
	}
	return b, nil
}

type AMSSConfig struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	APIKey   string `json:"api_key"`
}

func (c AMSSConfig) Valid() bool {
	return c.URL != "" && c.Username != "" && c.APIKey != ""
}

func (c AMSSConfig) OpenBucket(ctx context.Context) (*blob.Bucket, error) {
	opts := ssblob.Options{
		URL:      c.URL,
		Username: c.Username,
		Key:      c.APIKey,
	}
	b, err := ssblob.OpenBucket(&opts)
	if err != nil {
		return nil, fmt.Errorf("open bucket by Storage Service: %v", err)
	}
	return b, nil
}
