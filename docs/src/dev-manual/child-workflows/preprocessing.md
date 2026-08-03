# Preprocessing workflows

A preprocessing child workflow prepares a SIP after Enduro has downloaded it
and before Enduro sends a package to the preservation engine. It is the
extension point for validation, extraction, restructuring, metadata generation,
and other preparation needed by an organization.

The parent is Enduro's processing workflow. It waits for preprocessing to
finish, interprets the returned outcome, and requires the successful result to
identify a BagIt bag that it can continue processing.

## Execution and contract

Before starting the child, Enduro determines the extension and checksum for a
SIP that is not a directory and performs the duplicate check when
`ingest.allowDuplicates = false`. It may also extract the object as described
in [Extraction](#extraction).

The Temporal child workflow type is the configured `workflowName`. Enduro
dispatches it to the configured `taskQueue` with these options:

* `WorkflowID`: `<workflowName>-<SIP UUID>`
* `TaskQueue`: the configured `taskQueue`
* `ParentClosePolicy`: `PARENT_CLOSE_POLICY_TERMINATE`

The preprocessing worker must register the workflow using the configured
`workflowName`, poll the configured queue, and connect to Enduro's Temporal
namespace.

Enduro does not explicitly set a child timeout or retry policy. Other child
options follow Temporal's inheritance and default rules. Workflow executions
do not have a retry policy by default.

A preprocessing workflow is executed with `childwf.PreprocessingParams`:

| Field | Type | Value supplied by Enduro |
| --- | --- | --- |
| `User` | `*childwf.User` | The initiating user's email, or `nil` when no user is associated with the ingest. |
| `RelativePath` | `string` | The SIP path below the shared storage root. Enduro derives it from the `sharedPath` in its `[[childWorkflows]]` entry. |
| `SIPID` | `uuid.UUID` | The Enduro SIP UUID. |
| `BatchID` | `uuid.UUID` | The batch UUID, or `uuid.Nil` when the SIP is not in a batch. |
| `SIPName` | `string` | The SIP name recorded by Enduro. |

A preprocessing workflow returns a `childwf.PreprocessingResult`:

| Field | Type | Meaning |
| --- | --- | --- |
| `Outcome` | `childwf.Outcome` | `OutcomeSuccess`, `OutcomeSystemError`, or `OutcomeContentError`. |
| `CustomMetadata` | `childwf.CustomMetadata` | Opaque JSON values to make available to later custom workflows. |
| `RelativePath` | `string` | The final package path below the same shared storage root. The child must return a relative path, not its absolute path. |
| `Tasks` | `[]*childwf.Task` | The child workflow's returned task records. |

## Shared path

A preprocessing child reads and may change the SIP files. For filesystem
access, Temporal sends `RelativePath`; it neither transfers the files nor
exposes either worker's absolute path. The Enduro worker and custom worker must
therefore have access to the same storage.

Enduro downloads the SIP below the `sharedPath` configured in its
`[[childWorkflows]]` entry and derives `params.RelativePath` from the path below
that root.

The custom worker must configure the root as it sees it. The example Go
projects also call this setting `sharedPath`, but it belongs to the custom
project and is not part of the `pkg/childwf` contract. The two absolute paths
may differ. For example, the [template project] uses this mapping:

```text
Enduro sharedPath:       /home/enduro/preprocessing
Child sharedPath:        /home/enduro/shared
params.RelativePath:     ingest-dir/sip.zip

Path seen by Enduro:     /home/enduro/preprocessing/ingest-dir/sip.zip
Path seen by the child:  /home/enduro/shared/ingest-dir/sip.zip
```

Both full paths refer to the same file on `preprocessing-pvc`. The child joins
its own root to the received value before it reads or changes the SIP:

```go
sipPath := filepath.Join(cfg.SharedPath, params.RelativePath)
```

The result follows the same rule in the other direction. If the final package
stays in place, copy `params.RelativePath` to `result.RelativePath`. If the
child moves it, derive a new path relative to the child's shared root. Return a
path within the shared storage, not an absolute path from either worker.

On a successful outcome, Enduro cleans `result.RelativePath`, joins it to its
own `sharedPath`, and uses that full path for the rest of processing. See the
[template Tilt setup] for the settings and mounts that provide this storage in
the local development cluster.

## Extraction

The preprocessing `extract` setting controls only whether Enduro runs its
archive extraction step before it starts the child. It is not a field in
`childwf.PreprocessingParams`, so the child does not receive its value. Set it
to match what the registered workflow expects:

* With `extract = false`, the default, Enduro attempts to extract a downloaded
  file before it starts the child. After a successful extraction,
  `params.RelativePath` points to the extracted directory. If Enduro does not
  recognize the file as an archive, the path still points to the original
  file. Enduro leaves a downloaded directory unchanged. Any other extraction
  error stops processing before the child starts.
* With `extract = true`, Enduro skips its archive extraction step.
  `params.RelativePath` points to the original downloaded file or directory.
  The child decides whether extraction is needed and must update
  `result.RelativePath` if the final package is in a different place.

For a source that is not a directory, Enduro calculates the checksum and
performs any duplicate check before either extraction path. Shared storage is
always required. Regardless of the `extract` value, a successful child must
leave a BagIt bag in shared storage and return its relative directory path.
Enduro then classifies and validates the returned package and stops SIP
processing if it is not a BagIt bag.

The [template project] uses `extract = false` and bags the received path,
whereas the [SFA preprocessing workflow] uses `extract = true` and extracts the
original archive itself.

## Results, tasks, and metadata

Enduro does not receive `Tasks` as they are created. Returned tasks become
available only when Temporal returns a decodable result. A pending task that
Enduro creates for a [human decision](#human-decisions) is the exception.

For a decodable result, Enduro adopts the returned path and metadata only when
the outcome is successful. It then saves, attaches, and exposes the returned
tasks before applying the outcome behavior:

* `OutcomeSuccess`: use `RelativePath`, retain `CustomMetadata`, and continue
  processing.
* `OutcomeContentError`: set the SIP status to "failed" and exit the processing
  workflow with a content error.
* `OutcomeSystemError`: set the SIP status to "error" and exit the processing
  workflow with a system error.
* Any other value: treat the result as a system error.

Metadata from a successful preprocessing result is passed to the poststorage
workflow and, if processing ultimately succeeds, to the postbatch workflow.
Enduro treats the JSON values as opaque.

For an expected validation, policy, or processing failure, populate the result
outcome and completed tasks and return the result with a nil Go error. This
preserves the error details and lets Enduro apply the intended SIP failure
state. See [Results and workflow errors] for the shared rule and helper
functions.

If the child instead returns a Temporal workflow error, Enduro cannot decode a
result at the same time. It therefore receives no returned tasks, path, or
metadata and treats an error other than cancellation as a system error.

If the Enduro parent closes while preprocessing is active, the configured
parent close policy terminates the child.

## Human decisions

Preprocessing is the only custom extension point with an Enduro bridge for a
human decision. The bridge uses two Temporal Signals between child and parent,
a Query for API reads, and an Update for API writes.

The child initiates the exchange by signaling its exact parent execution and
then waiting for the response signal:

```go
info := workflow.GetInfo(ctx)
if info.ParentWorkflowExecution == nil {
    return nil, errors.New("decision request requires a parent workflow")
}
request := childwf.DecisionRequest{
    Message: "The package needs review.",
    Options: []string{"Continue", "Reject"},
}
err := workflow.SignalExternalWorkflow(
    ctx,
    info.ParentWorkflowExecution.ID,
    info.ParentWorkflowExecution.RunID,
    childwf.DecisionRequestSignalName,
    request,
).Get(ctx, nil)
if err != nil {
    return nil, err
}

var response childwf.DecisionResponse
workflow.GetSignalChannel(
    ctx,
    childwf.DecisionResponseSignalName,
).Receive(ctx, &response)
```

Production code should classify signaling failures into an appropriate result
or workflow error. The [SFA preprocessing workflow] contains a complete
implementation.

The full flow is:

1. The child signals `decision-request-signal` with
   `childwf.DecisionRequest{Message string, Options []string}`. Its message must
   not be blank and it must provide at least one option.
2. The Enduro processing workflow receives and validates the request. Only one
   request may be pending at a time. It creates a pending Enduro task named
   `Preprocessing workflow is waiting for user decision`, records the request
   in its Temporal workflow state, then changes the SIP and workflow to pending.
3. `GET /ingest/sips/{uuid}/decision` queries the parent with
   `child-decision-query`. The dashboard displays the message and options to a
   user with the `ingest:sips:decision` permission.
4. The user submits `{"option":"<selected option>"}` through
   `POST /ingest/sips/{uuid}/decision`. Enduro validates that the option was
   offered, creates `childwf.DecisionResponse{Option string}`, then sends the
   parent the tracked `child-decision-update` Update.
5. The parent records the response in workflow state, signals the exact child
   execution with `decision-response-signal`, restores processing statuses, and
   completes the pending Enduro task with the selected option.
6. The child's call to `Receive` returns. The child interprets
   `response.Option` and resumes its own workflow logic.

Signals are asynchronous messages recorded in Temporal history. The API waits
for the parent Update handler to accept and store the response, not for the
child to resume or finish. The pending request and response are durable
workflow state that Temporal restores from history. The visible task and entity
statuses are also persisted in Enduro's database. The child must keep waiting on
`decision-response-signal` to receive the choice.

Option strings have no predefined meaning. In particular, an option named
`Cancel` does not cancel the workflow automatically. The child must classify
and represent the chosen outcome itself.

## Enable the workflow

Add one `[[childWorkflows]]` entry with `type = "preprocessing"`, the worker's
task queue and registration name, and the required shared path. Set `extract`
according to whether Enduro or the child extracts the SIP. See the administrator
manual's [preprocessing configuration] for the complete settings and the
[custom workflow template] for worker and local Kubernetes setup.

[custom workflow template]: https://github.com/artefactual-sdps/custom-enduro-workflows/blob/main/README.md
[preprocessing configuration]: ../../admin-manual/configuration.md#preprocessing-child-workflow
[Results and workflow errors]: index.md#results-and-workflow-errors
[SFA preprocessing workflow]: https://github.com/artefactual-sdps/sfa-enduro-workflows/blob/main/internal/workflows/preprocessing.go
[template project]: https://github.com/artefactual-sdps/custom-enduro-workflows
[template Tilt setup]: https://github.com/artefactual-sdps/custom-enduro-workflows/blob/main/README.md#develop-with-enduro
