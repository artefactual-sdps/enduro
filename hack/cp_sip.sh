#!/usr/bin/env sh

set -eu

RESOURCE_NAME=$1
SIP_PATH=$2

# get k8s pod name from tilt resource name
POD_NAME="$(tilt get kubernetesdiscovery "$RESOURCE_NAME" -ojsonpath='{.status.pods[0].name}')"

# Copy the SIP to the Enduro watched directory
kubectl cp $SIP_PATH $POD_NAME:/home/enduro/sips/ --container=$RESOURCE_NAME
