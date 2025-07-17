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
				Bucket: validBucket,
				Name:   validName,
			},
			wantErrs: nil,
		},
		{
			name: "invalid ID - nil UUID",
			config: sipsource.Config{
				ID:     uuid.Nil,
				Bucket: validBucket,
				Name:   validName,
			},
			wantErrs: []error{
				sipsource.ErrInvalidID,
			},
		},
		{
			name: "invalid bucket config - nil bucket",
			config: sipsource.Config{
				ID:     validID,
				Bucket: nil,
				Name:   validName,
			},
			wantErrs: []error{
				sipsource.ErrInvalidBucket,
			},
		},
		{
			name: "invalid name - empty string",
			config: sipsource.Config{
				ID:     validID,
				Bucket: validBucket,
				Name:   "",
			},
			wantErrs: []error{
				sipsource.ErrInvalidName,
			},
		},
		{
			name: "multiple validation errors - nil UUID and empty name",
			config: sipsource.Config{
				ID:     uuid.Nil,
				Bucket: validBucket,
				Name:   "",
			},
			wantErrs: []error{
				sipsource.ErrInvalidID,
				sipsource.ErrInvalidName,
			},
		},
		{
			name: "multiple validation errors - nil bucket and empty name",
			config: sipsource.Config{
				ID:     validID,
				Bucket: nil,
				Name:   "",
			},
			wantErrs: []error{
				sipsource.ErrInvalidBucket,
				sipsource.ErrInvalidName,
			},
		},
		{
			name: "all invalid fields",
			config: sipsource.Config{
				ID:     uuid.Nil,
				Bucket: nil,
				Name:   "",
			},
			wantErrs: []error{
				sipsource.ErrInvalidID,
				sipsource.ErrInvalidBucket,
				sipsource.ErrInvalidName,
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
