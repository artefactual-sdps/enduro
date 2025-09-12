#!/usr/bin/env bash

TMP_DIR=/tmp/inject_vite_envs
mkdir $TMP_DIR

# Copy original distribution files to dashboard file root
rm -rf $ENDURO_DASHBOARD_ROOT/*
cp -r $ENDURO_DASHBOARD_DIST/* $ENDURO_DASHBOARD_ROOT/

# Get a comma delimited list of env var names starting with "VITE"
VITE_ENVS=$(printenv | awk -F= '$1 ~ /^VITE/ {print $1}' | sed 's/^/\$/g' | paste -sd,);
echo "Vite envs: ${VITE_ENVS}"

# Inject environment variables into distribution files and remove
# placeholders that were not replaced (env. vars. not set).
for file in $ENDURO_DASHBOARD_ROOT/assets/*.js;
do
    echo "Inject VITE environment variables into assets/$(basename $file)"
    envsubst $VITE_ENVS < $file > $TMP_DIR/$(basename $file)
    sed -E -i 's/\$VITE_[A-Z0-9_]+//g' $TMP_DIR/$(basename $file)
    cp $TMP_DIR/$(basename $file) $file
done
echo "Inject VITE environment variables into index.html"
envsubst $VITE_ENVS < $ENDURO_DASHBOARD_ROOT/index.html > $TMP_DIR/index.html
sed -E -i 's/\$VITE_[A-Z0-9_]+//g' $TMP_DIR/index.html
cp $TMP_DIR/index.html $ENDURO_DASHBOARD_ROOT/index.html

rm -rf $TMP_DIR
