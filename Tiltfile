version_settings(constraint=">=0.22.2")
secret_settings(disable_scrub=True)
load("ext://uibutton", "cmd_button", "text_input")
load('ext://dotenv', 'dotenv')

# Load tilt env file if it exists
dotenv_path = ".tilt.env"
if os.path.exists(dotenv_path):
  dotenv(fn=dotenv_path)

# Get preservation system (default: 'a3m')
PRES_SYS = os.environ.get('ENDURO_PRES_SYSTEM', 'a3m')
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

# Preprocessing
PREPROCESSING_PATH = os.environ.get("PREPROCESSING_PATH", "")
if PREPROCESSING_PATH != "":
  # Load preprocessing Tiltfile for Enduro
  load_dynamic(PREPROCESSING_PATH + "/Tiltfile.enduro")
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
k8s_resource("enduro", labels=["Enduro"], trigger_mode=trigger_mode)
k8s_resource("enduro-dashboard", port_forwards="8080:80", labels=["Enduro"], trigger_mode=trigger_mode)

if PRES_SYS == 'am':
  k8s_resource("enduro-am", labels=["Enduro"], trigger_mode=trigger_mode)
else:
  k8s_resource("enduro-a3m", labels=["Enduro"], trigger_mode=trigger_mode)

# Other resources
k8s_resource("keycloak", port_forwards="7470", labels=["Others"])
k8s_resource("mysql", port_forwards="3306", labels=["Others"])
k8s_resource(
  "minio",
  port_forwards=["7460:9001", "0.0.0.0:7461:9000"],
  labels=["Others"]
)
k8s_resource("redis", labels=["Others"])
k8s_resource("temporal", labels=["Others"])
k8s_resource("temporal-ui", port_forwards="7440:8080", labels=["Others"])

# Tools
k8s_resource("minio-setup-buckets", labels=["Tools"])
k8s_resource("mysql-create-locations", labels=["Tools"])
if PRES_SYS == 'am':
  k8s_resource("mysql-create-amss-location", labels=["Tools"])

# Observability
k8s_resource("grafana-agent", labels=["Observability"])
k8s_resource("grafana-tempo", labels=["Observability"])
k8s_resource("grafana", port_forwards="7490:3000", labels=["Observability"])

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
