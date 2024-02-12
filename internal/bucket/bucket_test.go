package bucket_test

import (
	"context"
	"testing"

	s3v2 "github.com/aws/aws-sdk-go-v2/service/s3"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/bucket"
)

func TestOpen(t *testing.T) {
	t.Parallel()

	type test struct {
		config  *bucket.Config
		errMsg  string
		require func(*blob.Bucket)
	}
	tests := map[string]test{
		"Opens URL-based config": {
			config: &bucket.Config{
				URL: "mem://",
			},
		},
		"Opens attr-based config": {
			config: &bucket.Config{
				Endpoint:  "http://foobar:12345",
				Bucket:    "name",
				Region:    "region",
				AccessKey: "access",
				SecretKey: "secret",
				PathStyle: true,
			},
			require: func(b *blob.Bucket) {
				var client *s3v2.Client
				assert.Equal(t, b.As(&client), true)

				opts := client.Options()
				assert.Equal(t, opts.Region, "region")
				assert.Equal(t, opts.UsePathStyle, true)

				_, err := client.ListBuckets(context.Background(), &s3v2.ListBucketsInput{})
				assert.ErrorContains(t, err, "http://foobar:12345/?x-id=ListBuckets")
			},
		},
		"Appends http if scheme is undefined": {
			config: &bucket.Config{
				Endpoint:  "foobar:12345",
				Bucket:    "name",
				Region:    "region",
				AccessKey: "access",
				SecretKey: "secret",
			},
			require: func(b *blob.Bucket) {
				var client *s3v2.Client
				assert.Equal(t, b.As(&client), true)

				_, err := client.ListBuckets(context.Background(), &s3v2.ListBucketsInput{})
				assert.ErrorContains(t, err, "http://foobar:12345/?x-id=ListBuckets")
			},
		},
		"Rejects nil config": {
			config: nil,
			errMsg: "config is undefined",
		},
		"Rejects non-existent shared config profile": {
			config: &bucket.Config{
				Profile: "profile",
			},
			errMsg: "load AWS default config: failed to get shared config profile, profile",
		},
		"Rejects URL-based config with unknown scheme": {
			config: &bucket.Config{
				URL: "unknown://",
			},
			errMsg: `open bucket from URL "unknown://": open blob.Bucket: no driver registered for "unknown" for URL "unknown:"; available schemes: file, mem, s3`,
		},
		"Rejects bucket with empty name": {
			config: &bucket.Config{
				Endpoint:  "foobar:12345",
				Bucket:    "",
				Region:    "region",
				AccessKey: "access",
				SecretKey: "secret",
			},
			errMsg: "open bucket: s3blob.OpenBucket: bucketName is required",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			b, err := bucket.Open(context.Background(), tc.config)
			if b != nil {
				defer b.Close()
			}

			if tc.errMsg != "" {
				assert.Assert(t, b == nil)
				assert.Error(t, err, tc.errMsg)
				return
			}
			assert.NilError(t, err)

			if tc.require != nil {
				tc.require(b)
			}
		})
	}
}
