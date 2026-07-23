version_settings(constraint=">=0.35.0")
secret_settings(disable_scrub=True)
ci_settings(
  timeout="10m",
  readiness_timeout="10m",
  k8s_grace_period="2m",
)
load("ext://uibutton", "cmd_button", "text_input")
load('ext://dotenv', 'dotenv')

# Load tilt env file if it exists
dotenv_path = ".tilt.env"
if os.path.exists(dotenv_path):
  dotenv(fn=dotenv_path)

# Get preservation system (default: 'am')
PRES_SYS = os.environ.get("ENDURO_PRES_SYSTEM", "am")
if PRES_SYS not in ("a3m", "am"):
  fail("Invalid ENDURO_PRES_SYSTEM: {pres_sys}.".format(pres_sys=PRES_SYS))

OBJECT_STORE = os.environ.get("OBJECT_STORE", "filesystem")
if OBJECT_STORE not in ("filesystem", "seaweedfs"):
  fail("Invalid OBJECT_STORE: {object_store}.".format(object_store=OBJECT_STORE))

true = ("true", "1", "yes", "t", "y")
LOCAL_A3M = os.environ.get("LOCAL_A3M", "").lower() in true
DASHBOARD_DEV = os.environ.get("DASHBOARD_DEV", "").lower() in true

LOCATION_JOB_PATH = "hack/kube/overlays/dev-{pres_sys}/mysql-create-{pres_sys}-location-job.yaml".format(
  pres_sys=PRES_SYS,
)
if PRES_SYS == "a3m" and OBJECT_STORE == "seaweedfs":
  LOCATION_JOB_PATH = "hack/kube/overlays/dev-a3m-seaweedfs/mysql-create-a3m-location-job.yaml"

def add_enduro_config_secret(yaml):
  config_path = "enduro.toml"
  if os.path.exists("enduro.local.toml"):
    config_path = "enduro.local.toml"

  config = str(read_file(config_path))
  secret_yaml = encode_yaml({
    "apiVersion": "v1",
    "kind": "Secret",
    "metadata": {"name": "enduro-config", "namespace": "enduro-sdps"},
    "type": "Opaque",
    "stringData": {"enduro.toml": config},
  })

  return [yaml, secret_yaml]

# Docker images
custom_build(
  ref="enduro:dev",
  command=["hack/build_docker.sh", "enduro"],
  deps=["."],
)

if PRES_SYS == 'am':
  custom_build(
    ref="enduro-am-worker:dev",
    command=["hack/build_docker.sh", "enduro-am-worker"],
    deps=["."],
  )
else:
  custom_build(
    ref="enduro-a3m-worker:dev",
    command=["hack/build_docker.sh", "enduro-a3m-worker"],
    deps=["."],
  )
  if LOCAL_A3M:
    docker_build("ghcr.io/artefactual-labs/a3m", context="../a3m")

dashboard_live_update = []
if DASHBOARD_DEV:
  dashboard_live_update = [
    fall_back_on("dashboard/vite.config.ts"),
    sync("dashboard/", "/app/"),
    run(
      "npm set cache /app/.npm && npm install-clean",
      trigger=[
        "dashboard/package.json",
        "dashboard/package-lock.json",
        "dashboard/.env*",
      ]
    ),
  ]

docker_build(
  "enduro-dashboard:dev",
  context="dashboard",
  target="builder" if DASHBOARD_DEV else "",
  live_update=dashboard_live_update,
)

# Get kube overlay path
KUBE_OVERLAY = 'hack/kube/overlays/dev-a3m'
if PRES_SYS == 'am':
  KUBE_OVERLAY = 'hack/kube/overlays/dev-am'
if OBJECT_STORE == "seaweedfs":
  KUBE_OVERLAY = "{overlay}-seaweedfs".format(overlay=KUBE_OVERLAY)

# Load Kustomize YAML
yaml = kustomize(KUBE_OVERLAY)
yaml = add_enduro_config_secret(yaml)

# The CHILD_WORKFLOW_PATHS environment variable is a colon-separated list of
# paths to child workflow directories. If set, we load each child workflow's
# Tiltfile to load resources required by the workflow (e.g. a Temporal worker).
CHILD_WORKFLOW_PATHS = os.environ.get("CHILD_WORKFLOW_PATHS", "")
if CHILD_WORKFLOW_PATHS != "":
  for path in CHILD_WORKFLOW_PATHS.split(":"):
    # Load child workflow Tiltfile for Enduro
    load_dynamic(path.strip() + "/Tiltfile")

# The preprocessing child workflow requires extra setup for a shared directory
MOUNT_PREPROCESSING_VOLUME = os.environ.get("MOUNT_PREPROCESSING_VOLUME", "")
if MOUNT_PREPROCESSING_VOLUME in true:
  # Get Enduro a3m/am worker k8s manifest
  if PRES_SYS == "a3m":
    pres_yaml, yaml = filter_yaml(yaml, name="^enduro-a3m$", kind="StatefulSet")
  else:
    pres_yaml, yaml = filter_yaml(yaml, name="^enduro-am$", kind="Deployment")
  # Append preprocessing volume and volume mount to worker container,
  # this will only work in single node k8s cluster deployments
  volume = {"name": "shared-dir", "persistentVolumeClaim": {"claimName": "preprocessing-pvc"}}
  volume_mount = {"name": "shared-dir", "mountPath": "/home/enduro/preprocessing"}
  pres_obj = decode_yaml(pres_yaml)
  if "volumes" not in pres_obj["spec"]["template"]["spec"]:
    pres_obj["spec"]["template"]["spec"]["volumes"] = []
  pres_obj["spec"]["template"]["spec"]["volumes"].append(volume)
  for container in pres_obj["spec"]["template"]["spec"]["containers"]:
    if container["name"] in ["enduro-a3m-worker", "enduro-am-worker"]:
      container["volumeMounts"].append(volume_mount)
  pres_yaml = encode_yaml(pres_obj)
  yaml = [yaml, pres_yaml]

# Load Kubernetes resources
k8s_yaml(yaml)

# Configure trigger mode
trigger_mode = TRIGGER_MODE_MANUAL
if os.environ.get('TRIGGER_MODE_AUTO', '').lower() in true:
  trigger_mode = TRIGGER_MODE_AUTO

# Enduro resources
enduro_resource_deps = ["temporal-schema-1-2-0-1"]
if OBJECT_STORE == "seaweedfs":
  enduro_resource_deps.append("seaweedfs")

k8s_resource(
  "enduro",
  labels=["Enduro"],
  port_forwards=["9000:9000", "9002:9002"],
  trigger_mode=trigger_mode,
  resource_deps=enduro_resource_deps,
)
k8s_resource(
  "enduro-dashboard",
  labels=["Enduro"],
  port_forwards="8080:80",
  trigger_mode=trigger_mode,
  resource_deps=["enduro"],
)

if PRES_SYS == 'am':
  k8s_resource(
    "ambox",
    labels=["Others"],
    port_forwards=["64080:64080", "64081:64081"],
  )
  k8s_resource(
    "enduro-am",
    labels=["Enduro"],
    trigger_mode=trigger_mode,
    resource_deps=enduro_resource_deps + ["ambox"],
  )
else:
  k8s_resource(
    "enduro-a3m",
    labels=["Enduro"],
    trigger_mode=trigger_mode,
    resource_deps=enduro_resource_deps,
  )

# Temporal resources
k8s_resource(
  "temporal-schema-1-2-0-1",
  labels=["Temporal"],
  resource_deps=["mysql"],
)
k8s_resource(
  "temporal-admintools",
  labels=["Temporal"],
)
k8s_resource(
  "temporal-frontend",
  labels=["Temporal"],
  resource_deps=["mysql"],
)
k8s_resource(
  "temporal-history",
  labels=["Temporal"],
  resource_deps=["temporal-schema-1-2-0-1"],
)
k8s_resource(
  "temporal-matching",
  labels=["Temporal"],
  resource_deps=["temporal-schema-1-2-0-1"],
)
k8s_resource(
  "temporal-namespace-1-2-0-1",
  labels=["Temporal"],
  resource_deps=["temporal-frontend"],
)
k8s_resource(
  "temporal-worker",
  labels=["Temporal"],
  resource_deps=["temporal-schema-1-2-0-1"],
)
k8s_resource(
  "temporal-web",
  labels=["Temporal"],
  port_forwards=["7440:8080"],
  resource_deps=["temporal-schema-1-2-0-1", "keycloak"],
)

# Other resources
k8s_resource("keycloak", labels=["Others"], port_forwards="7470")
k8s_resource("mysql", labels=["Others"], port_forwards="3306")
k8s_resource("redis", labels=["Others"])
if OBJECT_STORE == "seaweedfs":
  k8s_resource(
    "seaweedfs",
    labels=["Others"],
    port_forwards=["23646:23646"],
  )

# Tools
if PRES_SYS == 'am':
  k8s_resource(
    "mysql-create-am-location",
    labels=["Tools"],
    resource_deps=["enduro"],
  )
else:
  k8s_resource(
    "mysql-create-a3m-location",
    labels=["Tools"],
    resource_deps=["enduro"],
  )

# Observability (not in CI mode)
if config.tilt_subcommand != "ci":
  k8s_yaml(kustomize("hack/kube/overlays/observability"))
  k8s_resource("lgtm", labels=["Observability"], port_forwards=["7490:3000", "7491:9090"])
  k8s_resource("alloy", labels=["Observability"], resource_deps=["lgtm"])

# Buttons
cmd_button(
  "upload-sip",
  argv=[
    "sh",
    "-c",
    'if [ -n "$LOCAL_PATH" ]; then make upload-sip LOCAL_PATH="$LOCAL_PATH"; else unset LOCAL_PATH; make upload-sip; fi',
  ],
  location="nav",
  icon_name="cloud_upload",
  text="Upload SIP",
  inputs=[
    text_input("LOCAL_PATH", label="Local path"),
  ],
)
cmd_button(
  "flush",
  argv=[
    "sh",
    "-c",
    "kubectl config set-context --current --namespace enduro-sdps; \
    kubectl delete job --all; \
    kubectl create -f hack/kube/tools/mysql-recreate-databases-job.yaml; \
    kubectl wait --for=condition=complete --timeout=120s job --all; \
    kubectl rollout restart deployment enduro; \
    kubectl rollout restart {kind} enduro-{pres_sys}; \
    kubectl create -f {location_job_path};".format(
      pres_sys=PRES_SYS,
      kind="statefulset" if PRES_SYS == "a3m" else "deployment",
      location_job_path=LOCATION_JOB_PATH,
    ),
  ],
  location="nav",
  icon_name="delete",
  text="Flush"
)
