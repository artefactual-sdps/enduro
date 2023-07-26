#!/usr/bin/env bash

TMP_DIR=/tmp/inject_vite_envs
mkdir $TMP_DIR

# Copy original distribution files to dashboard file root
rm -rf $ENDURO_DASHBOARD_ROOT/*
cp -r $ENDURO_DASHBOARD_DIST/* $ENDURO_DASHBOARD_ROOT/

# Get a comma delimited list of env var names starting with "VITE"
VITE_ENVS=$(printenv | awk -F= '$1 ~ /^VITE/ {print $1}' | sed 's/^/\$/g' | paste -sd,);
echo "Vite envs: ${VITE_ENVS}"

# Inject environment variables into distribution files
for file in $ENDURO_DASHBOARD_ROOT/assets/*.js;
do
    echo "Inject VITE environment variables into $(basename $file)"
    envsubst $VITE_ENVS < $file > $TMP_DIR/$(basename $file)
    cp $TMP_DIR/$(basename $file) $file
done

rm -rf $TMP_DIR
