package activities

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/bucket"
	"go.artefactual.dev/tools/mockutil"
	"go.uber.org/mock/gomock"
	"gocloud.dev/blob"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/ingest/fake"
	storage_enums "github.com/artefactual-sdps/enduro/internal/storage/enums"
)

func TestUploadActivity(t *testing.T) {
	t.Parallel()

	type test struct {
		name        string
		sourceFile  bool
		createErr   error
		wantErr     string
		wantContent string
	}

	for _, tc := range []test{
		{
			name:        "Activity runs successfully",
			sourceFile:  true,
			wantContent: "contents-of-the-aip",
		},
		{
			name:       "Activity returns an error if CreateAIP fails",
			sourceFile: true,
			createErr:  errors.New("create failed"),
			wantErr:    "create failed",
		},
		{
			name:    "Activity returns an error if the source file is missing",
			wantErr: "open",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			aipUUID := uuid.New().String()
			aipName := "aip.7z"

			mockClient := fake.NewMockStorageClient(gomock.NewController(t))
			if tc.sourceFile {
				mockClient.EXPECT().
					CreateAip(
						mockutil.Context(),
						&goastorage.CreateAipPayload{
							UUID:      aipUUID,
							Name:      aipName,
							ObjectKey: aipUUID,
							Status:    storage_enums.AIPStatusPending.String(),
						},
					).
					Return(&goastorage.AIP{}, tc.createErr)
			}

			var tmpDir *fs.Dir
			if tc.sourceFile {
				tmpDir = fs.NewDir(t, "", fs.WithFile("aip.7z", "contents-of-the-aip"))
			} else {
				tmpDir = fs.NewDir(t, "")
			}
			defer tmpDir.Remove()

			sharedDir := fs.NewDir(t, "")
			defer sharedDir.Remove()
			stagingBucket, err := bucket.NewWithConfig(
				t.Context(),
				&bucket.Config{
					URL: "file://" + sharedDir.Path() + "?metadata=skip&no_tmp_dir=true",
				},
			)
			assert.NilError(t, err)
			defer stagingBucket.Close()

			activity := NewUploadActivity(mockClient, blob.PrefixedBucket(stagingBucket, "aips/"))

			_, err = activity.Execute(context.Background(), &UploadActivityParams{
				AIPPath: tmpDir.Join(aipName),
				AIPID:   aipUUID,
				Name:    aipName,
			})
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			assert.NilError(t, err)

			if tc.wantContent != "" {
				contents, err := os.ReadFile(sharedDir.Join("aips/", aipUUID))
				assert.NilError(t, err)
				assert.DeepEqual(t, string(contents), tc.wantContent)
			}
		})
	}
}
