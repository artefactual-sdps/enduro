package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

var (
	Version   = "(dev-version)"
	GitCommit = "(dev-commit)"
	BuildTime = "(dev-buildtime)"
	GoVersion = runtime.Version()
)

func Info(appName string) string {
	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, item := range info.Settings {
			if item.Key == "vcs.revision" {
				GitCommit = item.Value
			}
			if item.Key == "vcs.modified" && item.Value == "true" {
				GitCommit = "!" + GitCommit
			}
		}
	}
	return fmt.Sprintf("%s version %s (commit=%s) built on %s using %s",
		appName, Version, GitCommit, BuildTime, GoVersion)
}
