# Working with Archivematica

Enduro's dev-am overlay runs a bundled Archivematica instance (ambox) inside the
cluster. SFTP is exposed by that pod for transfer uploads, so you don't need to
run AM or SFTP on your host.

## Quick start

By default the development environment uses the dev-am overlay. If you were
using a3m previously, make sure to update the `.tilt.env` file to use
Archivematica: `ENDURO_PRES_SYSTEM=am`. Changing this `.tilt.env` is the only
change necessary, as the dev-am overlay updates the Enduro configuration as
required to integrate with the Archivematica `ambox` pod.

## Defaults (dev-am)

The dev-am overlay includes a default `enduro-am-secret` with:

- AM API user + key (both `test`)
- AMSS user + key (both `test`)
- AMSS location ID
- SSH private key for SFTP
- SFTP host key + known_hosts entry (pinned for stable connections)

Enduro connects to the in-cluster `ambox` service:

- AM Dashboard: `http://ambox.enduro-sdps:64080`
- AM Storage Service: `http://ambox.enduro-sdps:64081`
- SFTP: `ambox.enduro-sdps:64022`
- Transfer source: `transfers` (SFTP path `/`)

The dev-am overlay ships a fixed host key and the matching known_hosts entry.

## Overriding defaults

If you want different credentials or endpoints, provide your own secret and
update the dev-am overlay to reference it.

## Applying changes

If you edit `.tilt.env` or `enduro.toml` while Tilt is running, refresh the
Tiltfile and the `enduro-am` resource to apply the changes inside the cluster.
