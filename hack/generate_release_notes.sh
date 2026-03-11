#!/usr/bin/env sh

set -eu

OUTPUT_PATH=${1}
CURRENT_TAG=${GITHUB_REF_NAME}

LATEST_HEADING=$(
	awk '
		/^## \[/ {
			if ($0 == "## [Unreleased]") {
				next
			}

			sub(/^## /, "", $0)
			print
			exit
		}
	' "CHANGELOG.md"
)

LATEST_VERSION=$(
	printf '%s\n' "$LATEST_HEADING" | sed -n 's/^\[\([^]]*\)\].*/\1/p'
)

if [ -z "$LATEST_VERSION" ]; then
	echo "Error: failed to extract version from heading: $LATEST_HEADING" >&2
	exit 1
fi

if [ "$LATEST_VERSION" != "${CURRENT_TAG#v}" ]; then
	echo "Error: latest changelog version $LATEST_VERSION does not match current tag $CURRENT_TAG" >&2
	exit 1
fi

PREVIOUS_TAG=$(
	git tag --sort=-version:refname | awk -v current="$CURRENT_TAG" '
		$0 == current {
			getline
			print
			exit
		}
	'
)

CHANGELOG_ANCHOR=$(
	printf '%s' "$LATEST_HEADING" |
		tr '[:upper:]' '[:lower:]' |
		sed 's/[^a-z0-9 -]//g; s/ /-/g'
)

cat > "${OUTPUT_PATH}" << EOF
- Functional changelog: [CHANGELOG.md](https://github.com/artefactual-sdps/enduro/blob/main/CHANGELOG.md#${CHANGELOG_ANCHOR})
- Full changelog: https://github.com/artefactual-sdps/enduro/compare/${PREVIOUS_TAG}...${CURRENT_TAG}
EOF
