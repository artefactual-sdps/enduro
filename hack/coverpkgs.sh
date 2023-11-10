#!/usr/bin/env sh

set -eu

curdir=$(cd "$(dirname "$0")" && pwd)

cd ${curdir}/..

#
# List all packages relevant to coverage reporting.
#
# Usage example:
#
#  $ go test -race -coverprofile=covreport -covermode=atomic -coverpkg=$(hack/coverpkgs.sh) -v ./...
#  $ go tool cover -func=html
#

go list ./... |
	grep -v "/artefactual-sdps/enduro/hack" |
	grep -v "/artefactual-sdps/enduro/internal/api/gen" |
	grep -v "/artefactual-sdps/enduro/internal/api/design" |
	grep -v "/fake" |
	paste -sd","
