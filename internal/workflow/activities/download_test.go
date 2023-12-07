package activities_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"go.artefactual.dev/tools/mockutil"
	"go.artefactual.dev/tools/temporal"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	watcherfake "github.com/artefactual-sdps/enduro/internal/watcher/fake"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

func TestDownloadActivity(t *testing.T) {
	key := "transfer.zip"
	watcherName := "watcher"

	type test struct {
		name    string
		params  *activities.DownloadActivityParams
		rec     func(*watcherfake.MockServiceMockRecorder)
		want    string
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Downloads blob to a temp dir",
			params: &activities.DownloadActivityParams{
				Key:         key,
				WatcherName: watcherName,
			},
			rec: func(r *watcherfake.MockServiceMockRecorder) {
				r.Download(mockutil.Context(), gomock.AssignableToTypeOf((*os.File)(nil)), watcherName, key).
					DoAndReturn(func(ctx context.Context, w io.Writer, watcherName, key string) error {
						_, err := w.Write([]byte("’Twas brillig, and the slithy toves Did gyre and gimble in the wabe:"))
						return err
					})
			},
			want: "’Twas brillig, and the slithy toves Did gyre and gimble in the wabe:",
		},
		{
			name: "Non-retryable error when download fails",
			params: &activities.DownloadActivityParams{
				Key:         key,
				WatcherName: watcherName,
			},
			rec: func(r *watcherfake.MockServiceMockRecorder) {
				r.Download(mockutil.Context(), gomock.AssignableToTypeOf((*os.File)(nil)), watcherName, key).Return(
					fmt.Errorf("error loading watcher: unknown watcher %s", watcherName),
				)
			},
			wantErr: fmt.Sprintf("activity error (type: download-activity, scheduledEventID: 0, startedEventID: 0, identity: ): download blob: error loading watcher: unknown watcher %s", watcherName),
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			wsvc := watcherfake.NewMockService(gomock.NewController(t))
			if tt.rec != nil {
				tt.rec(wsvc.EXPECT())
			}

			env.RegisterActivityWithOptions(
				activities.NewDownloadActivity(logr.Discard(), wsvc).Execute,
				temporalsdk_activity.RegisterOptions{
					Name: activities.DownloadActivityName,
				},
			)

			enc, err := env.ExecuteActivity(activities.DownloadActivityName, tt.params)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				assert.Assert(t, temporal.NonRetryableError(err))
				return
			}
			assert.NilError(t, err)

			var res activities.DownloadActivityResult
			_ = enc.Get(&res)

			got, err := os.ReadFile(res.Path)
			assert.NilError(t, err)
			assert.Equal(t, string(got), tt.want)
		})
	}
}
