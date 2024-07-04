# Enduro SDPS

[Enduro][Enduro] is a new application under development by Artefactual Systems.
Originally created as a more stable replacement for Archivematica’s
[automation-tools][automation-tools] library of scripts, it has since evolved
into a flexible tool to be paired with preservation applications to provide
initial ingest activities such as content and structure validation, packaging,
and more.

While still under development, Enduro is already being used in production at
several large cultural heritage organizations. Enduro aims to cover our clients'
needs while exploring new and innovative ways to build durable and
fault-tolerant workflows suited for preservation. It is also a new space for
experimentation and research to build a distributed system that enables users to
run failure-oblivious and durable preservation workflows.

This version of Enduro can use either [a3m][a3m] or
[Archivematica][archivematica] as its preservation engine, alongside
[MinIO][MinIO] for object storage and [Temporal][Temporal] to manage the
workflow. This combination of tools is intended to be lightweight, scalable, and
easy to install.

[a3m]: https://github.com/artefactual-labs/a3m
[archivematica]: https://archivematica.org
[automation-tools]: https://github.com/artefactual/automation-tools
[Enduro]: https://github.com/artefactual-sdps/enduro
[MinIO]: https://min.io/
[Temporal]: https://temporal.io/
