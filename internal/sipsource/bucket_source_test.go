package sipsource_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"go.artefactual.dev/tools/bucket"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/memblob"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/sipsource"
)

func TestNewBucketSource(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		cfg     *sipsource.Config
		want    *sipsource.BucketSource
		wantErr string
	}

	sourceID := uuid.New()

	for _, tt := range []test{
		{
			name: "Returns an empty source if the source configuration is empty",
			want: &sipsource.BucketSource{},
		},
		{
			name: "Returns a valid SIP source",
			cfg: &sipsource.Config{
				ID:   sourceID,
				Name: "Test SIP Source",
				Bucket: &bucket.Config{
					URL: "mem://", // Use a memory bucket for testing.
				},
			},
			want: &sipsource.BucketSource{
				ID:   sourceID,
				Name: "Test SIP Source",
			},
		},
		{
			name: "Returns an error if the bucket URL is invalid",
			cfg: &sipsource.Config{
				ID:   sourceID,
				Name: "Test SIP Source",
				Bucket: &bucket.Config{
					URL: "invalid://", // Invalid URL to trigger an error.
				},
			},
			wantErr: "SIP source: new bucket source: open bucket from URL \"invalid://\"",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			got, err := sipsource.NewBucketSource(ctx, tt.cfg)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want,
				// We can't compare the Bucket and retentionPeriod directly, so ignore them.
				cmpopts.IgnoreFields(sipsource.BucketSource{}, "Bucket", "retentionPeriod"),
			)
		})
	}
}

func TestListItems(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		cfg     *sipsource.Config
		token   []byte
		limit   int
		want    *sipsource.Page
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Returns a list of bucket items",
			cfg: &sipsource.Config{
				ID:   uuid.New(),
				Name: "Test bucket source",
				Bucket: &bucket.Config{
					URL: "mem://",
				},
			},
			want: &sipsource.Page{
				Objects: []*sipsource.Object{
					{
						Key:     "sip1",
						ModTime: time.Now(),
						Size:    int64(len("SIP 1 content")),
						IsDir:   false,
					},
					{
						Key:     "sip2",
						ModTime: time.Now(),
						Size:    int64(len("SIP 2 content")),
						IsDir:   false,
					},
				},
				Limit:     100, // Default limit.
				NextToken: nil,
			},
		},
		{
			name: "Returns the first page of bucket items",
			cfg: &sipsource.Config{
				ID:   uuid.New(),
				Name: "Test bucket source",
				Bucket: &bucket.Config{
					URL: "mem://",
				},
			},
			limit: 1,
			want: &sipsource.Page{
				Objects: []*sipsource.Object{
					{
						Key:     "sip1",
						ModTime: time.Now(),
						Size:    int64(len("SIP 1 content")),
						IsDir:   false,
					},
				},
				Limit:     1,
				NextToken: []byte("sip1"),
			},
		},
		{
			name: "Returns the second page of bucket items",
			cfg: &sipsource.Config{
				ID:   uuid.New(),
				Name: "Test bucket source",
				Bucket: &bucket.Config{
					URL: "mem://",
				},
			},
			token: []byte("sip1"),
			limit: 1,
			want: &sipsource.Page{
				Objects: []*sipsource.Object{
					{
						Key:     "sip2",
						ModTime: time.Now(),
						Size:    int64(len("SIP 2 content")),
						IsDir:   false,
					},
				},
				Limit:     1,
				NextToken: nil,
			},
		},
		{
			name:    "Returns an error if the bucket is not configured",
			wantErr: "SIP source: missing bucket",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			source, err := sipsource.NewBucketSource(ctx, tt.cfg)
			if err != nil {
				t.Fatalf("Failed to create SIP source: %v", err)
			}
			defer source.Close()

			// Write some test data to the bucket.
			if source.Bucket != nil {
				for key, value := range map[string]string{
					"sip1": "SIP 1 content",
					"sip2": "SIP 2 content",
				} {
					if err := source.Bucket.WriteAll(ctx, key, []byte(value), nil); err != nil {
						t.Fatalf("Failed to write to bucket: %v", err)
					}
				}
			}

			got, err := source.ListObjects(ctx, tt.token, tt.limit)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, got, tt.want,
				cmpopts.EquateApproxTime(1*time.Second),
				cmpopts.IgnoreUnexported(blob.ListObject{}),
			)
		})
	}
}

func TestRetentionPeriod(t *testing.T) {
	t.Parallel()

	sourceID := uuid.New()
	negativeDuration := -1 * time.Second
	zeroDuration := 0 * time.Second
	oneHour := 1 * time.Hour

	for _, tt := range []struct {
		name string
		cfg  *sipsource.Config
		want time.Duration
	}{
		{
			name: "Returns zero retention period when not configured",
			cfg: &sipsource.Config{
				ID:   sourceID,
				Name: "Test SIP Source",
				Bucket: &bucket.Config{
					URL: "mem://",
				},
			},
			want: 0,
		},
		{
			name: "Returns zero retention period when configured as 0",
			cfg: &sipsource.Config{
				ID:   sourceID,
				Name: "Test SIP Source",
				Bucket: &bucket.Config{
					URL: "mem://",
				},
				RetentionPeriod: zeroDuration,
			},
			want: zeroDuration,
		},
		{
			name: "Returns configured retention period of 1 hour",
			cfg: &sipsource.Config{
				ID:   sourceID,
				Name: "Test SIP Source",
				Bucket: &bucket.Config{
					URL: "mem://",
				},
				RetentionPeriod: oneHour,
			},
			want: oneHour,
		},
		{
			name: "Returns configured negative retention period",
			cfg: &sipsource.Config{
				ID:   sourceID,
				Name: "Test SIP Source",
				Bucket: &bucket.Config{
					URL: "mem://",
				},
				RetentionPeriod: negativeDuration,
			},
			want: negativeDuration,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			source, err := sipsource.NewBucketSource(context.Background(), tt.cfg)
			assert.NilError(t, err)
			defer source.Close()

			assert.Equal(t, source.RetentionPeriod(), tt.want)
		})
	}
}
