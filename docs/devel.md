## Local/Development environment

Enduro uses Tilt to set up a local environment building the Docker images in a
Kubernetes cluster. It has been tested with k3d, Minikube and Kind.

### Requirements

- [Docker] (v18.09+)
- [kubectl]
- [Tilt] (v0.22.2+)

A local Kubernetes cluster:

- [k3d] _(recomended, used in CI)_
- [Minikube] _(tested)_
- [Kind] _(tested)_

It can run with other solutions like Microk8s or Docker for Desktop/Mac and
even against remote clusters, check Tilt's [Choosing a Local Dev Cluster] and
[Install] documentation for more information to install these requirements.

Additionally, follow the [Manage Docker as a non-root user] post-install guide
so that you don’t have to run Tilt with `sudo`. *Note that managing Docker as a
non-root user is **different** from running the docker daemon as a non-root user
(rootless).*

#### Dex host

To make authentication work from the host browser and from within the cluster,
the following entry needs to be added to your `/etc/hosts` file:

```
127.0.0.1 dex
```
For Windows/WSL2 users, open Notepad, as an Administrator, and then add the above
to your `etc/hosts` file located like `C:/Windows/System32/drivers/etc/hosts`.

### Requirements for development

While we run the services inside a Kubernetes cluster we recomend to install
Go, Node and other tools locally to ease the development process.

- [Go] (1.21+)
- [NPM and Node] (18+)
- GNU [Make] and [GCC]

If using Linux, Node.js binary distributions are available from [NodeSource].

#### Go tools

We use [bingo] to manage some Go tools and binaries needed to perform various
development operations.

bingo builds pinned tools in your `$GOBIN` path. If `$GOBIN` is undefined, we
try to set its value by expanding `$(go env GOPATH)/bin` since it is common for
Go developers to have previously defined `$GOPATH`.

Preferably, define `$GOBIN` in your environment and include the same directory
in your `$PATH` so your system knows where to find the executables, e.g.:

```
export GOBIN=$HOME/go/bin
export PATH=$HOME/go/bin:$PATH
```

We recommend to [set the environment strings permanently] - follow the link to
know more.

Then, clone this repository and install those tools:

```
git clone git@github.com:artefactual-sdps/enduro.git
cd enduro
make tools
```

These tools will be used through Makefile rules and the Tilt UI.

### Editor

As source-code editor, we strongly recommended [Visual Studio Code] for its
great out-of-the-box support for Go and TypeScript. The project includes some
basic settings for formatting and we suggest installing the following VSCcode
extensions:

- Go
- Vue Language Features (Volar)
- TypeScript Vue Plugin (Volar)
- Prettier - Code formatter
- ESLint

### Set up

Start a local Kubernetes cluster with a local registry. For example, with k3d:

```
k3d cluster create sdps-local --registry-create sdps-registry
```

Make sure kubectl is available and configured to use that cluster:

```
kubectl config view
```

Clone this repository and move into its folder if you have not done that
previously:

```
git clone git@github.com:artefactual-sdps/enduro.git
cd enduro
```

Bring up the environment:

```
tilt up
```

While the Docker images are built/downloaded and the Kubernetes resources are
created, hit `space` to open the Tilt UI in your browser. Check the [Tilt UI]
documentation to know more about it.

### Access

There are four services available from the host:

- Enduro dashboard: http://localhost:8080
- Minio console: http://localhost:7460 (username: minio, password: minio123)
- Temporal UI: http://localhost:7440

### Live updates

Tilt, by default, will watch for file changes in the project folder and it will
sync those changes, rebuild the Docker images and recreate the resources when
necessary. However, we have *disabled* auto-load within the Tiltfile to reduce
the use of hardware resources. There are refresh buttons on each resource in the
Tilt UI that allow triggering manual updates and re-executing jobs and local
resources. You can also set the `trigger_mode` env string to `TRIGGER_MODE_AUTO`
within your local `.tilt.env` file to override this change and enable auto mode.

The `enduro-dashboard` uses Vite to serve the application in development
with hot reload. The `enduro` and `enduro-a3m-worker` services require rebuilding
the entire images - these will take longer to update.

### Stop/start the environment

Run `ctrl-c` on the terminal where `tilt up` is running and stop the cluster
with:

```
k3d cluster stop sdps-local
```

To start the environment again:

```
k3d cluster start sdps-local
tilt up
```

### Clear the cluster

> Check the Tilt UI helpers below to just flush the existing data.

To remove the resources created by Tilt in the cluster, execute:

```
tilt down
```
Note that it will take some time to delete the persistent volumes when you
run `tilt down` and flushing the existing data does not delete the cluster.
To delete the volumes immediately, you can delete the cluster.

### Delete the cluster

Deleting the cluster will remove all the resources immediatly, deleting
cluster container from the host. With k3d, run:

```
k3d cluster delete sdps-local
```

### Tilt UI helpers

#### Upload to Minio

In the Tilt UI header there is a cloud icon/button that allows you to configure
and trigger an upload to the `sips` bucket in Minio. Click the caret to set the
path to a file/directory in the host and a Minio object name, then click the
cloud icon to trigger the upload.

For example, to upload an existing sample ZIP from the project folder (make
sure you update `/path/to/enduro` to the proper project folder in the host):

- Host path: `/path/to/enduro/hack/sampledata/StructB-AM.zip`
- Object name: `StructB-AM.zip`

Alternatively, you can use the Enduro API to upload the file like in the
following example:

```
curl \
  -F "file=@/path/to/enduro/hack/sampledata/StructB-AM.zip" \
  http://localhost:9000/upload/upload
```

#### Flush

Also in the Tilt UI header, click the trash button to flush the existing data.
This will recreate the MySQL databases and the MinIO buckets, and restart the
required resources.

#### Generators

Grouped as tools, there are some code generators:

- `gen-goa`: generates the Go API code based on the Goa design.
- `gen-dashboard-client`: generates the TypeScript client code for the API.
- `gen-ent`: generates the Go database code.

This resources need to be triggered manually by default, but they can be
configured to run automatically on code changes in the Tilt UI.

### Known issues

#### Minio uploads don't trigger workflows

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
[npm and node]: https://nodejs.org/
[nodesource]: https://github.com/nodesource/distributions
[make]: https://www.gnu.org/software/make/
[gcc]: https://gcc.gnu.org/
[bingo]: https://github.com/bwplotka/bingo
[visual studio code]: https://code.visualstudio.com/
[set the environment strings permanently]: https://unix.stackexchange.com/a/117470
