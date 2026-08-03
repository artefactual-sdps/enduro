# Poststorage workflows

A poststorage child workflow runs institution specific activities after an AIP
has been stored. It can notify an external system, synchronize metadata, or
perform other work that needs the stored AIP's identifier.

The parent is the SIP processing workflow. Enduro waits for the poststorage
result before completing that workflow.

## Execution and contract

For [a3m] processing, Enduro starts poststorage after the accepted AIP has been
moved to permanent storage and that operation has completed. For
[Archivematica], it starts after preservation processing has completed and
Enduro has registered the stored AIP. The poststorage workflow is only run if
the AIP is stored successfully; it is not run if preservation or storage fail,
or if the AIP is rejected.

The Temporal child workflow type is the configured `workflowName`. Enduro
dispatches it to the configured `taskQueue` with these options:

* `WorkflowID`: `<workflowName>-<AIP UUID>`
* `TaskQueue`: the configured `taskQueue`
* `ParentClosePolicy`: `PARENT_CLOSE_POLICY_TERMINATE`

The poststorage worker must register the workflow using the configured
`workflowName`, poll the configured queue, and connect to Enduro's Temporal
namespace.

Enduro does not explicitly set a child timeout or retry policy. Other child
options follow Temporal's inheritance and default rules. Workflow executions
do not have a retry policy by default.

A poststorage workflow is executed with `childwf.PostStorageParams`:

| Field | Type | Value supplied by Enduro |
| --- | --- | --- |
| `User` | `*childwf.User` | The initiating user's email, or `nil` when unavailable. |
| `AIPUUID` | `string` | The stored AIP's UUID as a string. |
| `CustomMetadata` | `childwf.CustomMetadata` | Metadata from successful preprocessing, or an empty value if no metadata was produced. |

The passed parameters do not include the SIP ID or name, a filesystem path, a
storage location, or a complete AIP record. Use the AIP UUID to query the Enduro
or the Archivematica Storage Service APIs if the integration needs more
information about the stored AIP.

A poststorage workflow returns a `childwf.PostStorageResult`:

| Field | Type | Meaning |
| --- | --- | --- |
| `Outcome` | `childwf.Outcome` | `OutcomeSuccess`, `OutcomeSystemError`, or `OutcomeContentError`. |
| `CustomMetadata` | `childwf.CustomMetadata` | Additional opaque JSON metadata. |
| `Tasks` | `[]*childwf.Task` | The child workflow's returned task records. |

## Result handling

The Enduro processing workflow waits for the poststorage workflow to complete
and then, in order:

1. merges returned custom metadata into the preprocessing metadata at the
   map's top level;
2. saves the returned tasks; and
3. interprets the returned outcome.

This order also applies to a decodable content or system failure result.

Enduro does not receive tasks in real time: they appear through the API and
dashboard only after the child has completed and Temporal has returned its
result.

The outcome determines the remaining SIP processing behavior:

* `OutcomeSuccess`: continue and finish normal processing.
* `OutcomeContentError`: set the SIP status to "failed" and exit the processing
  workflow with a content error.
* `OutcomeSystemError`: set the SIP status to "error" and exit the processing
  workflow with a system error.
* Any other value: treat the result as a system error.

For an expected validation, policy, or integration failure, record the details
in completed tasks, set the structured outcome, and return the result with a
nil Go error. Enduro then saves the task records before applying the failure
state. See [Results and workflow errors] for the shared rule and helpers.

An actual Temporal workflow error prevents Enduro from decoding the result, so
no returned tasks or metadata are available. Enduro treats an error other than
cancellation as a system error.

If the Enduro parent closes while poststorage is active, the Parent Close
Policy terminates the child.

Because storage has already completed, neither a structured failure nor a
workflow error rolls back or removes the AIP. Enduro may subsequently mark the
SIP and its processing workflow as failed or in error.

## How metadata is merged

Preprocessing metadata is the base map. For every entry returned by
poststorage, Enduro assigns its `json.RawMessage` to the same key in the base
map. Thus:

* a new poststorage key is added;
* an existing key is replaced in full by the poststorage value;
* Enduro does not recursively merge JSON objects inside a value; and
* an empty poststorage map leaves preprocessing metadata unchanged.

A poststorage value replaces the preprocessing value when their keys conflict.
Use distinct keys when both values must be retained.

The final merged metadata can be supplied to a postbatch workflow, but only if
the overall SIP processing workflow completes successfully. Metadata from a
poststorage result may be merged into the processing workflow's state before
Enduro applies its failure outcome, but failed or canceled SIP processing does
not return that metadata to the batch parent.

## Enable the workflow

Add one `[[childWorkflows]]` entry with `type = "poststorage"`, the worker's
task queue, and its registered workflow name. See the administrator manual's
[poststorage configuration] and the [custom workflow template].

For complete implementations used in production, see the [SFA APIS poststorage
workflow] and [SFA cantons poststorage workflow].

[a3m]: ../../user-manual/components/#preservation-engine
[Archivematica]: ../../user-manual/components/#preservation-engine
[custom workflow template]: https://github.com/artefactual-sdps/custom-enduro-workflows/blob/main/README.md
[poststorage configuration]: ../../admin-manual/configuration.md#poststorage-child-workflow
[Results and workflow errors]: index.md#results-and-workflow-errors
[SFA APIS poststorage workflow]: https://github.com/artefactual-sdps/sfa-enduro-workflows/blob/main/internal/workflows/poststorage_apis.go
[SFA cantons poststorage workflow]: https://github.com/artefactual-sdps/sfa-enduro-workflows/blob/main/internal/workflows/poststorage_cantons.go
