package storage_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	_ "gocloud.dev/blob/memblob"
	"gotest.tools/v3/assert"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/storage"
)

func TestNewInternalLocation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		config    *storage.LocationConfig
		canAccess bool
		errMsg    string
	}{
		{
			name: "Returns an internal URL location",
			config: &storage.LocationConfig{
				URL: "mem:///test-bucket",
			},
			canAccess: true,
		},
		{
			name: "Errors on an empty URL location",
			config: &storage.LocationConfig{
				URL: "",
			},
			errMsg: "invalid configuration",
		},
		{
			name: "Errors on an invalid URL location",
			config: &storage.LocationConfig{
				URL: "foo:///test-bucket",
			},
			errMsg: `open bucket by URL: open blob.Bucket: no driver registered for "foo" for URL "foo:///test-bucket"; available schemes: file, mem, s3, sftp`,
		},
	}
	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			loc, err := storage.NewInternalLocation(tc.config)
			if tc.errMsg != "" {
				assert.Error(t, err, tc.errMsg)
				return
			}

			assert.NilError(t, err)
			defer loc.Close()

			y, err := loc.Bucket().IsAccessible(context.Background())
			assert.NilError(t, err)
			assert.Equal(t, y, tc.canAccess)

			err = loc.Close()
			assert.NilError(t, err)
		})
	}
}

func TestNewLocation(t *testing.T) {
	t.Parallel()

	locationID := uuid.MustParse("314bbf3e-2fb0-4d86-910f-0c1bdfeda3a1")

	tests := []struct {
		name     string
		location *goastorage.Location
		uuid     uuid.UUID
		errMsg   string
	}{
		{
			name: "Returns a URL location",
			location: &goastorage.Location{
				UUID: locationID,
				Config: &goastorage.URLConfig{
					URL: "mem:///test-bucket",
				},
			},
			uuid: locationID,
		},
		{
			name: "Errors when URL Config is empty",
			location: &goastorage.Location{
				UUID:   locationID,
				Config: &goastorage.URLConfig{},
			},
			errMsg: "invalid configuration",
		},
		{
			name: "Errors on an invalid URL schema",
			location: &goastorage.Location{
				UUID: locationID,
				Config: &goastorage.URLConfig{
					URL: "foo:///test-bucket",
				},
			},
			errMsg: `open bucket by URL: open blob.Bucket: no driver registered for "foo" for URL "foo:///test-bucket"`,
		},
	}
	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			loc, err := storage.NewLocation(tc.location)
			if tc.errMsg != "" {
				assert.ErrorContains(t, err, tc.errMsg)
				return
			}

			assert.NilError(t, err)
			defer loc.Close()

			assert.Equal(t, loc.UUID(), tc.uuid)
			assert.Assert(t, loc.Bucket() != nil)

			err = loc.Close()
			assert.NilError(t, err)
		})
	}
}
