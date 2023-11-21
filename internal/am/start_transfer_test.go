package am_test

import (
	"net/http"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"go.artefactual.dev/amclient"
	"go.artefactual.dev/amclient/amclienttest"
	"go.artefactual.dev/tools/mockutil"
	"go.artefactual.dev/tools/temporal"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/am"
)

func TestStartTransferActivity(t *testing.T) {
	transferID := uuid.New().String()
	opts := am.StartTransferActivityParams{
		Name: "Testing",
		Path: "/tmp",
	}

	amcrDefault := func(m *amclienttest.MockPackageServiceMockRecorder, st http.Response) {
		m.Create(
			mockutil.Context(),
			&amclient.PackageCreateRequest{
				Name:             opts.Name,
				Type:             "zipfile",
				Path:             opts.Path,
				ProcessingConfig: "automated",
				AutoApprove:      true,
			},
		).Return(
			&amclient.PackageCreateResponse{ID: transferID},
			&amclient.Response{Response: &st},
			&amclient.ErrorResponse{Response: &st},
		)
	}

	type test struct {
		name   string
		want   am.StartTransferActivityResult
		amcr   func(*amclienttest.MockPackageServiceMockRecorder, http.Response)
		st     http.Response
		errMsg string
	}
	for _, tt := range []test{
		{
			name: "Returns transfer ID",
			amcr: func(mpsmr *amclienttest.MockPackageServiceMockRecorder, r http.Response) {
				mpsmr.Create(
					mockutil.Context(),
					&amclient.PackageCreateRequest{
						Name:             opts.Name,
						Type:             "zipfile",
						Path:             opts.Path,
						ProcessingConfig: "automated",
						AutoApprove:      true,
					},
				).Return(
					&amclient.PackageCreateResponse{ID: transferID},
					&amclient.Response{Response: &r},
					nil,
				)
			},
			want: am.StartTransferActivityResult{UUID: transferID},
		},
		{
			name:   "Returns an invalid credentials error",
			amcr:   amcrDefault,
			st:     http.Response{StatusCode: http.StatusUnauthorized},
			errMsg: "invalid Archivematica credentials",
		},
		{
			name:   "Returns an insufficient permissions error",
			amcr:   amcrDefault,
			st:     http.Response{StatusCode: http.StatusForbidden},
			errMsg: "insufficient Archivematica permissions",
		},
		{
			name:   "Returns a not found error",
			amcr:   amcrDefault,
			st:     http.Response{StatusCode: http.StatusNotFound},
			errMsg: "Archivematica transfer not found",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			ctrl := gomock.NewController(t)
			amps := amclienttest.NewMockPackageService(ctrl)

			if tt.amcr != nil {
				tt.amcr(amps.EXPECT(), tt.st)
			}

			env.RegisterActivityWithOptions(
				am.NewStartTransferActivity(logr.Discard(), &am.Config{}, amps).Execute,
				temporalsdk_activity.RegisterOptions{
					Name: am.StartTransferActivityName,
				},
			)

			future, err := env.ExecuteActivity(am.StartTransferActivityName, opts)
			if tt.errMsg != "" {
				assert.ErrorContains(t, err, tt.errMsg)
				assert.Assert(t, temporal.NonRetryableError(err))

				return
			}

			var r am.StartTransferActivityResult
			err = future.Get(&r)
			assert.NilError(t, err)
			assert.DeepEqual(t, r, am.StartTransferActivityResult{UUID: transferID})
		})
	}
}
