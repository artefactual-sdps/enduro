#!/usr/bin/env python3

"""Patch the rendered Temporal schema job before apply.

We intentionally keep using the upstream repo chart instead of vendoring a
patched copy. For chart 1.0.0-rc.2, the temporal-schema job init containers
carry SQL_* env vars, but the temporal-sql-tool path we hit still reads
MYSQL_* variables such as MYSQL_PORT. We add compatibility aliases rather than
replacing the original variables so the chart output stays as close as possible
to upstream.

This targets the upstream schema job only. If a future upstream chart release
removes this mismatch, this post-renderer should be dropped.

Python is acceptable here because Tilt's helm_resource extension already
requires python3 for its own helper scripts, so this post-renderer is not an
extra runtime dependency in this workflow.

Refs:
https://github.com/temporalio/helm-charts/blob/main/charts/temporal/templates/server-job.yaml#L47-L49
https://github.com/temporalio/helm-charts/blob/main/charts/temporal/templates/_admintools-env.yaml#L30-L33
https://github.com/temporalio/temporal/blob/release/v1.30.x/temporal/environment/env.go#L26-L27
"""

import sys
import yaml


MYSQL_ENV_MAP = {
    "SQL_HOST": "MYSQL_SEEDS",
    "SQL_PORT": "MYSQL_PORT",
    "SQL_USER": "MYSQL_USER",
    "SQL_PASSWORD": "MYSQL_PWD",
    "SQL_DATABASE": "MYSQL_DATABASE",
}
def add_mysql_aliases(env):
    env_by_name = {item.get("name"): item for item in env if isinstance(item, dict)}

    for source_name, target_name in MYSQL_ENV_MAP.items():
        source = env_by_name.get(source_name)
        if target_name in env_by_name or not source:
            continue

        cloned = {"name": target_name}
        if "value" in source:
            cloned["value"] = source["value"]
        if "valueFrom" in source:
            cloned["valueFrom"] = source["valueFrom"]
        env.append(cloned)


docs = [doc for doc in yaml.safe_load_all(sys.stdin.read()) if doc is not None]
for doc in docs:
    if doc.get("kind") != "Job":
        continue
    if not doc.get("metadata", {}).get("name", "").startswith("temporal-schema-"):
        continue

    init_containers = doc.get("spec", {}).get("template", {}).get("spec", {}).get("initContainers", [])
    for container in init_containers:
        env = container.get("env", [])
        if env:
            add_mysql_aliases(env)

yaml.safe_dump_all(docs, sys.stdout, sort_keys=False, explicit_start=True)
