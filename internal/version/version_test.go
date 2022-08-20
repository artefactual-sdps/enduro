package version

import (
	"runtime"
	"runtime/debug"
	"testing"

	"gotest.tools/v3/assert"
)

func TestInfo(t *testing.T) {
	t.Parallel()

	t.Run("works when ReadBuildInfo is not available", func(t *testing.T) {
		t.Parallel()

		buildInfoReader = func() (info *debug.BuildInfo, ok bool) {
			return nil, false
		}

		v := Info("test")

		assert.Equal(t, v, "test version (dev-version) (commit=(dev-commit)) built on (dev-buildtime) using "+runtime.Version())
	})

	t.Run("works when ReadBuildInfo is available", func(t *testing.T) {
		t.Parallel()

		buildInfoReader = func() (info *debug.BuildInfo, ok bool) {
			return &debug.BuildInfo{
				Settings: []debug.BuildSetting{
					{
						Key:   "vcs.revision",
						Value: "12345",
					},
					{
						Key:   "vcs.modified",
						Value: "true",
					},
				},
			}, true
		}

		v := Info("test")

		assert.Equal(t, v, "test version (dev-version) (commit=!12345) built on (dev-buildtime) using "+runtime.Version())
	})
}
