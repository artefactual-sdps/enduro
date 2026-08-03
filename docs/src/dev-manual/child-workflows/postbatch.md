# Postbatch workflows

A postbatch child workflow performs work after the SIPs in the batch have been
processed. Enduro's batch workflow is the postbatch workflow's parent and it
does not finalize the batch until postbatch returns. The SIP processing workflow
started for each SIP in the batch is a separate child execution. Postbatch is
suited to reports and external integrations that need the batch as a whole.

## When it runs

Postbatch is reachable only after every SIP in a batch first reaches `validated`
status. If polling finds a terminal SIP status before every SIP reaches
`validated`, Enduro fails the batch and skips postbatch. Once all SIPs validate,
Enduro lets their processing workflows continue and polls until every SIP
reaches a final ingest status.

If all SIPs were ingested, Enduro waits for every processing workflow to finish
and then proceeds to postbatch. If any SIP failed, entered an error state, or
was canceled after validation, the batch becomes pending:

* Continuing makes Enduro wait for every processing workflow to finish before
  constructing the postbatch parameters.
* Canceling skips postbatch, marks the batch and ingested SIPs as canceled, and
  requests deletion of AIPs for ingested or already canceled SIPs.

Continuing a partial batch does **not** filter the postbatch input to successful
SIPs. Enduro includes one `PostbatchSIP` entry for every SIP originally in the
batch.

## Execution and contract

The Temporal child workflow type name is the configured `workflowName`. Enduro
dispatches it to the configured `taskQueue` with these options:

* `WorkflowID`: `<workflowName>-<batch UUID>`
* `TaskQueue`: the configured `taskQueue`
* `ParentClosePolicy`: `PARENT_CLOSE_POLICY_TERMINATE`

The postbatch worker must register the workflow using the configured
`workflowName`, poll the configured queue, and connect to Enduro's Temporal
namespace.

Enduro does not explicitly set a child timeout or retry policy. Other child
options follow Temporal's inheritance and default rules. Workflow executions
do not have a retry policy by default.

A postbatch workflow is executed with `childwf.PostbatchParams`:

| Field | Type | Value supplied by Enduro |
| --- | --- | --- |
| `User` | `*childwf.User` | The initiating user's email, or `nil` when unavailable. |
| `Batch` | `*childwf.PostbatchBatch` | Batch `UUID`, `Identifier`, and total `SIPSCount`. |
| `SIPs` | `[]*childwf.PostbatchSIP` | One record with final data for every SIP in the batch. |

The exact `PostbatchSIP` fields are:

| Field | Type | Value supplied by Enduro |
| --- | --- | --- |
| `UUID` | `uuid.UUID` | SIP UUID. |
| `Name` | `string` | SIP name. |
| `AIPID` | `*uuid.UUID` | UUID of the stored AIP when one was recorded; otherwise `nil`. |
| `FileCount` | `int32` | The final number of files recorded for the SIP. |
| `CustomMetadata` | `childwf.CustomMetadata` | Metadata returned from the SIP processing workflow, if it completed successfully. |

`AIPID` is not a status field. A SIP that failed after storage can still have an
`AIPID`. Failed or canceled processing produces no `CustomMetadata` in the
postbatch entry, even if preprocessing or poststorage produced metadata,
because only a successful processing result returns merged metadata to the
batch. Select successfully processed SIPs according to the data and business
rules required by the integration.

A postbatch workflow returns a `childwf.PostbatchResult`:

| Field | Type | Meaning |
| --- | --- | --- |
| `Outcome` | `childwf.Outcome` | A structured outcome set by the child. |
| `Message` | `string` | A result message set by the child. |

## Result and failure behavior

Enduro logs every decoded `Outcome` and `Message` but does not branch on the
result, save it in Enduro's database, or expose it through the API or dashboard.
A decodable result returned with a nil Go error therefore allows the batch to
finish as ingested, regardless of its structured outcome.

`PostbatchResult` has no task or metadata fields, and the batch parent does not
collect tasks from this child. Postbatch tasks cannot be displayed through the
Enduro API or dashboard.

For an expected postbatch or domain failure that should not fail the batch, set
the result's `Outcome` and `Message` and return it with a nil workflow error.
The details are then available only in the parent workflow log. To fail the
batch, return a workflow error; on error Enduro will receive a nil
`childwf.PostbatchResult`.

If the postbatch child returns a workflow error or is canceled while the batch
parent is waiting, `Get` returns an error, no structured result is decoded, and
Enduro sets the batch status to failed. Already stored AIPs are not removed as
a consequence of a postbatch failure. If the parent workflow closes because of a
failure or timeout while the child is active, the configured parent close
policy terminates the child and no structured result is available.

The custom postbatch child exchanges no signals with Enduro. The
`batch-decision-signal` for a partial batch is exchanged only between the
Enduro API and the batch parent.

## Enable the workflow

Add one `[[childWorkflows]]` entry with `type = "postbatch"`, the worker's task
queue, and the workflow type name. See the administrator manual's [postbatch
configuration] and the [custom workflow template].

The [CVA postbatch workflow] is a complete example used in production. Its
[CSV activity] demonstrates selecting entries with a usable AIP UUID.

[CVA postbatch workflow]: https://github.com/artefactual-sdps/cva-enduro-workflows/blob/main/internal/workflows/postbatch.go
[CSV activity]: https://github.com/artefactual-sdps/cva-enduro-workflows/blob/main/internal/activities/create_csv.go#L96-L102
[custom workflow template]: https://github.com/artefactual-sdps/custom-enduro-workflows/blob/main/README.md
[postbatch configuration]: ../../admin-manual/configuration.md#postbatch-child-workflow
