# Glossary

This glossary of terms outlines the domain-specific language used when
discussing the Enduro functionality described throughout the User Manual, as it
relates to ingest and digital preservation. While some technical terms are
included, the definitions provided here may not always match exactly how
these terms are used in the related technologies used by Enduro.

If other glossary terms appear in the definition of a term, they will be
**bolded** the first time they are used, so readers are aware that a related
definition is available.

-----

## Activity

See [Task](#task).

## Agent

An actor (human, machine, or software) associated with one or more **events**
in a preservation **workflow**.

## Archival Information Package (AIP)

A type of **package** produced as the output of the **preservation engine**,
consisting of one or more original **objects** deemed worthy of preservation,
any preservation **derivatives** created during the **workflow** based on the
current format policies of the system, and all associated **metadata**, so that
the package can be understood and its objects properly rendered by the intended
designated community of future users.

The AIPs that Enduro produces are self-describing, system-agnostic, and
standardized based on open formats and standards for long-term preservation.
Metadata packaged in an AIP will include any metadata added to the original
**SIP**, as well as any technical, administrative, structural, and preservation
metadata generated by Enduro as part of the preservation workflow. Depending on
the **preservation policies** defined, one or more AIPs are derived from a
**Processing Information Package (PIP)** that is transformed as part of a
preservation **workflow**. AIPs are sent to the **Preservation Storage Service**
for long-term storage and preservation.

The term Archival Information Package was originally defined in the
[OAIS](https://en.wikipedia.org/wiki/Open_Archival_Information_System) Reference
Model created by the Consultative Committee for Space Data Systems (and since
formally recognized as an international standard in ISO 14721). A detailed
description of the AIP and its components is available in the Archivematica
documentation - see: [AIP
structure](https://www.archivematica.org/docs/latest/user-manual/archival-storage/aip-structure/#aip-structure).

## Child workflow

An ancillary **workflow** that is spawned by another workflow. This capability
allows breaking down complex **tasks** into manageable units, improving
maintainability, extensibility and scaling operations.

In Enduro, we recommend using child workflows to manage user-specific tasks and
**preservation actions** in a workflow, such as **SIP** validation against a set
of institution-specific requirements. This keeps locally specific
implementations nicely separated from more general, reusable workflow activities
available in Enduro Ingest.

## Content failure

An failure in an **ingest** **workflow** that occurs during **SIP** validation,
related to the submitted package's structure, files, or metadata, that must be
resolved by the original package submitter or another operator. See also:
[System error](#system-error).

## Derivative

A derived version of an **object**, typically in a different format and intended
for a distinct use, that may be created during a preservation **workflow**,
depending on the configured **preservation policies**. The two most common types
of derivatives used in Enduro are:

* **Preservation derivative**: a derived version of a preservation object using
  **file formats** deemed more suitable for long-term preservation, based on
  factors such as: open format specifications, widespread adoption and support,
  lossless algorithms if compression is used, de jure or de facto
  standardization, etc. Preservation derivatives are often included in **AIPs**
  alongside original objects to increase the likelihood of successful future
  access, in case the original objects use proprietary and/or rare formats that
  may not be easily rendered in the future.
* **Access derivative**: a derived version of a preservation object using file
  formats deemed more suitable for dissemination and access across space and
  time (particularly over networked environments), based on factors such as:
  open format specifications, smaller file sizes, widespread adoption and
  support (particularly in web browsers), high compression rates, etc. Access
  derivatives are typically used in the creation of **Dissemination Information
  Packages (DIPs)**.

## Directory

A conceptual **object** representing a collection of **files**, used to provide
human-understandable groupings of related digital objects in a filesystem.
Directories may also contain subdirectories, and are typically organized
hierarchically. Often used in the organization of files in a **package**, such
as a **SIP** or **AIP**. Sometimes colloquially referred to as a "folder."

## Dissemination Information Package (DIP)

A type of **package** derived from one or more **AIPs** for delivery to an end
user in response to an authorized access request. A DIP will typically include
**derivative** versions of the original preservation **objects** using different
**file formats** more suitable to access environments (i.e. lower resolutions,
more commonly supported by web browsers, etc).

The term Dissemination Information Package was originally defined in the
[OAIS](https://en.wikipedia.org/wiki/Open_Archival_Information_System) Reference
Model created by the Consultative Committee for Space Data Systems (and since
formally recognized as an international standard in ISO 14721).

## Document

An **object** consisting of information or data fixed to some media.

## Event

A preservation-relevant action that involves at least one **object** and/or
**agent**. Events are typically captured in the preservation **metadata** of a
**package** during a preservation **workflow** using the
[PREMIS](https://www.loc.gov/standards/premis/) metadata specification, so that
an archival chain of custody is preserved across all **tasks** that might occur
in a workflow.

## File

A named and ordered sequence of binary data, organized as a content bitstream
that can be stored and transmitted, and typically bearing a specific **format**.

## File format

A standardized method of specifying how bits are
used to encode information in a digital storage medium such as a **file**. A
format's associated specification may be proprietary or open.

## Ingest

A phase in a preservation **workflow** describing all the **preservation
policy**-defined **tasks** performed on a **SIP** when it is received from a
**producer**, prior to preservation. Ingest task examples include checking the
file format against a list of acceptable formats and checking for hidden files.
Typically, ingest concludes by finalizing the transformation of a SIP into a
**PIP** for preservation processing by the **preservation engine**.

## Intellectual entity

The [PREMIS v3 Data Dictionary](https://www.loc.gov/standards/premis/) defines
an Intellectual Entity as any "*Coherent set of content that is described as a
unit: for example, a book, a map, a photograph, a serial. An Intellectual Entity
can include other Intellectual Entities; for example, a Web Site can include a
Web Page, a Web Page can include a photograph.*" ([footnote-1](#1)). A
**package**, being a grouping of **files** and related **metadata** described as
a unit, is an example of an intellectual entity that may contain additional
intellectual entities - or alternatively, one package may represent a subset of
a larger intellectual entity spread across multiple packages.

## Metadata

Structured information that characterizes and contextualizes another data source
(such as an **object** or **package**), especially for the purposes of
documenting, describing, preserving, and/or managing that resource.

There are many types of metadata, such as:

* **Descriptive metadata**: Information that supports the identification and
  discovery of a resource, describing the contents and their relationships to
  other resources and entities. For example, the title of a document.
* **Administrative metadata**: Information that supports the internal management
  and governance of a resource, capturing its provenance, any related rights
  statements or limitations, selection criteria, and/or local policies or
  agreements impacting its use, acquisition, or disposition. For example, the
  date a record was acquired by an archive.
* **Preservation metadata**: Information necessary to manage and preserve a
  resource over time including the documentation of any preservation actions
  taken to ensure the authenticity of the resource, and the ongoing efforts to
  ensure its preservation and usability over time, across space, and through
  technological change. For example, documenting the process of producing a
  checksum of a file when receiving it from an agency.
* **Technical metadata**: Information that describes the technical properties
  and characteristics of a resource required to render or process it. For
  example, the bit depth of an audio **file**.
* **Structural metadata**: Information that documents the relationship within
  and among a resource, enabling end users to understand its structure and how
  to navigate complex **objects**. For example, the ordering of individual TIFF
  scans making up a book object.

## Normalization

A preservation processing **task** that creates a copy of a **file** in a
different **format** based on a **preservation policy**, as part of a strategy
to support long-term preservation and/or access.

When normalization occurs for preservation (i.e. during **AIP** creation), the
goal is generally to avoid technical obsolescence, and target formats are
typically chosen based on open licenses and specifications, lossless
compression, ubiquity of support and use, and community consensus on best
practice. When normalization is performed for access (i.e. during **DIP**
creation), target formats are typically chosen based on file size and
compression levels, ease of transmission, and ubiquity of support and adoption
(particularly on the public web).

## Object

An **intellectual entity** that describes something of enduring value, and under
consideration for preservation. There is at least one object in every
**package**, and each object can contain zero or more **files**. An object may
be described by a **producer** using any desired descriptive **metadata** that
is meaningful to the designated community.

## Package

An **intellectual entity** describing a type of collection, composed of a set of
**files** and related **metadata**, assembled together for a particular purpose.
A package will contain one or more **objects**, zero or more **files**, and
**metadata**. Types of packages used in Enduro include: **SIPs**, **PIPs**,
**AIPs**, and **DIPs**.

## Pipeline

An instance of a **preservation engine** configured with specific **preservation
policies**, through which **packages** pass as part of a **workflow**.

## Post-ingest

Post-ingest is a phase in a **preservation workflow** describing all the
**preservation policy**-defined tasks performed on an **AIP** following
preservation processing. Post-ingest task examples include: compression and/or
encryption of AIPs, sending any **metadata** to an external system (such as an
archival management system), **DIP** generation and delivery, geo-redundant
**replication**, as well as the transfer of the AIP to its final storage
location.

## Pre-ingest

Any manual examination, review, and/or additional preparation steps performed on
**SIPs** received from **producers** prior to **ingest**. Can include activities
such as appraisal, selection,and SIP arrangement.

## Preservation action

A **workflow** component composed of one or more **tasks** performed on a
**package** to support preservation, informed by the preservation policies
configured in the preservation system.

## Preservation engine

The component of Enduro that manages the transformation of **PIPs** into
**AIPs** in a **workflow** during preservation processing. See:
[Components](components.md).

## Preservation policy

A policy or business rule defining the target inputs and/or outcomes of a
preservation action or **workflow** **task**, and any supporting contextual
information. For example, setting a policy in the **preservation engine** to
**normalize** a JPEG image file to a JPEG2000 during **AIP** creation.

## Preservation processing

A phase in a preservation **workflow** describing all the **tasks** that occur
after **ingest**, when a **PIP** is sent to the **preservation engine** for
transformation into one or more **AIPs**. Prior to final bagging, compression,
and storage, any AIPs created during ingest may also undergo any additional
post-ingest activities defined in the system. Examples of preservation
processing workflow steps include file format characterization, and preservation
**normalization**.

## Preservation storage service

A component of Enduro that acts as an interface between the **preservation
engine** and whatever storage devices have been configured for the long-term
preservation of **AIPs** produced by Enduro **workflows**. See:
[Components](components.md).

## Processing Information Package (PIP)

A type of **package** derived from a **SIP** following all **ingest**
activities, intended to be sent to the **preservation engine** for
transformation into one or more **Archival Information Packages (AIPs)** for
long-term preservation.

In Enduro, a PIP is a transitional package state in the preservation
**workflow** intended to standardize inputs to the preservation engine, and not
a package type that Enduro operators will typically interact with directly.

## Processing storage service

A component of Enduro that acts as Enduro's storage back-end for any local
**package** interactions that are needed as part of **ingest** and preservation
**workflows**. See: [Components](components.md).

## Producer

Persons and/or organizations responsible for the submission of **packages**
(**SIPs**) to be considered for long-term preservation by the preservation
system and its operators.

## Replication

A **post-preservation** activity to produce duplicate copies (i.e. replicas) of
a **package** such as an **AIP**, often in separate geo-redundant locations from
the original, to mitigate risk and assist in recovery should any issue be found
with the original package throughout its ongoing preservation.

## Submission Information Package (SIP)

A type of **package** submitted by a **producer** to be considered for long-term
preservation. A SIP is also sometimes called a **transfer**.

The term Submission Information Package was originally defined in the
[OAIS](https://en.wikipedia.org/wiki/Open_Archival_Information_System) Reference
Model created by the Consultative Committee for Space Data Systems (and since
formally recognized as an international standard in ISO 14721).

## System error

A technical error in a **workflow** that is caused by a problem in one or more
of the applications, networks, or connecting interfaces, typically requiring a
system administrator or developer to resolve. See also:
[Content failure](#content-failure).

## Task

An operation performed on a **file** or **package** in the context of a
**workflow**. Also called an **activity**.

## Transfer

See: [Submission Information Package (SIP)](#submission-information-package-sip)

## Watched directory

A filesystem **directory** used that is monitored for changes (e.g. adding,
deleting, or renaming a **file**), and where such a change can trigger one or
more subsequent actions.

For example, a watched directory may be used to automate the beginning of a
preservation **workflow** whenever a ZIP file is placed in the watched
directory. Alternatively, for multi-file uploads, a watcher may wait for a
specific filename or extension to be present before grabbing all content in the
directory for processing.

## Workflow

A sequence of **tasks** and/or preservation actions managed by Enduro.

Additionally, in the **workflow engine**, a workflow is a long-running, stateful
process that orchestrates the execution of tasks over time.

## Workflow engine

A component of Enduro that handles preservation processing as part of a
**workflow**, typically taking **Processing Information Packages(PIPs)** as
input, and outputting one or more **Archival Information Packages (AIPs)** for
long-term preservation. See: [Components](components.md).

-----

## Footnotes

### 1

PREMIS Editorial Committee - "PREMIS Data Dictionary for Preservation Metadata,"
Version 3.0. June 2015. Glossary, p. 270.
