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

1. In Minio, navigate to the Object Browser and select your upload bucket. In
   this example, the upload bucket is called `sips`.

   ![The Object Browser page in Minio. The body of the page shows four buckets:
   aips, perma-aips1, perma-aips2, and sips.](screenshots/minio-buckets.jpeg)

2. Click on **Upload** and then select **Upload file**. This will open a file
   browser.

   ![The sips bucket page in Minio, with the Upload button circled in red. The
   bucket contains two transfers already.](screenshots/minio-upload.jpeg)

3. In the file browser, locate your transfer package and upload it to Minio.
   Once the progress bar has completed, Enduro will begin processing the transfer. 

### View tasks in Enduro

1. In Enduro, navigate to the Packages tab. The list of packages will show the
   most recent package first. You will also see the UUID of the package, when
   processing started, and the UUID of the location where the package is stored.
   The Status column will display one of five possible statuses:

   * **Done**: The current workflow or task has completed without errors.
   * **Error**: The current workflow has encountered an error it could not resolve
     and failed.
   * **In Progress**: The current workflow is still processing.
   * **Queued**: The current workflow is waiting for an available worker to begin.
   * **Pending**: The current workflow is awaiting a user decision.

   ![The Packages tab in Enduro. The body of the screen shows a table that lists
   all of the packages that have been processed by the Enduro
   instance.](screenshots/enduro-packages-tab.jpeg)

2. For more information about the package, click on the name of the package to
   access the package detail page.

   ![The package detail page in Enduro. The body of the screen shows a table that lists
   all of the packages that have been processed by the Enduro
   instance.](screenshots/enduro-package-detail.jpeg)

3. At the bottom of the package detail page, there is a list of **Preservation
   actions** undertaken on each package. Clicking on the arrow will open a list
   showing all the tasks that comprise the preservation action.

   ![alt](screenshots/enduro-preservation-actions-expand.jpeg)

## Download AIP

1. If your AIP has been successfully processed, the workflow status for the
   Create AIP Preservation Action should be set to Done. This is shown in two
   different places on the page - in the **AIP creation details** section of the
   main body of the page as well as under **Preservation actions** at the
   bottom.

   ![alt](screenshots/enduro-create-aip-done.jpeg)

2. You can download the AIP by clicking on **Download** in the **Package
   details** section.

### Move AIP

You can move packages to other storage locations that have been connected to the
Enduro instance. In this example, all of the storage locations are configured
through Minio.

1. On the package detail page in Enduro, select **Choose storage location**.

   ![alt](screenshots/enduro-choose-storage-location.jpeg)

2. All storage locations will be displayed in the pop-up window. Storage
   locations that are available will have a **Move** button to the right of the
   location name. If there is no Move button, the package is either already
   stored in that location or the location is available for some other reason.
   Select **Move** to move the package to your preferred location.

   ![alt](screenshots/enduro-available-storage-locations.jpeg)

3. An admonition will appear indicating that the package is being moved. You may
   need to refresh the page to see that the package has been successfully moved.

4. A new Preservation Action called **Move package** will appear at the bottom
   of the page. You can click on the arrow to see more information about the
   move.

   ![alt](screenshots/enduro-move-preservation-action.jpeg)

