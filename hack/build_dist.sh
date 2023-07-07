#!/usr/bin/env sh
#
# Runs `go build` with flags configured for binary distribution. All
# it does differently from `go build` is burn git commit and version
# information into the binaries, so that we can track down user
# issues.

set -eu

MODULE_PATH="${MODULE_PATH:-github.com/artefactual-sdps/enduro}"

IFS=".$IFS" read -r major minor patch < internal/version/VERSION.txt
version_path=${MODULE_PATH}/internal/version
git_hash=$(git rev-parse HEAD)
if ! git diff-index --quiet HEAD; then
	git_hash="${git_hash}-dirty"
fi
short_hash=$(echo "$git_hash" | cut -c1-9)

long_suffix="-t$short_hash"
MINOR="$major.$minor"
SHORT="$MINOR.$patch"
LONG="${SHORT}$long_suffix"
GIT_HASH="$git_hash"
VERSION_PATH="$version_path"

if [ "$1" = "shellvars" ]; then
	cat <<EOF
VERSION_MINOR="$MINOR"
VERSION_SHORT="$SHORT"
VERSION_LONG="$LONG"
VERSION_GIT_HASH="$GIT_HASH"
VERSION_PATH="$VERSION_PATH"
EOF
	exit 0
fi

ldflags="-X ${VERSION_PATH}.Long=${LONG} -X ${VERSION_PATH}.Short=${SHORT} -X ${VERSION_PATH}.GitCommit=${GIT_HASH}"

exec go build -ldflags "$ldflags" "$@"
