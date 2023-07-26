#!/usr/bin/env bash

TMP_DIR=/tmp/inject_vite_envs
ASSETS_DIR=$ENDURO_DASHBOARD_ROOT/assets

mkdir $TMP_DIR
cp $ASSETS_DIR/*.js $TMP_DIR

# Get a comma delimited list of env var names starting with "VITE"
VITE_ENVS=$(printenv | awk -F= '$1 ~ /^VITE/ {print $1}' | sed 's/^/\$/g' | paste -sd,);
echo "Vite envs: ${VITE_ENVS}"

for file in $TMP_DIR/*.js;
do
    echo "Inject VITE environment variables into $(basename $file)"
    envsubst $VITE_ENVS < $file > $ASSETS_DIR/$(basename $file)
done

rm -rf $TMP_DIR
