package activities_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"go.artefactual.dev/tools/mockutil"
	"go.artefactual.dev/tools/temporal"
	"go.opentelemetry.io/otel/trace/noop"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	watcherfake "github.com/artefactual-sdps/enduro/internal/watcher/fake"
	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

func TestDownloadActivity(t *testing.T) {
	key := "jabber.txt"
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
				var T string
				r.Download(mockutil.Context(), gomock.AssignableToTypeOf(T), watcherName, key).
					DoAndReturn(func(ctx context.Context, dest, watcherName, key string) error {
						w, err := os.Create(dest)
						if err != nil {
							return err
						}

						_, err = w.Write([]byte("’Twas brillig, and the slithy toves Did gyre and gimble in the wabe:"))
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
				var T string
				r.Download(mockutil.Context(), gomock.AssignableToTypeOf(T), watcherName, key).Return(
					fmt.Errorf("error loading watcher: unknown watcher %s", watcherName),
				)
			},
			wantErr: fmt.Sprintf("activity error (type: download-activity, scheduledEventID: 0, startedEventID: 0, identity: ): download: error loading watcher: unknown watcher %s", watcherName),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			wsvc := watcherfake.NewMockService(gomock.NewController(t))
			if tt.rec != nil {
				tt.rec(wsvc.EXPECT())
			}

			env.RegisterActivityWithOptions(
				activities.NewDownloadActivity(logr.Discard(), noop.Tracer{}, wsvc).Execute,
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
