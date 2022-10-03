package types

import (
	"encoding/json"
	"errors"
	"fmt"
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

type configTypes struct {
	S3         *S3Config   `json:"s3,omitempty"`
	SFTPConfig *SFTPConfig `json:"sftp,omitempty"`
}
