# Managing ingest workflows

Everything you need to know to manage SIP ingests after they are uploaded

-----

## SIP view pages

The SIP view page provides a high-level overview of a given SIP and any ingest
workflows that have been and their outcome. Additionally, any
[related packages](#related-packages) derived from the SIP during processing are
linked as well.

![The SIP view page, showing a successfully ingested SIP](../screenshots/sip-details-ingested.png)

The page title will be the SIP name. The SIP UUID is used to construct the page
URL. The page is then organized into 3 main sections:

* [SIP details](#sip-details)
* [Related packages](#related-packages)
* [Ingest workflow details area](#workflows-and-activities)

### SIP details

SIP details are found in the body of the SIP view page. This section provides a
few high-level metadata elements about the SIP and its ingest, including:

* **Name**: Name of the SIP at ingest time. Should match the page title.
* **UUID**: The unique identifier associated with the SIP. Either extracted from
  the SIP metadata, or assigned by Enduro if no existing UUID is found.

!!! tip

    You can quickly copy the UUID by clicking the icon to the right of it

     ![The UUID of a SIP, showing the tooltip for copying it](../screenshots/uuid-copy.png)

* **Status**: The status of the SIP. Uses the same statuses as those shown on
  the SIP browse page.
* **Uploaded by**: The user associated with initiating the SIP ingest by
  uploading the package. If [authentication is enabled][iac], then depending on
  what user properties are available from the provider, Enduro will first try to
  show the user name. If a name is not available, an email address will be used
  instead, and if neither are provided then Enduro will use a locally generated
  UUID to uniquely identify the uploader. If authentication is not enabled
  and/or the identity of the uploader cannot be known (for example, ingest is
  started via a [watched location upload][watched-location]), Enduro will simply
  show "Unknown" in the  Uploaded by field.
* **Started**: Timestamp of when the ingest workflow started.
* **Completed**: Timestamp of when the ingest workflow ended. An estimate of the
  total time of the ingest will also be shown below in parentheses.

!!! tip

    Internally, Enduro will store timestamps in Coordinated Universal Time, i.e.
    [UTC](https://en.wikipedia.org/wiki/Coordinated_Universal_Time). However,
    the user interface will then render those timestamps based on your browser's
    or operating system's configured timezone settings.

### Related packages

The related packages widget, on the right side of the SIP view page, shows any
packages derived from the original upload during the ingest. This can include:

**Related AIPs**: If the ingest is successful, one or more AIPs will be created
following ingest, depending on the ingest workflow activities.

![The related packages widget, showing an AIP](../screenshots/related-pkg-aip.png)

You can click the "View" button to go to the related AIP view page, where the
AIP can be downloaded, deleted, and otherwise managed.

**Failed SIPs**: Alternatively, if the SIP encounters either a [content failure]
or a [system error] during the ingest workflow, a copy of the SIP at the time of
failure, with `Failed_SIP_` appended to its name, will be shown instead. You can
click the "Download" button to download a local copy of the failed SIP for
inspection and/or fixes before reattempting ingest.

![The related packages widget showing a Failed SIP](../screenshots/related-pkg-failed.png)

## Workflows and activities

The bottom half of the SIP view page contains the **Ingest workflow details**. A
card with summary information about the ingest workflow will be shown here,
including:

* Workflow name
* Workflow status
* Completed timestamp (with an estimated duration next to it)

![The Workflow details header, collapsed](../screenshots/workflow-details-collapsed.png)

Click anywhere on the header card to expand it and see more information about
[the tasks run](#workflow-tasks) as part of the workflow.

### Workflow task status legend

Both [workflows][workflow] and their component [tasks][task] have a controlled
vocabulary of **statuses** that can tell you more about the current state or
outcome of a given process.

Clicking the blue "( ? )" question mark icon next to the Ingest workflow details
header will reveal a legend explaining the various task statuses and their
meaning:

![The task status legend](../screenshots/task-status-legend.png)

**Workflow tasks** can have the following statuses:

* **DONE**: The task has completed successfully
* **FAILED**: The related package has failed to meet this task's policy-defined
  criteria
* **IN PROGRESS**: The task is still processing
* **PENDING**: The task is awaiting a user decision
* **ERROR**: The task has encountered a system error it could not resolve

**Workflows** have their own status as well. Most of these are similar to the
task statuses, with a few additional statuses:

* **QUEUED**: The workflow is waiting for an available worker to begin
* **CANCELED**: The workflow has been canceled by a user

#### Errors vs failures

To help operators better understand the cause of an unsuccessful workflow,
Enduro uses different statuses for  a [content failure] and a [system error].

When a task in an ingest workflow fails validation due to some element of the
submitted content (e.g. SIP structure, metadata, files, etc) not matching the
criteria defined in the workflow task, this is a **content failure**, and the
related task will be given a status of: **FAILED**. When the workflow finishes
running as far as it can, it will then be given the same failed status.

These issues are generally ones that can be fixed by the original producer or by
an Enduro operator, as they relate to the contents and structure of the SIP, and
not the system itself. Producers and/or operators can then choose to:

* Download the failed package from the [Related packages](#related-packages)
  widget
* Use the details shown in the related
  [workflow task](#workflow-tasks) to better understand the issue
* Identify and fix the issue in the SIP
* Resubmit the SIP for ingest

Conversely, Enduro will use an **ERROR** status when a **system error**
interrupts one or more ingest tasks, causing the workflow to halt. This might be
due to network interruptions, disk space issues, hardware malfunctions, or
software bugs - generally, a system administrator will be needed to resolve the
issue upstream before ingest can be tried again.

#### Pending tasks and workflows

A **PENDING** task or workflow means that all workflow activity is paused,
**waiting for input** from an operator before proceeding.

In such cases, buttons allowing an operator to input a decision are generally
provided and the workflow remains paused until input is received. For example, a
package deletion request initiated by an operator might then show "Approve" and
"Deny" buttons in the workflow details header.

### Workflow tasks

A workflow is a sequence of tasks managed by Enduro. The **Ingest workflow
details** area will list all workflows that have been run against a given
package in ascending order, with the most recent on the top. Typically, most SIP
ingest pages will only include 1 workflow.

Click anywhere on the **workflow header card** to expand it and see a list of
all tasks run as part of that workflow. Tasks are also shown in ascending order,
with the most recent tasks at the top of the list.

![A workflow header card expanded to show the task list below](../screenshots/workflow-details-expanded.png)

Tasks shown in this area will include both those ingest tasks performed by
Enduro, as well as tasks run by the configured [preservation engine] if the SIP
passes initial validation and transformation.

Task cards will include:

* A **task number** assigned by Enduro, indicating the order the task was run in
  the workflow
* The **task name** in bold, helping to explain what activity is being performed
* A **status** - see [above](#workflow-task-status-legend) for details on each
  task status meaning
* A **timestamp** - if the task has completed, this will list the completed
  timestamp. If the task is still running or if it does not complete
  successfully (i.e. a failure or error), it will show a timestamp of when the
  task started running

Additionally, those ingest tasks run by Enduro will include an additional
description of the **task outcome**:

![Task detail cards with a successful outcome](../screenshots/task-details-success.png)

### Errors and failed package downloads

If an ingest task **fails** or encounters an **error**, Enduro will attempt to
continue running any remaining validation tasks to gather as much information
about the SIP as possible, but will terminate the workflow before transforming
the package and delivering it to the [preservation engine].

The **task details** will then provide operators with additional context on the
problem encountered.

![Task details card with a failed outcome](../screenshots/task-details-failure.png)

If desired, you can then download the SIP from the [Related packages
widget](#related-packages) to inspect it.

[content failure]: ../glossary.md#content-failure
[iac]: ../../admin-manual/iac.md
[preservation engine]: ../glossary.md#preservation-engine
[system error]: ../glossary.md#system-error
[task]: ../glossary.md#task
[watched-location]: submitting-content.md#initiate-ingest-via-a-watched-location-upload
[workflow]: ../glossary.md#workflow
