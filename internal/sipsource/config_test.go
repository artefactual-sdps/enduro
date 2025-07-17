package sipsource_test

import (
	"testing"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/bucket"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/sipsource"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	validID := uuid.New()
	validBucket := &bucket.Config{
		URL: "s3://test-bucket",
	}
	validName := "test-source"

	tests := []struct {
		name     string
		config   sipsource.Config
		wantErrs []error
	}{
		{
			name: "valid config",
			config: sipsource.Config{
				ID:     validID,
				Name:   validName,
				Bucket: validBucket,
			},
			wantErrs: nil,
		},
		{
			name:     "empty config",
			config:   sipsource.Config{},
			wantErrs: nil,
		},
		{
			name: "invalid ID - nil UUID",
			config: sipsource.Config{
				ID:     uuid.Nil,
				Name:   validName,
				Bucket: validBucket,
			},
			wantErrs: []error{
				sipsource.ErrMissingID,
			},
		},
		{
			name: "invalid bucket config - nil bucket",
			config: sipsource.Config{
				ID:     validID,
				Name:   validName,
				Bucket: nil,
			},
			wantErrs: []error{
				sipsource.ErrMissingBucket,
			},
		},
		{
			name: "invalid name - empty string",
			config: sipsource.Config{
				ID:     validID,
				Name:   "",
				Bucket: validBucket,
			},
			wantErrs: []error{
				sipsource.ErrMissingName,
			},
		},
		{
			name: "invalid bucket config - empty URL and endpoint",
			config: sipsource.Config{
				ID:     validID,
				Name:   validName,
				Bucket: &bucket.Config{},
			},
			wantErrs: []error{
				sipsource.ErrInvalidConfig,
			},
		},
		{
			name: "invalid bucket config - both URL and endpoint set",
			config: sipsource.Config{
				ID:   validID,
				Name: validName,
				Bucket: &bucket.Config{
					URL:      "s3://test-bucket",
					Endpoint: "https://s3.amazonaws.com",
				},
			},
			wantErrs: []error{
				sipsource.ErrInvalidConfig,
			},
		},
		{
			name: "multiple validation errors - nil UUID and empty name",
			config: sipsource.Config{
				ID:     uuid.Nil,
				Name:   "",
				Bucket: validBucket,
			},
			wantErrs: []error{
				sipsource.ErrMissingID,
				sipsource.ErrMissingName,
			},
		},
		{
			name: "multiple validation errors - nil bucket and empty name",
			config: sipsource.Config{
				ID:     validID,
				Name:   "",
				Bucket: nil,
			},
			wantErrs: []error{
				sipsource.ErrMissingName,
				sipsource.ErrMissingBucket,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()

			if tt.wantErrs != nil {
				for _, e := range tt.wantErrs {
					assert.ErrorIs(t, err, e)
				}
			} else {
				assert.NilError(t, err)
			}
		})
	}
}
