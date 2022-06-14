version_settings(constraint=">=0.22.2")

load("ext://uibutton", "cmd_button", "text_input")

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

k8s_yaml(kustomize("hack/kube/overlays/dev"))

k8s_resource("enduro-dashboard", port_forwards="3000")

k8s_resource("minio", port_forwards=["7460:9001", "0.0.0.0:7461:9000"])

k8s_resource("opensearch-dashboards", port_forwards="7500:5601")

k8s_resource("temporal-ui", port_forwards="7440:8080")

local_resource(
  "gen-api",
  cmd="docker run \
    --rm \
    --user $(id -u):$(id -g) \
    --volume $(pwd):/src \
    fixl/goagen:3.7.5 gen \
      github.com/artefactual-labs/enduro/internal/api/design \
      -o internal/api \
  ",
  deps=["internal/api"],
  ignore=["internal/api/gen"],
  auto_init=False
)

local_resource(
  "gen-dashboard-client",
  cmd="docker run \
    --rm \
    --user $(id -u):$(id -g) \
    --volume $(pwd):/local \
    openapitools/openapi-generator-cli:v6.0.0 generate \
      --input-spec /local/internal/api/gen/http/openapi.json \
      --generator-name typescript-fetch \
      --output /local/dashboard/src/openapi-generator/ \
      --skip-validate-spec \
      -p 'generateAliasAsModel=false' \
      -p 'withInterfaces=true' \
      -p 'supportsES6=true' \
  ",
  deps=["internal/api/gen"],
  auto_init=False
)

cmd_button(
  "minio-upload",
  argv=[
    "sh",
    "-c",
    "docker run \
      --rm \
      --add-host=host-gateway:host-gateway \
      --entrypoint=/bin/bash \
      --volume $HOST_PATH:/sampledata/$OBJECT_NAME \
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
