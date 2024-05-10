# Working with Archivematica

If you choose Archivematica (AM) as the preservation system in your local
development environment, you will need access to an AM instance outside the
Kubernetes cluster. The target AM instance must have an SFTP server configured
to be able to upload transfers from Enduro.

## Configuration

In order to configure the connection to the SFTP server and the AM API, you'll
need to create three `.secret` files inside the `dev-am` overlay folder. These
files are used by a [Kustomize secret generator] to add the `enduro-am-secret`
inside the cluster, and they are not tracked in the repository.

!!! note

    The secret generator is used via Tilt so users CANNOT manually apply
    the Kubernetes commands for these operations. Understanding the way these
    files are used is NOT required to work with Archivematica.

### Quick checklist for configuration files

#### `.am.secret`

- Location: `hack/kube/overlays/dev-am/.am.secret`
- **Contents to check:**
    - AM API address (e.g.,`http://host.k3d.internal:62080`)
    - User credentials (`user=test`, `api_key=test`)
    - SFTP configuration
      details (`sftp_host=`, `sftp_port=`, `sftp_user=`, `sftp_remote_dir=`,
      `sftp_private_key_passphrase=`).
    - Archivematica Storage Service credentials(`amss_url=`,
      `amss_user=`, `amss_api_key=`). These credentials are used by the
      *mysql-create-amss-location-job.yaml* job to add an AMSS location to the
      *enduro_storage* database.

#### `.id_ed25519.secret`

- Location: `hack/kube/overlays/dev-am/.id_ed25519.secret`
- **Contents to check:**
    - SSH private key (Ensure it starts with `-----BEGIN
      OPENSSH PRIVATE KEY-----` and ends with `-----END
      OPENSSH PRIVATE KEY-----`)

#### `.known_hosts.secret`

- Location: `hack/kube/overlays/dev-am/.known_hosts.secret`
- **Contents to check:**
    - Known hosts entries (Look for entries starting with
      `|1|` and containing `ssh-rsa`, `ecdsa-sha2-nistp256`,
      `ssh-ed25519` etc.)

#### `.tilt.env`

- Location: `root/`
- **Contents to check:**
    - `ENDURO_PRES_SYSTEM = "am"`

#### `enduro.toml`

- Location: `root/`
- **Contents to check:**
    - `[preservation] taskQueue` variable must be set to "am"

!!! note

    If you modify these files in a running environment, you NEED to refresh the
    `(Tiltfile)` (first) and the `enduro-am` (second) resources in the Tilt UI
    to apply those changes inside the cluster.

### `hack/kube/overlays/dev-am/.am.secret`

AM API and SFTP configuration:

    address=http://host.k3d.internal:62080
    user=test
    api_key=test
    sftp_host=host.k3d.internal
    sftp_port=2222
    sftp_user=archivematica
    sftp_remote_dir=/enduro_transfers
    sftp_private_key_passphrase=
    amss_url=http://host.k3d.internal:62081
    amss_user=test
    amss_api_key=secret

### `hack/kube/overlays/dev-am/.id_ed25519.secret`

SSH key used to authenticate against the SFTP server:

    -----BEGIN OPENSSH PRIVATE KEY-----
    ...
    -----END OPENSSH PRIVATE KEY-----

### `hack/kube/overlays/dev-am/.known_hosts.secret`

Known hosts needed for a seamless connection to the SFTP server:

    |1|LukyHignP9f6C5UMHNeJsrpLozk=|I448JU6j5g4jCZxHTI0YdYckZlw= ssh-rsa ...
    |1|0OoDhmFh2UJAjMRqU68Fq1tpJUI=|Yk0nWoBneUp5ByxpuuMrc/GWrM0= ecdsa-sha2-nistp256 ...
    |1|rc8AmaUEs81zyOtVSk4dGM7snaE=|tuYxgEJdh2T1WwDh5rHfN1jrVIs= ssh-ed25519 ...

### `root/.tilt.env`

Enduro preservation system value needed for the Archivematica container:

    ENDURO_PRES_SYSTEM = "am"

There is more information on the configuration of the [Tilt Environment].

### `root/enduro.toml`

Preservation system value needed for workflow to be Archvimatica specific:

    [preservation]
    taskqueue = "am"

An amssLocationId is needed for AIP download:

    [am]
    amssLocationId = "e0ed8b2a-8ae2-4546-b5d8-f0090919df04"

!!! note

    The amssLocationId UUID must match the AMSS location ID in the
    *enduro_storage* database — it does *not* need to match the actual location
    UUID in the AM Storage Service.

[kustomize secret generator]: https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-kustomize/#create-a-secret
[tilt environment]: devel.md#tilt-environment-configuration
