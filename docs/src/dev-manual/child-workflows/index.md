# Custom child workflows

Custom Temporal child workflows let an organization extend Enduro's ingest
process without adding custom rules to Enduro itself. As it processes a SIP, the
[processing workflow] starts **preprocessing** and **poststorage** workflows as
children; once the SIP workflows in a batch have ended, the [batch workflow]
can start a **postbatch** child. This guide describes the contracts and behavior
of all three extension points, while the administrator manual covers their
[child workflow configuration].

## Temporal concepts

An Enduro child workflow uses the following Temporal concepts:

* A **Workflow Definition** is deterministic orchestration code. A **Workflow
  Execution** is one running instance of that definition, and a **Workflow
  Type** is the name registered for it. Temporal records execution history and
  replays workflow code to restore its state. See [Temporal Workflows].
* An **Activity** is a function or method that performs a specific action, such
  as filesystem I/O or an API call. Activity code does not have to be
  deterministic and should normally be idempotent. Workflow code schedules and
  coordinates activities. See [Temporal Activities].
* A **Worker** registers workflow and activity implementations and polls a
  **Task Queue**. Enduro dispatches each custom child workflow to its configured
  task queue, so a custom worker must poll that exact queue and register the
  configured workflow name. See [Temporal Workers] and [Temporal Task Queues].
* A **Child Workflow Execution** is started by another workflow in the same
  Temporal namespace. Parent and child have separate histories and do not share
  local state. They exchange serializable parameters, results, and, when
  needed, messages. See [Temporal Child Workflows].
* A **Signal** is an asynchronous write. The sender cannot receive a result
  from the signal handler. Enduro normally exchanges parameters and a final
  result with its child workflows, and uses Signals only for preprocessing
  decisions. A preprocessing child sends a decision request to its parent with
  a Signal, and the parent signals the selected response back. See [Temporal
  message passing] for more about Signals.

Temporal documents SDKs for [.NET, Go, Java, PHP, Python, Ruby, Rust, and
TypeScript][Temporal SDKs]. This guide uses the [Temporal Go SDK] because Enduro
and the existing custom workflow projects are written in Go.

## Extension points

Each extension point has a distinct contract and effect on its parent workflow.

### Preprocessing

Preprocessing runs after Enduro downloads the SIP and performs its initial file
checks, and before preservation processing. The `extract` setting determines
whether Enduro attempts archive extraction before the child or leaves extraction
to the child. This extension supports validation, extraction, restructuring,
metadata generation, and other preparation before preservation.

**Contract:** The child receives the SIP's relative working path and
identifiers. It returns a relative path to the modified SIP, which must be
structured as a [BagIt] bag. It also returns data about the tasks run by the
child workflow, and optional custom metadata that can be read by poststorage
and postbatch workflows.

**Behavior:** A successful result replaces the processing path and supplies
metadata to later custom workflows. A content or system outcome stops SIP
processing; Enduro does not adopt the returned path or metadata. See
[Preprocessing workflows] for the complete contract and behavior.

### Poststorage

Poststorage runs after the AIP has been stored and registered and before the
SIP processing workflow completes. It supports hooks and integrations that
need a stored AIP.

**Contract:** The child receives the AIP UUID and preprocessing metadata. It
returns tasks, an outcome, and more metadata.

**Behavior:** Enduro waits for the result and merges the returned metadata. A
content or system outcome stops the SIP workflow but does not remove the stored
AIP. See [Poststorage workflows] for the complete contract and behavior.

### Postbatch

Postbatch runs after the workflows that process the SIPs in a batch have ended.
If only part of the batch is then ingested, it runs only after a user chooses to
continue. It supports reports and integrations that need data from the completed
SIP workflows in a batch.

**Contract:** The child receives batch data and one entry for every SIP that
was originally in the batch, not only the SIPs that succeeded. It returns an
outcome and message.

**Behavior:** Enduro logs the structured result but does not otherwise act on
its outcome. An actual Temporal workflow error fails the batch. See [Postbatch
workflows] for the complete contract and behavior.

## Configuration and dispatch

Enduro accepts at most one configuration for each child workflow `type`.
`taskQueue` and `workflowName` are deployment values, not fixed Enduro
constants. The Enduro child workflows configuration for both values much exactly
match the values registered in the actual child workflow.

```toml
[[childWorkflows]]
type = "preprocessing"
taskQueue = "custom-enduro"
workflowName = "preprocessing"
extract = false
sharedPath = "/home/enduro/preprocessing"
```

Child workflows inherit `[temporal].namespace` from their Enduro parent, and
the custom worker must connect to that namespace. A `[[childWorkflows]]` entry
therefore has no separate `namespace` field. See [child workflow configuration]
for all three types.

## The shared `pkg/childwf` package

[`pkg/childwf`][childwf source] is the public Go package containing the data
contracts shared by Enduro and projects that define custom workflows:

```go
import "github.com/artefactual-sdps/enduro/pkg/childwf"
```

It defines concrete named types, structs, constants, and helper functions; it
does not define a workflow interface. The important types are:

| Type | Kind and purpose |
| --- | --- |
| `Outcome` | Named `int` with `OutcomeSuccess` (`0`), `OutcomeSystemError` (`1`), and `OutcomeContentError` (`2`). |
| `CustomMetadata` | Named `map[string]json.RawMessage` for opaque JSON data passed between child workflows. |
| `User` | Struct containing the initiating user's `Email`, when available. |
| `PreprocessingParams`, `PreprocessingResult` | Preprocessing input and result structs. |
| `PostStorageParams`, `PostStorageResult` | Poststorage input and result structs. |
| `PostbatchParams`, `PostbatchBatch`, `PostbatchSIP`, `PostbatchResult` | Postbatch input, nested batch/SIP data, and result structs. |
| `Task` | A child task record with `Name`, `Message`, `Outcome`, `StartedAt`, and `CompletedAt`. |
| `TaskOutcome` | Named `string` with `unspecified`, `success`, `system failure`, and `validation failure` constants. |
| `DecisionRequest`, `DecisionResponse` | Signal payload structs used for a human decision. |

The parameters and results are ordinary structs serialized by Temporal. A Go
workflow is not required to use the convenience helpers, but it must accept and
return data compatible with the configured extension's contract. Importing
`pkg/childwf` keeps its types and constants aligned with Enduro.

### Version compatibility

`pkg/childwf` is versioned with the Enduro Go module, not as an independent
module. Pin an Enduro release in the custom project's `go.mod` and compare it
with the version deployed by the Enduro service. Before upgrading, review
changes to the parameter structs, result structs, signal payloads, and enum
values, then run the custom workflow's Temporal tests against the new version.
An added field may remain compatible with existing serialized data, while a
renamed field, changed type, or changed meaning may not be.

Temporal workflows that run for a long time also replay old histories with
current worker code. Keep workflow changes deterministic and follow the
[Temporal Go SDK versioning guidance] when a change would break replay.

### Tasks and result helpers

`childwf.NewTask` creates a task with an `unspecified` outcome and start time,
but does not append it to a result.
`(*Task).Complete` sets its completion time, outcome, and formatted message;
`(*Task).Succeed` is the success shorthand. `(*Task).IsSuccess` reports whether
the outcome is `success`.

`PreprocessingResult` and `PostStorageResult` each add three helpers:

* `(*PreprocessingResult).NewTask` and `(*PostStorageResult).NewTask` create a
  task and append it to the result.
* `ValidationError` sets `OutcomeContentError` and completes a task with a
  validation failure. It prefixes the message with `Content error:` and joins
  multiple message arguments with a blank line.
* `SystemError` sets `OutcomeSystemError` and completes a task with a system
  failure. It prefixes the message with `System error:` and joins multiple
  message arguments with a blank line.

The following fragment shows how to classify `activityErr` when an activity
cannot complete and `validationFailure` when it completes but finds a content
problem. In this example, the workflow can still build a trustworthy result,
so it reports either condition to Enduro in that result:

```go
result := &childwf.PreprocessingResult{
    Outcome:      childwf.OutcomeSuccess,
    RelativePath: params.RelativePath,
}
task := result.NewTask(workflow.Now(ctx), "Validate metadata")

// Run the activity here, assigning activityErr and validationFailure.
completedAt := workflow.Now(ctx)

switch {
case activityErr != nil:
    workflow.GetLogger(ctx).Error(
        "Metadata validation could not be completed",
        "error", activityErr,
    )
    result.SystemError(
        completedAt,
        task,
        "Metadata validation could not be completed.",
        "Try again or ask an administrator to investigate.",
    )
case validationFailure != "":
    result.ValidationError(
        completedAt,
        task,
        "Metadata does not meet the required policy.",
        validationFailure,
    )
default:
    task.Succeed(completedAt, "Metadata is valid")
}

return result, nil
```

Both error helpers complete the task and set the result outcome. `SystemError`
does not include the underlying Go error automatically, so log it and add any
safe details that an operator needs to the task message. Task messages are
visible through the Enduro API and dashboard, so do not include secrets. On
success, `Succeed` completes the task but does not change the result outcome.
The same pattern works with `PostStorageResult`.

Each error helper overwrites the current result outcome. Do not call
`ValidationError` after `SystemError`, because that would replace
`OutcomeSystemError` with `OutcomeContentError`. Return after a system error or
otherwise preserve the intended overall outcome.

Using these helpers is optional. A workflow may construct the result and tasks
directly, but should use deterministic workflow time, set the overall outcome,
and finish every task it returns.

When Enduro saves returned preprocessing or poststorage tasks, it maps child
task outcomes as follows:

| Child task outcome | Enduro task status |
| --- | --- |
| `TaskOutcomeUnspecified` | `unspecified` |
| `TaskOutcomeSuccess` | `done` |
| `TaskOutcomeSystemFailure` | `error` |
| `TaskOutcomeValidationFailure` | `failed` |

The returned task list is a completion log, not a live task channel. Enduro
receives and saves it only when the child workflow completes successfully from
Temporal's perspective and its result can be decoded.

### Custom metadata

Custom metadata values must be valid JSON. Decide which workflow owns each key
in the map to avoid conflicts with another custom workflow:

```go
result.CustomMetadata = childwf.CustomMetadata{
    "acme": json.RawMessage(`{"recordID":"A-123"}`),
}
```

Enduro treats each value as opaque. It does not validate its schema or merge
inside the JSON value. See each workflow page for when metadata is accepted,
merged, or omitted.

## Results and workflow errors

A Temporal child completes with either a completion result or a workflow
failure. When `ChildWorkflowFuture.Get` returns an error, it does not also
decode a usable completion result. Consequently, returning a populated Go
result together with a workflow error does not deliver that result to Enduro.
Enduro cannot then read the outcome, save returned tasks, or propagate custom
metadata.

`PreprocessingResult` and `PostStorageResult` do not have an error or message
field at the result's top level, so put failure details for operators in the
relevant `Task.Message`. `PostbatchResult` instead has a direct `Message` field
and no task list. See [Temporal Go child workflows] and [Temporal Go error
handling] for the underlying SDK behavior and error model.

For expected policy, validation, or processing failures where Enduro still
needs structured data:

1. Handle and classify the failure rather than discarding it.
2. Put an explanation for operators in the relevant task message or the
   postbatch result message.
3. Set the matching `Outcome` and task outcome.
4. Return the populated result with a nil workflow error.

This completion is successful only from Temporal's point of view. Enduro still
interprets preprocessing and poststorage content and system outcomes as
failures in SIP processing.

Return a workflow error for invalid arguments or violated invariants when no
trustworthy result can be produced.

## Build a custom workflow worker

Use the [custom workflow template README] to create, configure, build, and run a
Go worker with Enduro and Tilt. It covers the project layout, names to change,
worker and activity registration, configuration, build commands, container
image, and projects with more workflow examples.

See [temporal-activities] for reusable Go activities and their configuration.

[BagIt]: https://www.rfc-editor.org/rfc/rfc8493
[batch workflow]: https://github.com/artefactual-sdps/enduro/blob/main/internal/workflow/batch.go
[child workflow configuration]: ../../admin-manual/configuration.md#child-workflows
[childwf source]: https://github.com/artefactual-sdps/enduro/tree/main/pkg/childwf
[custom workflow template README]: https://github.com/artefactual-sdps/custom-enduro-workflows/blob/main/README.md
[postbatch workflows]: postbatch.md
[poststorage workflows]: poststorage.md
[preprocessing workflows]: preprocessing.md
[processing workflow]: https://github.com/artefactual-sdps/enduro/blob/main/internal/workflow/processing.go
[Temporal Activities]: https://docs.temporal.io/activities
[temporal-activities]: https://github.com/artefactual-sdps/temporal-activities
[Temporal Child Workflows]: https://docs.temporal.io/child-workflows
[Temporal Go SDK]: https://docs.temporal.io/develop/go
[Temporal Go child workflows]: https://docs.temporal.io/develop/go/workflows/child-workflows
[Temporal Go error handling]: https://docs.temporal.io/develop/go/best-practices/error-handling
[Temporal Go SDK versioning guidance]: https://docs.temporal.io/develop/go/workflows/versioning
[Temporal message passing]: https://docs.temporal.io/encyclopedia/workflow-message-passing
[Temporal SDKs]: https://docs.temporal.io/develop
[Temporal Task Queues]: https://docs.temporal.io/task-queue
[Temporal Workers]: https://docs.temporal.io/workers
[Temporal Workflows]: https://docs.temporal.io/workflows
