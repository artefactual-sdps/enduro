# Components

## MinIO

[MinIO](https://min.io/) is a flexible, high performance object storage
platform. Enduro uses MinIO as its storage back-end for both uploading
submission information packages (SIPs) and storing archival information packages
(AIPs). Material intended for preservation can be uploaded to MinIO either
through the user interface or via command line using the [MinIO
client](https://min.io/docs/minio/linux/reference/minio-mc.html). Any time new
content is uploaded to a designated bucket in MinIO, a transfer is started in
Enduro.

## Temporal

[Temporal](https://temporal.io/) is responsible for orchestrating Enduro's
workflows - that is, for kicking off tasks, managing them, and recording them as
auditable events. It also manages retries and timeouts, resulting in a reliable
platform that can process digital objects for preservation in a highly automated
environment.

## a3m

[a3m](https://github.com/artefactual-labs/a3m) is a streamlined version of
[Archivematica](https://archivematica.org) that is wholly focused on AIP
creation. It does not have external dependencies, integration with access
systems, search capabilities, or a graphical interface. It was designed to
reduce the bulk of Archivematica's extraneous functions for users operating at a
large scale who are more focused on throughput of digital objects for
preservation.
