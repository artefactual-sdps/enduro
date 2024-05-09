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
				URL: "mem://",
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
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			loc, err := storage.NewInternalLocation(tc.config)
			if tc.errMsg != "" {
				assert.Error(t, err, tc.errMsg)
				return
			}
			assert.NilError(t, err)
			assert.Equal(t, loc.UUID(), uuid.Nil)
		})
	}
}

func TestNewLocation(t *testing.T) {
	t.Parallel()

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
					URL: "mem://",
				},
			},
			uuid: locationID,
		},
		{
			name: "Returns an S3 location",
			location: &goastorage.Location{
				UUID: locationID,
				Config: &goastorage.S3Config{
					Bucket: "perma-aips-1",
					Region: "planet-earth",
				},
			},
			uuid: locationID,
		},
		{
			name: "Returns an SFTP location",
			location: &goastorage.Location{
				UUID: locationID,
				Config: &goastorage.SFTPConfig{
					Address:   "sftp.example.com",
					Username:  "test",
					Password:  "Test123!",
					Directory: "deposit",
				},
			},
			uuid: locationID,
		},
		{
			name: "Returns an AMSS location",
			location: &goastorage.Location{
				UUID: locationID,
				Config: &goastorage.AMSSConfig{
					APIKey:   "Secret1",
					URL:      "http://localhost:8080",
					Username: "test",
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
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			loc, err := storage.NewLocation(tc.location)
			if tc.errMsg != "" {
				assert.ErrorContains(t, err, tc.errMsg)
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, loc.UUID(), tc.uuid)
		})
	}
}

func TestLocation_Bucket(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		location *goastorage.Location
		uuid     uuid.UUID
		errMsg   string
	}{
		{
			name: "Returns a URL config bucket",
			location: &goastorage.Location{
				UUID: locationID,
				Config: &goastorage.URLConfig{
					URL: "mem://",
				},
			},
			uuid: locationID,
		},
		{
			name: "Errors on an invalid bucket driver",
			location: &goastorage.Location{
				UUID: locationID,
				Config: &goastorage.URLConfig{
					URL: "foo://test-bucket",
				},
			},
			errMsg: `open bucket by URL: open blob.Bucket: no driver registered for "foo" for URL "foo://test-bucket"`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			loc, err := storage.NewLocation(tc.location)

			assert.NilError(t, err)

			b, err := loc.OpenBucket(context.Background())
			if tc.errMsg != "" {
				assert.ErrorContains(t, err, tc.errMsg)
				return
			}
			defer b.Close()

			assert.Assert(t, b != nil)
		})
	}
}
