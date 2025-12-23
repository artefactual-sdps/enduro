# Working with Archivematica

Enduro's dev-am overlay runs a bundled Archivematica instance (ambox) inside the
cluster. SFTP is exposed by that pod for transfer uploads, so you don't need to
run AM or SFTP on your host.

## Quick start

By default, dev-am uses Archivematica. Only touch these if you previously
switched to a3m:

1. Make sure Enduro uses Archivematica (in `.tilt.env`):

       ENDURO_PRES_SYSTEM=am

2. Ensure the preservation task queue is set to AM (in `enduro.toml`):

       [preservation]
       taskQueue = "am"

That's it. The dev-am overlay ships defaults that match the `ambox` pod.

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
