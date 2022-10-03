package types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/rukavina/sftpblob"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

type LocationConfig struct {
	Value configVal
}

func (c LocationConfig) MarshalJSON() ([]byte, error) {
	types := configTypes{}

	switch c := c.Value.(type) {
	case *S3Config:
		types.S3 = c
	case *SFTPConfig:
		types.SFTPConfig = c
	default:
		return nil, fmt.Errorf("unsupported config type: %T", c)
	}

	return json.Marshal(types)
}

func (c *LocationConfig) UnmarshalJSON(blob []byte) error {
	types := configTypes{}

	if err := json.Unmarshal(blob, &types); err != nil {
		return err
	}

	switch {
	// TODO: return error if we have more than one config assigned (mutually exclusive)
	case types.S3 != nil:
		c.Value = types.S3
	case types.SFTPConfig != nil:
		c.Value = types.SFTPConfig
	default:
		return errors.New("undefined configuration document")
	}

	return nil
}

type configVal interface {
	Valid() bool
	OpenBucket() (*blob.Bucket, error)
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

func (c S3Config) OpenBucket() (*blob.Bucket, error) {
	sessOpts := session.Options{}
	sessOpts.Config.WithRegion(c.Region)
	sessOpts.Config.WithEndpoint(c.Endpoint)
	sessOpts.Config.WithS3ForcePathStyle(c.PathStyle)
	sessOpts.Config.WithCredentials(
		credentials.NewStaticCredentials(
			c.Key, c.Secret, c.Token,
		),
	)
	sess, err := session.NewSessionWithOptions(sessOpts)
	if err != nil {
		return nil, err
	}
	return s3blob.OpenBucket(context.Background(), sess, c.Bucket, nil)
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

func (c SFTPConfig) OpenBucket() (*blob.Bucket, error) {
	sftpUrl := fmt.Sprintf("sftp://%s:%s@%s/%s", c.Username, c.Password, c.Address, c.Directory)

	url, err := url.Parse(sftpUrl)
	if err != nil {
		return nil, err
	}

	return sftpblob.OpenBucket(url, nil)
}

type configTypes struct {
	S3         *S3Config   `json:"s3,omitempty"`
	SFTPConfig *SFTPConfig `json:"sftp,omitempty"`
}
