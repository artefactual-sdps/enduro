# Local/Development environment

## Requirements

### Standard installation

Enduro uses Tilt to set up a local environment building the Docker images in a
Kubernetes cluster. It has been tested with k3d, Minikube and Kind.

- [Docker] (v18.09+)
- [kubectl]
- [Tilt] (v0.22.2+)

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
- [Node.js and npm] (22+)
- GNU [Make] and [GCC]

If using Linux, Node.js binary distributions are available from [NodeSource].

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
| MinIO console | <http://localhost:7460> | `minio`     | `minio123`     |
| Temporal UI   | <http://localhost:7440> | `admin`     | `admin123`     |
| Grafana       | <http://localhost:7490> | `admin`     | `admin123`     |
| Keycloak      | <http://localhost:7470> | `keycloak`  | `keycloak123`  |

## Submit your first transfer

You're all set up! Go ahead and [submit your first transfer].

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
```

### TRIGGER_MODE_AUTO

Enables live updates on code changes for the enduro services.

### ENDURO_PRES_SYSTEM

Determines the preservation system between Archivematica (`am`) and a3m
(`a3m`), defaults to `a3m`. Check the [Working with Archivematica] docs if you
are planning to use Archivematica as preservation system.

### LOCAL_A3M

Build and use a local version of a3m. Requires to have the `a3m` repository
cloned as a sibling of this repository folder.

### PREPROCESSING_PATH

Relative path to a preprocessing child workflow repository. It loads a Tiltfile
called `Tiltfile.enduro` from that repository and mounts a presistent volume
claim (PVC) in the preservation system pod. That PVC must be defined in the
preprocessing and be called `preprocessing-pvc`. Check the [Preprocessing child
workflow] docs to configure the child workflow execution.

## Tilt UI helpers

### Upload to MinIO

In the Tilt UI header there is a cloud icon/button that allows you to configure
and trigger an upload to the `sips` bucket in MinIO. Click the caret to set the
path to a file/directory in the host and a MinIO object name, then click the
cloud icon to trigger the upload.

For example, to upload an existing sample ZIP from the project folder (make
sure you update `/path/to/enduro` to the proper project folder in the host):

- Host path: `/path/to/enduro/hack/sampledata/StructB-AM.zip`
- Object name: `StructB-AM.zip`

Alternatively, you can submit a transfer using the `/ingest/sip/upload` API via
cURL by running the following make target:

```bash
make upload-sample-transfer
```

### Flush

Also in the Tilt UI header, click the trash button to flush the existing data.
This will recreate the MySQL databases and the MinIO buckets, and restart the
required resources.

### Generators

Grouped as tools, there are some code generators:

- `gen-goa`: generates the Go API code based on the Goa design.
- `gen-dashboard-client`: generates the TypeScript client code for the API.
- `gen-ent`: generates the Go database code.

This resources need to be triggered manually by default, but they can be
configured to run automatically on code changes in the Tilt UI.

## Known issues

### MinIO uploads don't trigger workflows

The setup of the Minio buckets and the communication between Minio and Redis
is sometimes not setup properly. To solve it, from the Tilt UI, restart the
`minio` resource and then trigger the `minio-setup-buckets` resource.

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
[nodesource]: https://github.com/nodesource/distributions
[make]: https://www.gnu.org/software/make/
[gcc]: https://gcc.gnu.org/
[visual studio code]: https://code.visualstudio.com/
[working with archivematica]: archivematica.md
[preprocessing child workflow]: preprocessing.md
[submit your first transfer]: devel.md#upload-to-minio
[bine]: https://github.com/artefactual-labs/bine
