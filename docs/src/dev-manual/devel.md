# Local/Development environment

## Requirements

### Standard installation

Enduro uses Tilt to set up a local environment building the Docker images in a
Kubernetes cluster. It has been tested with k3d, Minikube and Kind.

- [Docker] (v18.09+)
- [kubectl]
- [Tilt] (v0.35.0+)

A local Kubernetes cluster:

- [k3d] _(recomended, used in CI)_
- [Minikube] _(tested)_
- [Kind] _(tested)_

It can run with other solutions like Microk8s, Docker for Desktop/Mac, or
Lima/Colima (tested on macOS including Apple Silicon), and even against remote
clusters. Check Tilt's [Choosing a Local Dev Cluster] and [Install]
documentation for more information to install these requirements.

Additionally, follow the [Manage Docker as a non-root user] post-install guide
so that you don’t have to run Tilt with `sudo`. _Note that managing Docker as a
non-root user is **different** from running the docker daemon as a non-root user
(rootless)._

### Keycloak host

To make authentication work from the host browser and from within the cluster,
the following entry needs to be added to your `/etc/hosts` file:

```text
127.0.0.1 keycloak
```

For Windows/WSL2 users, open Notepad, as an Administrator, and then add the above
to your `etc/hosts` file located like `C:/Windows/System32/drivers/etc/hosts`.

## Requirements for development

While we run the services inside a Kubernetes cluster we recomend to install
Go, Node.js and other tools locally to ease the development process.

- [Go] (1.21+)
- [Node.js and npm] (see `/.node-version`)
- GNU [Make] and [GCC]

For the dashboard, use the Node version described in `/.node-version`. Minor or
patch drift within the supported major version is usually acceptable, but for
reproducibility we recommend using [nvm] to install and select the exact pinned
version. If using Linux, [NodeSource] is also available for installing Node.js
binaries system-wide, but it is less convenient when switching between project
specific versions.

## Editor

As source-code editor, we strongly recommended [Visual Studio Code] for its
great out-of-the-box support for Go and TypeScript. The project includes some
basic settings for formatting and we suggest installing the following VSCcode
extensions:

- Go
- Vue Language Features (Volar)
- TypeScript Vue Plugin (Volar)
- Prettier - Code formatter
- ESLint

## Managing development binaries with bine

This project uses [bine] to manage common development tools.

For example, to run the `atlas` tool, you can use:

```bash
go tool bine run atlas --help
```

If you want to run tools directly from your shell, update your `PATH`:

- Bash/Zsh: `source <(go tool bine env)`
- Fish: `go tool bine env | source`
- POSIX: `eval "$(go tool bine env)"`

See [bine](https://github.com/artefactual-labs/bine) for more details.

## Set up

Start a local Kubernetes cluster with a local registry. For example, with k3d:

```bash
k3d cluster create sdps-local --registry-create sdps-registry
```

Make sure kubectl is available and configured to use that cluster:

```bash
kubectl config view
```

Clone this repository and move into its folder if you have not done that
previously:

```bash
git clone git@github.com:artefactual-sdps/enduro.git
cd enduro
```

Bring up the environment:

```bash
tilt up
```

While the Docker images are built/downloaded and the Kubernetes resources are
created, hit `space` to open the Tilt UI in your browser. Check the [Tilt UI]
documentation to know more about it.

## Access

There are four services available from the host, three of them using SSO with
Keycloack:

| Service       | URL                     | Username    | Password       |
| ------------- | ----------------------- | ----------- | -------------- |
| Dashboard     | <http://localhost:8080> | `admin`     | `admin123`     |
| Temporal UI   | <http://localhost:7440> | `admin`     | `admin123`     |
| Grafana       | <http://localhost:7490> | `admin`     | `admin123`     |
| Keycloak      | <http://localhost:7470> | `keycloak`  | `keycloak123`  |

## Submit your first SIP

You're all set up! Go ahead and [submit your first SIP].

## Live updates

Tilt, by default, will watch for file changes in the project folder and it will
sync those changes, rebuild the Docker images and recreate the resources when
necessary. However, we have _disabled_ auto-load within the Tiltfile to reduce
the use of hardware resources. There are refresh buttons on each resource in the
Tilt UI that allow triggering manual updates and re-executing jobs and local
resources. You can also set the `trigger_mode` env string to `TRIGGER_MODE_AUTO`
within your local `.tilt.env` file to override this change and enable auto mode.

The `enduro-dashboard` uses Vite to serve the application in development
with hot reload. The `enduro` and `enduro-a3m-worker` services require rebuilding
the entire images - these will take longer to update.

## Stop/start the environment

Run `ctrl-c` on the terminal where `tilt up` is running and stop the cluster
with:

```bash
k3d cluster stop sdps-local
```

To start the environment again:

```bash
k3d cluster start sdps-local
tilt up
```

## Clear the cluster

> Check the Tilt UI helpers below to just flush the existing data.

To remove the resources created by Tilt in the cluster, execute:

```bash
tilt down
```

Note that it will take some time to delete the persistent volumes when you
run `tilt down` and flushing the existing data does not delete the cluster.
To delete the volumes immediately, you can delete the cluster.

## Delete the cluster

Deleting the cluster will remove all the resources immediatly, deleting
cluster container from the host. With k3d, run:

```bash
k3d cluster delete sdps-local
```

## Tilt environment configuration

A few configuration options can be changed by having a `.tilt.env` file
located in the root of the project. Example:

```text
TRIGGER_MODE_AUTO=true
ENDURO_PRES_SYSTEM=a3m
LOCAL_A3M=true
DASHBOARD_DEV=true
CHILD_WORKFLOW_PATHS='../preprocessing-acme:../acme-enduro-workflows'
MOUNT_PREPROCESSING_VOLUME=true
```

Tilt also renders the `enduro-config` Kubernetes secret from a TOML file in the
project root. If `enduro.local.toml` exists, Tilt uses it instead of
`enduro.toml`. When you change either file while Tilt is running, refresh the
Tiltfile and then refresh the `enduro` and active worker resource
(`enduro-am` or `enduro-a3m`) so the updated secret is applied in the cluster.

### TRIGGER_MODE_AUTO

Enables live updates on code changes for the enduro services.

### ENDURO_PRES_SYSTEM

Determines the preservation system between Archivematica (`am`) and a3m
(`a3m`), defaults to `am`. a3m is a lightweight Archivematica derivative,
but it has seen little adoption and is largely unmaintained. Check the
[Working with Archivematica] docs for more details about Archivematica.

### LOCAL_A3M

Build and use a local version of a3m. Requires to have the `a3m` repository
cloned as a sibling of this repository folder.

### DASHBOARD_DEV

If `DASHBOARD_DEV` is truthy (`t`, `true`, `y`, `yes`, `1`), Tilt builds the
dashboard image with the `builder` target and serves the dashboard with the
Vite development server, including hot reload. Otherwise, Tilt uses the default
image target and serves the dashboard with Nginx.

### CHILD_WORKFLOW_PATHS

A colon (:) separated list of relative paths to child workflow repositories. At
startup Tilt will attempt to load a `Tiltfile` file from each path which will
add any workflow specific resources to the Tilt environment (e.g. a child
worker). See the [Administrator configuration] docs for instructions on
configuring the child workflows.

### MOUNT_PREPROCESSING_VOLUME

If MOUNT_PREPROCESSING_VOLUME is truthy (t, true, y, yes, 1) Tilt mounts a
persistent volume claim (PVC) in the `enduro-am-worker` or `enduro-a3m-worker`
pod, depending on the preservation engine used. The PVC must be defined in the
preprocessing manifests and be called `preprocessing-pvc`.

## Tilt UI helpers

### Submit a SIP

In the Tilt UI header there is a cloud icon/button that uploads a SIP through
the internal API. Open the button menu to set a local ZIP path, or leave the
path empty to upload the default sample SIP.

Alternatively, you can submit a SIP using the `/ingest/sip/upload` API via cURL
by running the following make target:

```bash
make upload-sip
```

By default, this uploads `internal/testdata/zipped_transfer/small.zip`. To
upload a different ZIP file, pass its local path:

```bash
make upload-sip LOCAL_PATH=/path/to/sip.zip
```

### Flush

Also in the Tilt UI header, click the trash button to flush the existing data.
This will recreate the Enduro MySQL databases, and restart the required
resources. It does not recreate the Temporal databases or clean the
internal-storage volume; data in those locations remains after the flush.

## Interact with the internal-storage volume

The local development environment mounts the shared internal-storage PVC in the
`enduro` pod at `/home/enduro/internal-storage`. The volume contains internal
folders used by different local workflows:

- `ingest`: internal storage for the ingest domain.
- `storage`: internal storage for the storage domain.
- `sip-source`: filesystem-backed SIP source.
- `watched`: filesystem watched location.
- `watched-complete`: successfully ingested watched-location SIPs.
- `perma-aips`: local permanent AIP storage.

Use the `enduro` pod as an access point for copying, listing, and removing files
in any of those folders. Declare the following variables related to the Enduro
Kubernetes pod in the same BASH script as the subsequent `kubectl` commands:

```bash
NAMESPACE=enduro-sdps
CONTAINER=enduro
INTERNAL_STORAGE=/home/enduro/internal-storage
POD=$(kubectl -n "$NAMESPACE" get pod -l app=enduro \
  -o jsonpath='{.items[0].metadata.name}')
```

Copy a single local file ("./sip.zip") into the "sip-source" folder in the
Enduro pod:

```bash
kubectl -n "$NAMESPACE" cp -c "$CONTAINER" ./sip.zip \
  "$POD:$INTERNAL_STORAGE/sip-source/sip.zip"
```

Copy a single local file into the filesystem watched location under an ignored
temporary name, then rename it when the copy is complete. The default watcher
configuration ignores names ending in `.tmp`, so the watcher will not process
the file until it has its final name:

```bash
kubectl -n "$NAMESPACE" cp -c "$CONTAINER" ./sip.zip \
  "$POD:$INTERNAL_STORAGE/watched/sip.zip.tmp"

kubectl -n "$NAMESPACE" exec "$POD" -c "$CONTAINER" -- \
  mv "$INTERNAL_STORAGE/watched/sip.zip.tmp" \
    "$INTERNAL_STORAGE/watched/sip.zip"
```

Copy a full local directory ("./example-sip") into the filesystem watched
location under an ignored temporary directory, then rename it when the copy is
complete. Copy the SIP contents into the temporary directory so the final
renamed directory is the SIP root:

```bash
kubectl -n "$NAMESPACE" exec "$POD" -c "$CONTAINER" -- \
  mkdir -p "$INTERNAL_STORAGE/watched/example-sip.tmp"

tar -C ./example-sip -cf - . | \
  kubectl -n "$NAMESPACE" exec -i "$POD" -c "$CONTAINER" -- \
    tar -C "$INTERNAL_STORAGE/watched/example-sip.tmp" -xf -

kubectl -n "$NAMESPACE" exec "$POD" -c "$CONTAINER" -- \
  mv "$INTERNAL_STORAGE/watched/example-sip.tmp" \
    "$INTERNAL_STORAGE/watched/example-sip"
```

Copy the contents of a local directory into one of the internal folders without
creating an extra parent directory in the volume:

```bash
tar -C ./local-sips -cf - . | \
  kubectl -n "$NAMESPACE" exec -i "$POD" -c "$CONTAINER" -- \
    tar -C "$INTERNAL_STORAGE/sip-source" -xf -
```

List the internal-storage volume locations, or inspect a specific location:

```bash
kubectl -n "$NAMESPACE" exec "$POD" -c "$CONTAINER" -- \
  ls -la "$INTERNAL_STORAGE"

kubectl -n "$NAMESPACE" exec "$POD" -c "$CONTAINER" -- \
  ls -la "$INTERNAL_STORAGE/sip-source"

kubectl -n "$NAMESPACE" exec "$POD" -c "$CONTAINER" -- \
  ls -la "$INTERNAL_STORAGE/watched"

kubectl -n "$NAMESPACE" exec "$POD" -c "$CONTAINER" -- \
  ls -la "$INTERNAL_STORAGE/watched-complete"

kubectl -n "$NAMESPACE" exec "$POD" -c "$CONTAINER" -- \
  ls -la "$INTERNAL_STORAGE/perma-aips"
```

Copy a stored AIP from the internal-storage volume to your host. The source path
uses the storage object key. If you download the same AIP from the Enduro UI,
the downloaded file uses the full AIP name with its extension, not just the UUID;
with the default a3m configuration, that extension should be `.7z`.

```bash
kubectl -n "$NAMESPACE" cp -c "$CONTAINER" \
  "$POD:$INTERNAL_STORAGE/perma-aips/1a1453ed-ec55-4bf3-8900-37df6ee1634d" \
  "./1a1453ed-ec55-4bf3-8900-37df6ee1634d.7z"
```

Delete an individual file or folder from the volume:

```bash
kubectl -n "$NAMESPACE" exec "$POD" -c "$CONTAINER" -- \
  rm -f "$INTERNAL_STORAGE/sip-source/sip.zip"

kubectl -n "$NAMESPACE" exec "$POD" -c "$CONTAINER" -- \
  rm -rf "$INTERNAL_STORAGE/sip-source/example-sip"
```

Clean the watcher completed directory. This removes every completed SIP from
`watched-complete` while keeping the directory itself:

```bash
kubectl -n "$NAMESPACE" exec "$POD" -c "$CONTAINER" -- \
  find "$INTERNAL_STORAGE/watched-complete" -mindepth 1 -maxdepth 1 \
    -exec rm -rf {} +
```

[administrator configuration]: ../admin-manual/configuration.md
[docker]: https://docs.docker.com/get-docker/
[kubectl]: https://kubernetes.io/docs/tasks/tools/#kubectl
[tilt]: https://docs.tilt.dev/tutorial/1-prerequisites.html#install-tilt
[k3d]: https://k3d.io/v5.4.3/#installation
[minikube]: https://minikube.sigs.k8s.io/docs/start/
[kind]: https://kind.sigs.k8s.io/docs/user/quick-start#installation
[choosing a local dev cluster]: https://docs.tilt.dev/choosing_clusters.html
[install]: https://docs.tilt.dev/install.html
[manage docker as a non-root user]: https://docs.docker.com/engine/install/linux-postinstall/#manage-docker-as-a-non-root-user
[tilt ui]: https://docs.tilt.dev/tutorial/3-tilt-ui.html
[go]: https://go.dev/doc/install
[Node.js and npm]: https://nodejs.org/
[nvm]: https://github.com/nvm-sh/nvm
[nodesource]: https://github.com/nodesource/distributions
[make]: https://www.gnu.org/software/make/
[gcc]: https://gcc.gnu.org/
[visual studio code]: https://code.visualstudio.com/
[working with archivematica]: archivematica.md
[submit your first SIP]: devel.md#submit-a-sip
[bine]: https://github.com/artefactual-labs/bine
