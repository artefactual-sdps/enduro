# Ingest

In Enduro, **[Ingest](../glossary.md#ingest)** is defined as a phase in a
preservation workflow describing all the preservation policy-defined tasks
performed on a SIP when it is received from a producer, prior to preservation.
Typically this phase covers **validation** activities (performed against SIP
files, structure, and/or metadata) as well as any **package transformations**
(removal of unneeded or temporary files, restructuring, etc) to optimize the
package for further processing by the preservation engine.

At installation, Enduro's default ingest functionality is minimal - see the
[Default ingest workflow](managing-ingest-workflows.md#default-ingest-workflow)
documentation for more details. However, Enduro's workflows are intended to be
customized via the addition of **[child workflow activities][child workflow]**,
which can be designed to implement the specific ingest needs of a given
organization.

The Enduro project maintains general default workflow activities in a separate
code repository, called [temporal-activities][temporal-activities]. An example
of child workflow activities for a specific organization can be seen in the
[preprocessing-sfa][preprocessing-sfa] repository. Artefactual also maintains a
template that organizations can use to create their own child workflow
activities repository, called [preprocessing-base][preprocessing-base].

[child workflow]: ../../dev-manual/preprocessing.md
[preprocessing-base]: https://github.com/artefactual-sdps/preprocessing-base
[preprocessing-sfa]: https://github.com/artefactual-sdps/preprocessing-sfa
[temporal-activities]: https://github.com/artefactual-sdps/temporal-activities
