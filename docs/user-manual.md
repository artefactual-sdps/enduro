# User manual

This is a user manual for SDPS Enduro, a version of the Enduro project that
uses a3m to preserve digital objects.

## What is SDPS Enduro?

Enduro is a tool that was developed to automate the processing of transfers in 
multiple Archivematica pipelines. It uses [Minio](https://min.io/) for object
storage and [Temporal](https://temporal.io/) to manage the workflow.

The a3m project is a streamlined version of Archivematica that is wholly focused
on AIP creation. It does not have external dependencies, integration with access
sytems, search capabilities, or a graphical interface. It was designed to reduce
the bulk of Archivematica's extraneous functions for users operating at a large
scale, who are more focused on throughput of digital objects for preservation.

This version of Enduro uses a3m instead of Archivematica to preserve digital
objects. This combination of tools is intended to be lightweight, scalable, and 
easy to install.

## Installation and setup

To do

## Using Enduro + a3m

This documentation is a work in progress. Since a3m is a derivative of
Archivematica, you can refer to the [Archivematica
documentation](https://archivematica.org/docs/latest/) for much more detailed
explanations of specific facets of the system.

### Prepare digital objects

Digital objects and their metadata can be packaged in a few different ways for
upload to the system. a3m is format-agnostic, meaning that it can accept any
file that you pass to the system for processing. A single transfer can be
homogenous or it can be a mix of many different formats. In all cases, the
digital objects must be packaged together as either a `.zip`, `.tgz`, or
`.tar.gz`.

a3m reuses two of the transfer types from Archivematica - zipped directory and
zipped bag. a3m will automatically recognize the transfer type and adjust its
processing workflow accordingly.

**Zipped directory**: digital objects that have been packaged together using the
`.zip`, `.tgz`, or `.tar.gz` packaging format. When a zipped directory transfer
starts, the zip will be unpacked. The internal structure of a zipped directory
transfer can either be a loose collection of files, or it can include structures
like a metadata directory.

**Zipped bags**: digital objects that have been packaged according to the [BagIt
File Packaging Format](https://tools.ietf.org/html/rfc8493), colloquially known
as bags. Bags must be packaged together using the `.zip`, `.tgz`, or `.tar.gz`
packaging format. Archivematica will verify the bag early on in the transfer
process, looking at manifest information created during the bagging process such
as checksums and the payload oxum.

For more information on how a3m/Archivematica implement the BagIt specification,
please see [Unzipped and zipped
bags](https://www.archivematica.org/docs/latest/user-manual/transfer/bags/#bags)
in the Archivematica documentation.

### Upload to Minio

### View tasks in Temporal

### Retrieve AIPs