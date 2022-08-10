version_settings(constraint=">=0.22.2")
load("ext://uibutton", "cmd_button", "text_input")

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
      trigger=["dashboard/package.json", "dashboard/package-lock.json"]
    ),
  ]
)

# All Kubernetes resources
k8s_yaml(kustomize("hack/kube/overlays/dev"))

# Enduro resources
k8s_resource("enduro", labels=["Enduro"])
k8s_resource("enduro-a3m", labels=["Enduro"])
k8s_resource("enduro-dashboard", port_forwards="3000", labels=["Enduro"])

# Other resources
k8s_resource("mysql", labels=["Others"])
k8s_resource(
  "minio",
  port_forwards=["7460:9001",
  "0.0.0.0:7461:9000"],
  labels=["Others"]
)
k8s_resource("opensearch", labels=["Others"])
k8s_resource(
  "opensearch-dashboards",
  port_forwards="7500:5601",
  labels=["Others"]
)
k8s_resource("redis", labels=["Others"])
k8s_resource("temporal", labels=["Others"])
k8s_resource("temporal-ui", port_forwards="7440:8080", labels=["Others"])

# Tools
k8s_resource("minio-setup-buckets", labels=["Tools"])
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
    "kubectl delete job --all -n sdps; \
    kubectl create -f hack/kube/tools/mysql-recreate-databases-job.yaml; \
    kubectl create -f hack/kube/tools/minio-recreate-buckets-job.yaml; \
    kubectl create -f hack/kube/tools/opensearch-delete-index-job.yaml; \
    kubectl wait --for=condition=complete --timeout=30s job --all -n sdps; \
    kubectl rollout restart deployment temporal -n sdps; \
    kubectl rollout restart deployment enduro -n sdps; \
    kubectl rollout restart statefulset enduro-a3m -n sdps;",
  ],
  location="nav",
  icon_name="delete",
  text="Flush",
)
