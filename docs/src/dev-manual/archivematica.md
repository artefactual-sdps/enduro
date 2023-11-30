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

    If you modify these files in a running environment, you need to refresh the
    `(Tiltfile)` (fist) and the `enduro-am` (second) resources in the Tilt UI
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

[kustomize secret generator]: https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-kustomize/#create-a-secret
