<p align="left">
  <a href="https://github.com/artefactual-sdps/enduro/releases/latest"><img src="https://img.shields.io/github/v/release/artefactual-sdps/enduro.svg?color=orange"/></a>
  <img src="https://github.com/artefactual-sdps/enduro/workflows/Test/badge.svg"/>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg"/></a>
  <a href="https://goreportcard.com/report/github.com/artefactual-sdps/enduro"><img src="https://goreportcard.com/badge/github.com/artefactual-sdps/enduro"/></a>
  <a href="https://codecov.io/gh/artefactual-sdps/enduro"><img src="https://img.shields.io/codecov/c/github/artefactual-sdps/enduro"/></a>
</p>

# Enduro

Enduro is a tool designed to automate the processing of transfers in multiple
Archivematica pipelines. It's part of a preservation solution that is being
used by the [National Health Archive (NHA)] and the [National Center for Truth
and Reconciliation (NCTR)].

It's a **proof of concept** at its very early stages. It aims to cover our
client's needs while exploring new and innovative ways to build durable and
fault-tolerant workflows suited for preservation.

## Further reading

Visit https://enduroproject.netlify.com for more details.

## Local environment

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
so that you don’t have to run Tilt with `sudo`.

### Set up

Start a local Kubernetes cluster with a local registry. For example, with k3d:

```
k3d cluster create sdps-local --registry-create sdps-registry
```

Make sure kubectl is available and configured to use that cluster:

```
kubectl config view
```

Clone this repository and move into its folder:

```
git clone https://github.com/artefactual-sdps/enduro.git
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

- Enduro dashboard: http://localhost:3000
- Minio console: http://localhost:7460 (username: minio, password: minio123)
- Temporal UI: http://localhost:7440
- Opensearch dashboards: http://localhost:7500

### Live updates

Tilt will watch for file changes in the project folder and it will sync those
changes, rebuild the Docker images and recreate the resources when necessary.
The `enduro-dashboard` uses Vite to serve the application in development with
hot reload. The `enduro` and `enduro-a3m-worker` services require rebuilding
the entire images - these will take longer to update.

Additionally, there are refresh buttons on each resource in the Tilt UI that
allow triggering manual updates and re-executing jobs and local resources.

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

However, that won't remove all the persistent volumes at the moment, to do so
run:

```
kubectl delete pvc,pv --all
```

### Delete the cluster

Deleting the cluster will remove all the resources from above, but it will also
remove the container registry with the Docker images and the cluster container
from the host. With k3d, run:

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

#### Flush

Also in the Tilt UI header, click the trash button to flush the existing data.
This will recreate the MySQL databases and the MinIO buckets, delete the
Opensearch index and restart the required resources.

### Known issues

#### Minio uploads don't trigger workflows

The setup of the Minio buckets and the communication between Minio and Redis
is sometimes not setup properly. To solve it, from the Tilt UI, restart the
`minio` resource and then trigger the `minio-setup-buckets` resource.

[national health archive (nha)]: https://www.piql.com/norwegians-digital-health-data-to-be-preserved-for-future-generations/
[national center for truth and reconciliation (nctr)]: https://nctr.ca/about/about-the-nctr/our-mandate/
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
