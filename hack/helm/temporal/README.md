# Temporal Helm chart rendering

The rendered Temporal Kubernetes manifest is included at
[`hack/kube/base/temporal.yaml`](../../kube/base/temporal.yaml) and loaded
by kustomize during `tilt up`. The manifest comes from the `temporal/temporal`
chart, rendered with [`values.yaml`](./values.yaml).

## Regenerate the manifest

Run this from the repo root:

```bash
helm repo add temporal https://go.temporal.io/helm-charts/
helm template temporal temporal/temporal \
  --skip-tests \
  --version 1.0.0 \
  --namespace enduro-sdps \
  -f hack/helm/temporal/values.yaml \
  > hack/kube/base/temporal.yaml
```

## Manual edits after rendering

The rendered output is not applied verbatim. After regenerating
[`hack/kube/base/temporal.yaml`](../../kube/base/temporal.yaml), keep the
manual `MYSQL_*` alias blocks in the `temporal-schema-1` job. The upstream
chart renders `SQL_*`, but the Temporal schema tool path used here still looks
for `MYSQL_*` variables such as `MYSQL_PORT`. This patch applies to the
`create-default-store`, `manage-schema-default-store`,
`create-visibility-store`, and `manage-schema-visibility-store` init
containers. Each of those containers needs `MYSQL_SEEDS`, `MYSQL_PORT`,
`MYSQL_DATABASE`, `MYSQL_USER`, and `MYSQL_PWD` copied from the corresponding
`SQL_*` values.
