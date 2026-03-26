version_settings(constraint=">=0.35.0")
secret_settings(disable_scrub=True)
ci_settings(timeout="10m", readiness_timeout="10m")
load("ext://uibutton", "cmd_button", "text_input")
load('ext://dotenv', 'dotenv')
load('ext://helm_resource', 'helm_resource', 'helm_repo')

# Load tilt env file if it exists
dotenv_path = ".tilt.env"
if os.path.exists(dotenv_path):
  dotenv(fn=dotenv_path)

# Get preservation system (default: 'am')
PRES_SYS = os.environ.get('ENDURO_PRES_SYSTEM', 'am')
if PRES_SYS not in ("a3m", "am"):
  fail("Invalid ENDURO_PRES_SYSTEM: {pres_sys}.".format(pres_sys=PRES_SYS))

true = ("true", "1", "yes", "t", "y")
LOCAL_A3M = os.environ.get("LOCAL_A3M", "").lower() in true

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

docker_build(
  "enduro-dashboard:dev",
  context="dashboard",
  # Comment the following line to serve the app with Nginx instead of the Vite
  # dev server
  target="builder",
  live_update=[
    fall_back_on("dashboard/vite.config.js"),
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
)

# Get kube overlay path
KUBE_OVERLAY = 'hack/kube/overlays/dev-a3m'
if PRES_SYS == 'am':
  KUBE_OVERLAY = 'hack/kube/overlays/dev-am'

# Load Kustomize YAML
yaml = kustomize(KUBE_OVERLAY)

# The CHILD_WORKFLOW_PATHS environment variable is a colon-separated list of 
# paths to child workflow directories. If set, we load each child workflow's
# Tiltfile.enduro to load resources required by the workflow (e.g. a Temporal 
# worker).
CHILD_WORKFLOW_PATHS = os.environ.get("CHILD_WORKFLOW_PATHS", "")
if CHILD_WORKFLOW_PATHS != "":
  for path in CHILD_WORKFLOW_PATHS.split(":"):
    # Load child workflow Tiltfile for Enduro
    load_dynamic(path.strip() + "/Tiltfile.enduro")

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

# Load Temporal Helm chart
helm_repo("temporal-helm", "https://go.temporal.io/helm-charts/")

# Use the upstream chart and patch the rendered schema.
# See hack/helm/temporal/post_renderer.py for details.
helm_resource(
  "temporal",
  "temporal-helm/temporal",
  namespace="enduro-sdps",
  flags=[
    "--version", "1.0.0-rc.2",
    "-f", "hack/helm/temporal/values.yaml",
    "--post-renderer", "./hack/helm/temporal/post_renderer.py",
  ],
  resource_deps=["temporal-helm", "mysql"],
  labels=["Others"],
  links=[link("http://localhost:7440", "Temporal UI")],
)

# Add the UI port-forward as a separate local resource. Attaching directly to
# the aggregated helm_resource is not reliable because that aggregates multiple
# workloads representing the whole chart. Run kubectl port-forward in a retry
# loop so it reconnects automatically after Temporal restarts.
local_resource(
  "temporal-web-port-forward",
  serve_cmd="""
    sh -c '
      while true; do
        kubectl wait --namespace enduro-sdps --for=condition=available --timeout=180s deployment/temporal-web || true
        kubectl port-forward --namespace enduro-sdps services/temporal-web 7440:8080 || true
        sleep 2
      done
    '
  """,
  resource_deps=["temporal"],
  allow_parallel=True,
)

# Configure trigger mode
trigger_mode = TRIGGER_MODE_MANUAL
if os.environ.get('TRIGGER_MODE_AUTO', '').lower() in true:
  trigger_mode = TRIGGER_MODE_AUTO

# Enduro resources
k8s_resource(
  "enduro",
  labels=["Enduro"],
  port_forwards=["9000:9000", "9002:9002"],
  trigger_mode=trigger_mode,
  resource_deps=["temporal"],
)
k8s_resource(
  "enduro-dashboard",
  labels=["Enduro"],
  port_forwards="8080:80",
  trigger_mode=trigger_mode
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
    resource_deps=["temporal", "ambox"],
  )
else:
  k8s_resource(
    "enduro-a3m",
    labels=["Enduro"],
    trigger_mode=trigger_mode,
    resource_deps=["temporal"],
  )

# Other resources
k8s_resource("keycloak", labels=["Others"], port_forwards="7470")
k8s_resource("mysql", labels=["Others"], port_forwards="3306")
k8s_resource(
  "minio",
  labels=["Others"],
  port_forwards=["7460:9001", "0.0.0.0:7461:9000"]
)
k8s_resource("redis", labels=["Others"])

# Tools
k8s_resource("minio-setup-buckets", labels=["Tools"], resource_deps=["minio"])
if PRES_SYS == 'am':
  k8s_resource(
    "mysql-create-amss-location",
    labels=["Tools"],
    resource_deps=["enduro"],
  )
else:
  k8s_resource(
    "mysql-create-locations",
    labels=["Tools"],
    resource_deps=["enduro"],
  )

# Observability (not in CI mode)
if config.tilt_subcommand != "ci":
  k8s_yaml(kustomize("hack/kube/overlays/observability"))
  k8s_resource("prometheus-server", labels=["Observability"], port_forwards="7491:9090")
  k8s_resource("grafana-alloy", labels=["Observability"])
  k8s_resource("grafana-tempo", labels=["Observability"])
  k8s_resource("grafana", labels=["Observability"], port_forwards="7490:3000")

# Buttons
cmd_button(
  "minio-upload",
  argv=[
    "sh",
    "-c",
    "docker run --rm \
      --add-host=host-gateway:host-gateway \
      --entrypoint=/bin/bash \
      -v $HOST_PATH:/sampledata/$OBJECT_NAME \
      minio/mc -c ' \
        mc alias set enduro http://host-gateway:7461 minio minio123; \
        mc cp -r /sampledata/$OBJECT_NAME enduro/sips/$OBJECT_NAME; \
      ' \
    ",
  ],
  location="nav",
  icon_name="cloud_upload",
  text="Minio upload",
  inputs=[
    text_input("HOST_PATH", label="Host path"),
    text_input("OBJECT_NAME", label="Object name"),
  ]
)
cmd_button(
  "flush",
  argv=[
    "sh",
    "-c",
    "kubectl config set-context --current --namespace enduro-sdps; \
    kubectl delete job --all; \
    kubectl create -f hack/kube/tools/mysql-recreate-databases-job.yaml; \
    kubectl create -f hack/kube/tools/minio-recreate-buckets-job.yaml; \
    kubectl wait --for=condition=complete --timeout=120s job --all; \
    kubectl rollout restart deployment temporal; \
    kubectl rollout restart deployment enduro; \
    kubectl rollout restart statefulset enduro-{pres_sys}; \
    kubectl create -f hack/kube/base/mysql-create-locations-job.yaml;".format(pres_sys=PRES_SYS),
  ],
  location="nav",
  icon_name="delete",
  text="Flush"
)
