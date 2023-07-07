// Package version provides the version that the binary was built at.
package version

import (
	_ "embed"
	"fmt"
	"runtime/debug"
	"strings"
)

//go:embed VERSION.txt
var version string

var (
	// Long is a full version number for this build.
	Long = ""

	// Short is a short version number for this build.
	Short = ""

	// GitCommit is the git commit of the build.
	GitCommit = ""

	// GitDirty is the vcs.modified setting returned by debug.ReadBuildInfo.
	GitDirty bool
)

func init() {
	if Long != "" && Short != "" {
		// Built in the recommended way, using build_dist.sh.
		return
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		Long = strings.TrimSpace(version) + "-ERR-BuildInfo"
		Short = Long
		return
	}
	var dirty string // "-dirty" suffix if dirty
	var commitHashAbbrev, commitDate string
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			GitCommit = s.Value
			if len(s.Value) >= 9 {
				commitHashAbbrev = s.Value[:9]
			}
		case "vcs.time":
			if len(s.Value) >= len("yyyy-mm-dd") {
				commitDate = s.Value[:len("yyyy-mm-dd")]
				commitDate = strings.ReplaceAll(commitDate, "-", "")
			}
		case "vcs.modified":
			if s.Value == "true" {
				dirty = "-dirty"
				GitDirty = true
			}
		}
	}

	// Backup path, using Go 1.18's built-in git stamping.
	Short = strings.TrimSpace(version) + "-dev" + commitDate
	Long = Short + "-t" + commitHashAbbrev + dirty
}

func Info(appName string) string {
	return fmt.Sprintf("%s version %s", appName, Long)
}
