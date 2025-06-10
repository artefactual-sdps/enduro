package activities_test

import (
	"fmt"
	"path/filepath"
	"testing"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/workflow/activities"
)

func testZIP(t *testing.T) string {
	d := fs.NewDir(t, "enduro-test", fs.WithFile("sip", "\x50\x4b\x03\x04"))
	return filepath.Join(d.Path(), "sip")
}

func testTXT(t *testing.T) string {
	d := fs.NewDir(t, "enduro-test", fs.WithFile("sip.txt", ""))
	return filepath.Join(d.Path(), "sip.txt")
}

func TestGetSIPExtensionActivity(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		params  *activities.GetSIPExtensionActivityParams
		want    *activities.GetSIPExtensionActivityResult
		wantErr string
	}
	for _, tt := range []test{
		{
			name:   "Returns the SIP file extension",
			params: &activities.GetSIPExtensionActivityParams{Path: testZIP(t)},
			want:   &activities.GetSIPExtensionActivityResult{Extension: ".zip"},
		},
		{
			name:    "Fails to return the extension of a missing SIP file",
			params:  &activities.GetSIPExtensionActivityParams{Path: "/missing/sip.zip"},
			wantErr: fmt.Sprintf("%s: open SIP file:", activities.GetSIPExtensionActivityName),
		},
		{
			name:    "Fails to return the extension of an unrecognized SIP file",
			params:  &activities.GetSIPExtensionActivityParams{Path: testTXT(t)},
			wantErr: activities.ErrInvalidArchive.Error(),
		},
		{
			name:    "Fails to return the extension of a directory",
			params:  &activities.GetSIPExtensionActivityParams{Path: fs.NewDir(t, "enduro-test").Path()},
			wantErr: fmt.Sprintf("%s: identify SIP format:", activities.GetSIPExtensionActivityName),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				activities.NewGetSIPExtensionActivity().Execute,
				temporalsdk_activity.RegisterOptions{Name: activities.GetSIPExtensionActivityName},
			)
			enc, err := env.ExecuteActivity(activities.GetSIPExtensionActivityName, tt.params)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			var res activities.GetSIPExtensionActivityResult
			err = enc.Get(&res)
			assert.NilError(t, err)
			assert.DeepEqual(t, &res, tt.want)
		})
	}
}
