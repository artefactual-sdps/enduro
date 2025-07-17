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
		want    *sipsource.SIPBucketSource
		wantErr string
	}

	sourceID := uuid.New()

	for _, tt := range []test{
		{
			name: "Returns an empty source if the source configuration is empty",
			want: &sipsource.SIPBucketSource{},
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
			want: &sipsource.SIPBucketSource{
				ID:   sourceID,
				Name: "Test SIP Source",
			},
		},
		{
			name: "Returns an error if the source ID is missing",
			cfg: &sipsource.Config{
				Name:   "Test SIP Source",
				Bucket: &bucket.Config{URL: "mem://"},
			},
			wantErr: "SIP source: missing ID",
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
			wantErr: "SIP source: open bucket from URL \"invalid://\"",
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
				// We can't compare the bucket directly, so ignore it.
				cmpopts.IgnoreFields(sipsource.SIPBucketSource{}, "Bucket"),
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
				Name: "Test SIP Source",
				Bucket: &bucket.Config{
					URL: "mem://",
				},
			},
			want: &sipsource.Page{
				Items: []*sipsource.Item{
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
				Name: "Test SIP Source",
				Bucket: &bucket.Config{
					URL: "mem://",
				},
			},
			limit: 1,
			want: &sipsource.Page{
				Items: []*sipsource.Item{
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
				Name: "Test SIP Source",
				Bucket: &bucket.Config{
					URL: "mem://",
				},
			},
			token: []byte("sip1"),
			limit: 1,
			want: &sipsource.Page{
				Items: []*sipsource.Item{
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

			got, err := source.ListItems(ctx, tt.token, tt.limit)
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
