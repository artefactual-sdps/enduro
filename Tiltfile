version_settings(constraint=">=0.22.2")
load("ext://uibutton", "cmd_button", "text_input")

def dotenv(fn):
  """Read environment strings from a file."""
  f = read_file(fn, default="")
  lines = str(f).splitlines()
  for line in lines:
    v = line.split('=', 1)
    if len(v) < 2:
      continue
    os.putenv(v[0], v[1])

# Docker images
docker_build("enduro:dev", context=".")
docker_build(
  "enduro-a3m-worker:dev",
  context=".",
  target="enduro-a3m-worker"
)
docker_build(
  "enduro-dashboard:dev",
  context="dashboard",
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

# All Kubernetes resources
k8s_yaml(kustomize("hack/kube/overlays/dev"))

# Configure trigger mode
dotenv(fn=".tilt.env")
trigger_mode = TRIGGER_MODE_AUTO
if os.environ.get('TRIGGER_MODE_MANUAL', ''):
  trigger_mode = TRIGGER_MODE_MANUAL

# Enduro resources
k8s_resource("enduro", labels=["Enduro"], trigger_mode=trigger_mode)
k8s_resource("enduro-a3m", labels=["Enduro"], trigger_mode=trigger_mode)
k8s_resource("enduro-internal", port_forwards="9000", labels=["Enduro"], trigger_mode=trigger_mode)
k8s_resource("enduro-dashboard", port_forwards="3000", labels=["Enduro"], trigger_mode=trigger_mode)

# Other resources
k8s_resource("dex", port_forwards="5556", labels=["Others"])
k8s_resource("ldap", labels=["Others"])
k8s_resource("mysql", port_forwards="3306", labels=["Others"])
k8s_resource(
  "minio",
  port_forwards=["7460:9001",
  "0.0.0.0:7461:9000"],
  labels=["Others"]
)
k8s_resource("redis", labels=["Others"])
k8s_resource("temporal", labels=["Others"])
k8s_resource("temporal-ui", port_forwards="7440:8080", labels=["Others"])

# Tools
k8s_resource("minio-setup-buckets", labels=["Tools"])
k8s_resource("mysql-create-locations", labels=["Tools"])
local_resource(
  "gen-goa",
  cmd="make gen-goa",
  auto_init=False,
  trigger_mode=TRIGGER_MODE_MANUAL,
  deps=["internal/api"],
  ignore=["internal/api/gen"],
  labels=["Tools"]
)
local_resource(
  "gen-dashboard-client",
  cmd="make gen-dashboard-client",
  auto_init=False,
  trigger_mode=TRIGGER_MODE_MANUAL,
  deps=["internal/api/gen"],
  labels=["Tools"]
)
local_resource(
  "gen-ent",
  cmd="make gen-ent",
  auto_init=False,
  trigger_mode=TRIGGER_MODE_MANUAL,
  deps=["internal/storage/persistence/ent/schema"],
  labels=["Tools"]
)

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
    "kubectl config set-context --current --namespace sdps; \
    kubectl delete job --all; \
    kubectl create -f hack/kube/tools/mysql-recreate-databases-job.yaml; \
    kubectl create -f hack/kube/tools/minio-recreate-buckets-job.yaml; \
    kubectl wait --for=condition=complete --timeout=120s job --all; \
    kubectl rollout restart deployment temporal; \
    kubectl rollout restart deployment enduro; \
    kubectl rollout restart statefulset enduro-a3m; \
    kubectl rollout restart deployment dex; \
    kubectl create -f hack/kube/base/mysql-create-locations-job.yaml;",
  ],
  location="nav",
  icon_name="delete",
  text="Flush"
)
