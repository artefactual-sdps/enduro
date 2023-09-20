// Package bucket provides a mechanism to open buckets using configuration.
package bucket

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

type Config struct {
	// Bucket URL as described in https://gocloud.dev/howto/blob/.
	URL string

	// Alternatively, a less flexible way to access S3-compatible buckets.
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	Token     string
	Profile   string
	Region    string
	PathStyle bool
}

// Open a bucket provided its configuration. Unless specified in the URL field,
// it will open buckets with s3blob.OpenBucketV2.
func Open(ctx context.Context, c *Config) (*blob.Bucket, error) {
	if c == nil {
		return nil, errors.New("config is undefined")
	}

	if c.URL != "" {
		b, err := blob.OpenBucket(ctx, c.URL)
		if err != nil {
			return nil, fmt.Errorf("open bucket from URL %q: %v", c.URL, err)
		}
		return b, nil
	}

	addr := c.Endpoint
	if u, err := url.Parse(c.Endpoint); err == nil {
		if !strings.HasPrefix(u.Scheme, "http") {
			addr = "http://" + addr
		}
	}

	awscfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithSharedConfigProfile(c.Profile),
		config.WithRegion(c.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				c.AccessKey, c.SecretKey, c.Token,
			),
		),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					if service == s3.ServiceID {
						return aws.Endpoint{URL: addr}, nil
					}
					return aws.Endpoint{}, &aws.EndpointNotFoundError{}
				},
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load AWS default config: %v", err)
	}

	client := s3.NewFromConfig(awscfg, func(opts *s3.Options) {
		opts.UsePathStyle = c.PathStyle
	})
	b, err := s3blob.OpenBucketV2(ctx, client, c.Bucket, nil)
	if err != nil {
		return nil, fmt.Errorf("open bucket: %v", err)
	}

	return b, nil
}
