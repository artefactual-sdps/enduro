# Ingest

In Enduro, **[Ingest]** is defined as a phase in a preservation workflow
describing all the preservation policy-defined tasks performed on a SIP when it
is received from a producer, prior to preservation. Typically this phase covers
**validation** activities (performed against SIP files, structure, and/or
metadata) as well as any **package transformations** (removal of unneeded or
temporary files, restructuring, etc) to optimize the package for further
processing by the preservation engine.

Enduro provides a minimal [Default ingest workflow] that organizations can
extend with [child workflows].

[child workflows]: ../glossary.md#child-workflow
[Default ingest workflow]: managing-ingest-workflows.md#default-ingest-workflow
[Ingest]: ../glossary.md#ingest
