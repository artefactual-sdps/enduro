#!/usr/bin/env sh

set -eu

TARGET=${1:-}
if [ -z "$TARGET" ]; then
	echo "Service name must be supplied, e.g.:"
	echo "\t $ `basename "$0"` enduro"
	exit 1
fi
case "$TARGET" in
	"enduro")
		IMAGE_NAME="enduro"
		TARGET="enduro"
		;;
	"enduro-a3m-worker")
		IMAGE_NAME="enduro-a3m-worker"
		TARGET="enduro-a3m-worker"
		;;
	*)
		echo "Accepted values: enduro, enduro-a3m-worker.";
		exit 1;
		;;
esac

eval $(./hack/build_dist.sh shellvars)

DEFAULT_IMAGE_NAME="${IMAGE_NAME}:${VERSION_SHORT}"
TILT_EXPECTED_REF=${EXPECTED_REF:-}
IMAGE_NAME="${TILT_EXPECTED_REF:-$DEFAULT_IMAGE_NAME}"
BUILD_OPTS="${BUILD_OPTS:-}"

env DOCKER_BUILDKIT=1 docker build \
	-t "$IMAGE_NAME" \
	-f "Dockerfile" \
	--build-arg="TARGET=$TARGET" \
	--build-arg="VERSION_PATH=$VERSION_PATH" \
	--build-arg="VERSION_LONG=$VERSION_LONG" \
	--build-arg="VERSION_SHORT=$VERSION_SHORT" \
	--build-arg="VERSION_GIT_HASH=$VERSION_GIT_HASH" \
	$BUILD_OPTS \
		.