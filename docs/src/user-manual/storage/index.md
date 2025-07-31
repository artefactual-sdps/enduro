# Storage

The Storage component of Enduro is a lightweight interface between the
configured [preservation engine] and whatever storage devices have been
configured for the long-term preservation and ongoing maintenance of [AIPs][AIP]
produced by Enduro workflows.

Currently, Enduro uses the **Archivematica Storage Service (AMSS)** to provide
the underlying functionality for managing AIPs and their storage locations. For
more details see the [AMSS documentation][AMSS].

If you are using [a3m] as your preservation engine instead of [Archivematica],
an administrator will need to configure a storage solution to be used as an AIP
store. For more information, consult the Administrator's manual.

This section of the User manual will cover the management of Locations and AIPs.

[a3m]: https://github.com/artefactual-labs/a3m
[AIP]: ../glossary.md#archival-information-package-aip
[AMSS]: https://www.archivematica.org/docs/storage-service-latest/
[Archivematica]: https://archivematica.org
[preservation engine]: ../glossary.md#preservation-engine
