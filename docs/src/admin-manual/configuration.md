# Configuration

This page describes the various configuration files and settings that Enduro
supports.

!!! important

    Some of the information in Enduro's configuration files includes connection
    information that should not be publicly exposed. To increase the security of
    your deployment, you can encrypt the configuration files.

    Additionally, it is possible to set any / all configuration value(s) using
    [environment variables](https://linuxize.com/post/how-to-set-and-list-environment-variables-in-linux/).
    The naming convention rules are as follows:

    * Each variable must begin with `ENDURO_`
    * Use all upppercase
    * Change any periods in the configuration section headers to underscores,
      and do the same for any parameters within a section

    For example, to create an environment variable of the `claimPath` parameter
    in the `[api.auth.oidc.abac]` configuration section, its name should be:

    * `ENDURO_API_AUTH_OIDC_ABAC_CLAIMPATH`

## The enduro.toml file

The `enduro.toml` is Enduro's primary configuration file. This file defines
communication channels and methods with the various applications used by Enduro,
sets filesystem permissions for packages Enduro will interact with during
workflows, establishes storage locations, and more.

At installation time, the `enduro.toml` file can be found in the root directory
of the project.

!!! tip

    For production bare-metal server installations, this configuration file can
    be moved during setup. When starting the Enduro workers after installation,
    its new location can be passed as a parameter, using `--config`. For
    example:

    ```bash
    go run cmd/enduro/main.go --config=path/to/enduro.toml
    ```

In many cases, the TOML file's settings blocks will correspond to one of the
packages found in Enduro's `internal` directory. In such cases, the package will
contain a `config.go` file that defines the variables to be included in the
config file. At run time, Enduro will then use the values added to this primary
configuration file when executing activities that use the referenced packages.

For example, the [database connection settings](#database-connection) described
below correspond to the `/internal/db` package, with the configuration values
listed here defined in `/internal/db/config.go`.

Not all settings listed below correspond directly to a package, and not all
packages in the `internal` directory have configurable settings found here.

!!! tip

    Note that some [child workflow](../dev-manual/preprocessing.md) activities
    may have their own configuration files.

    For example, see the
    [ffvalidate](https://github.com/artefactual-sdps/temporal-activities/tree/main/ffvalidate)
    activity in the temporal-activities repository, which requires an additional
    CSV configuration file listing an allowed list of file-formats in a SIP when
    used for ingest validation.

### Debug mode for development

The application log captures a record of discrete Enduro events to aid in
diagnosing system errors and bugs. These settings control the format and
verbosity of the logging information emitted by Enduro, which can be useful to
increase during development.

!!! tip

    For additional information on logging and development, see the following
    pages in the Developer's Manual:

    * [Logging](../dev-manual/logging.md)
    * [Local/Development environment](../dev-manual/devel.md)

**Default values**:

```toml
debug = true
debugListen = "127.0.0.1:9001"
verbosity = 2
```

* `debug`: Enables or disables debug mode. Accepted values are `true` or
  `false`. When set to true, Enduro will log to the console using a
  human-readable text output, and key elements are colorized (e.g. using red to
  make ERROR more visible, etc). When set to false, logs are captured in JSON
  format to support machine parsing and analysis.
* `debugListen`: the IP and port to use if configuring an observability server.
* `verbosity`: Sets the verbosity of the log output. Note that errors are
  **always** logged - this setting will control the verbosity of other log
  events. Values can range from 0-128, with 0 being the least verbose and
  capturing only the most important INFO messages. Set this value to 128 to
  ensure that _all_ possible events are captured in their most verbose form.

!!! note

    Artefactual developers have not tested or used the observability server
    configuration - it may not work as expected out of the box.

    At present, the highest configured log verbosity value used is around 3.

### Temporal configuration

[Temporal] is Enduro's default [workflow engine], and is required for Enduro to
run ingest, preservation, and storage workflows. Currently, no other workflow
engines can be substituted. These configuration settings allow Enduro to connect
and communicate with Temporal.

**Default values**:

```toml
[temporal]
namespace = "default"
address = "temporal.enduro-sdps:7233"
taskQueue = "global"
```

* `namespace`: An internal namespace label for the Temporal workflows. This only
  needs to be changed if you expect to be running multiple different workflows
  running at once, so they don't clash.
* `address`: Tells Enduro where to find Temporal - address and port connection
  information.
* `taskQueue`: In Temporal, a [Task Queue] is a lightweight, dynamically
  allocated queue that one or more workers can poll for tasks. Temporal supports
  many kinds of parallelization for a distributed application architecture, and
  and task queues can be namespaced to support such a set-up. At present
  however, Enduro will only support the default `global` value for the general
  ingest workflow queue.

### Internal API

Configuration information for Enduro's internal API calls across different
[components]. Unlike the other API configuration section [below](#api), the
internal API does not use or support authentication. As such, the configured
port should **not be exposed publicly**.

**Default values**:

```toml
[internalapi]
listen = "0.0.0.0:9002"
debug = false
```

* `listen`: tells Enduro what IP address and port to use for internal API
  communication across components.
* `debug`: Enables (set to `true`) or disables (set to `false`) debug mode for
  the internal API. When enabled, debug prints detailed information about
  incoming requests and outgoing responses, including all headers, parameters,
  and bodies.

### API

The next set of four configuration blocks all relate to Enduro's primary
external facing [API]. This first block sets the basic configuration of the API.

!!! tip

    See also:

    * [Enduro API documentation](../dev-manual/api.md)

**Default values**:

```toml
[api]
listen = "0.0.0.0:9000"
debug = false
corsOrigin = "http://localhost"
```

* `listen`: Specifies the address and port the API will bind to for
  communication.
* `debug`: Enables (set to `true`) or disables (set to `false`) debug mode for
  the API. When enabled, debug prints detailed information about incoming
  requests and outgoing responses, including all headers, parameters,
  and bodies.
* `corsOrigin`: [CORS] (short for Cross-Origin Resource Sharing) is a security
  mechanism implemented by web browsers that controls how web pages can access
  resources from different domains, ports, or protocols. This setting defines a
  policy for Enduro to use the specified value as the primary origin domain.

#### Enable API authentication

This setting determines whether or not authentication is enabled for Enduro's
external-facing API. Currently, OpenID Connect ([OIDC]) is the only protocol
supported for authentication. When enabled, the API will check for and verify
the access token submitted with each request. The API client is responsible for
obtaining an access token from the provider.

!!! tip

    See also:

    * [Identity and access control](iac.md)

**Default value**:

```toml
[api.auth]
enabled = true
```

#### OIDC authentication provider configuration

This setting block is used to configure the OpenID Connect ([OIDC]) provider
when API authentication is [enabled](#enable-api-authentication). The default
value assumes that [Keycloak] will be used for handling OIDC single sign-on
requests.

For more details on OIDC configuration, consult the [OIDC specification].

**Default values**:

```toml
[api.auth.oidc]
providerURL = "http://keycloak:7470/realms/artefactual"
clientID = "enduro"
skipEmailVerifiedCheck = false
```

* `providerURL`: Defines the OIDC provider URL. This parameter is required when
  API authentication is enabled.
* `clientID`: Defines the OIDC client ID. The client ID must be included in the
  intended audience (`aud`) claim of the submitted access token. Also required
  when API authentication is enabled.
* `skipEmailVerifiedCheck`: One of the standard claim elements that can be
  requested and/or returned in an OIDC ID token or UserInfo response is the
  `email_verified` claim - When this claim value is true, this means that the
  OIDC provider "took affirmative steps to ensure that this e-mail address was
  controlled by the End-User at the time the verification was performed." When
  `skipEmailVerifiedCheck` is set to **false**, Enduro will check any submitted
  tokens or UserInfo responses for the verified email claim. When set to
  **true**, this check is skipped.

#### Enable Attribute Based Access Control for the API OIDC authentication

Enduro uses Attribute Based Access Control ([ABAC]) to manage permissions and
access. When ABAC is enabled for the API, it will check a configurable
multivalue claim against the defined required attributes based on each
endpoint's configuration.

**Default values**:

```toml
[api.auth.oidc.abac]
enabled = true
claimPath = "enduro"
claimPathSeparator = ""
claimValuePrefix = ""
useRoles = false
rolesMapping =
```

* `enabled`: Set to `true` to enable ABAC, or `false` to disable ABAC.
* `claimPath`: The claim path of the Enduro attributes within the access token.
  If the claim path is nested, include all fields separated by the
  `claimPathSeparator` defined below. This element is required when `enabled` is
  set to "true."
* `claimPathSeparator`: The separator used to split nested claim path fields.
  The default empty value of "" will try to match the claim path as-is to a
  top-level field from the access token.
* `claimValuePrefix`: Used to add a prefix to filter the values of the
  configured claim. If the claim contains values unrelated to Enduro's ABAC, the
  values relevant to Enduro should be prefixed so they are the only values used
  for access control.

    For example, a claim with values `["enduro:*", "unrelated"]` will be
    filtered to a value of `["*"]` when `claimValuePrefix = "enduro:"`.

    The default empty value of "" will not filter any value.

* `useRoles`: To simplify access control management for smaller organizations or
  those not requiring fine-grained access control, it is possible to predefine
  a set of roles with specific attributes, to which users can be assigned. Set
  this value to `true` to switch to role-based access control, and define the
  attributes associated with each role in the `rolesMapping` element below. This
  value is set to `false` by default.
* `rolesMapping`: If `useRoles` is set to true, this element is where you can
  specify a JSON-formatted string mapping each role you'd like to define to a
  set of attributes. The expected format is as follows:

    ```json
    {
      "role1": ["attribute1", "attribute2"],
      "role2": ["attribute1", "attribute2", "attribute3", "attribute4"]
    }
    ```

    For example:

    ```toml
    rolesMapping = '{"admin": ["*"], "operator": ["ingest:sips:list", "ingest:sips:read", "ingest:sips:upload", "ingest:sips:workflows:list"], "readonly": ["ingest:sips:list", "ingest:sips:read", "ingest:sips:workflows:list"]}'
    ```

    Because `useRoles` is set to false by default, `rolesMapping` is empty by
    default.

#### Redis event messenger authentication for the API

Enduro uses [Redis] as a watcher and messaging queue - see the [Components]
documentation for more information. These settings provide Redis with an
authorization ticket so it can access the API.

**Default values**:

```toml
[api.auth.ticket.redis]
address = "redis://redis.enduro-sdps:6379"
prefix = "enduro"
```

* `address`: Binds Redis to a specific address and port.
* `prefix`: Defines a namespace that can be used as a prefix to differentiate
  instances if you have multiple Enduro installations running. Otherwise, there
  is no reason to change the default "enduro" value.

### Database connection

These settings configure the connection information for Enduro's MySQL database.
Note that other MySQL distributions (like MariaDB or Percona) can be used, but
other database types are not currently supported in Enduro.

**Default values**:

```toml
[database]
driver = "mysql"
dsn = "enduro:enduro123@tcp(mysql.enduro-sdps:3306)/enduro"
migrate = true
```

* `driver`: Defines the database type. At present, `mysql` is the only supported
  value, even if other MySQL distributions are being used.
* `dsn`: Data Source Name. Specifies the database connection information,
  including username, password, Transmission Control Protocol (TCP) information
  for host-to-host communication, and database name.
* `migrate`: Determines whether database schema migrations are run when the
  Enduro worker starts. Set to `true` (enabled) by default.

    This _could_ be set to `false` (disabled) in a production database to
    protect against the risk of data loss from faulty migrations, but an
    administrator would need to remember to **re-enable the setting prior to any
    application upgrades**, to ensure that the database schema matches any new
    application changes or additions post-upgrade.

!!! tip

    For more information on database migrations, see:

    * [Database migrations](../dev-manual/db-migrations.md)

### Event queue

Enduro uses [Redis] as a watcher and event queue - see the [Components]
documentation for more information. At this time, Redis is the only supported
event queue for Enduro. The settings below configure Enduro's ability to connect
and communicate with Redis.

**Default values**:

```toml
[event]
redisAddress = "redis://redis.enduro-sdps:6379"
redisChannel = "enduro-ingest-events"
```

* `redisAddress`: Binds Redis to a specific address and port.
* `redisChannel`: Redis can be configured to use different channels for
  different queues. In this example configuration, we use same Redis
  installation and address for multiple different event listeners (see also
  [Storage event listener](#storage-event-listener) below for example), but
  different channels for each to avoid messaging conflicts.

    If you have multiple Enduro instances each talking to different event
    queues, then you would need to uniquely namespace each to avoid conflicts,
    using this parameter. Otherwise, there is no need to change the default
    value.

### Extract activity unix permissions

These settings define the POSIX filesystem permissions that are applied to
extracted files and directories when a compressed package is received and the
internal extract activity is executed. For security purposes to protect the
integrity of the received content, the default permissions are somewhat more
restrictive than the defaults used by many Linux systems.

**Default values**:

```toml
[extractActivity]
dirMode = "0o700"
fileMode = "0o600"
```

* `dirMode`: Defines the permissions assigned to directories in extracted
  packages. The default value is `0o700`, which ensures that only the owner (in
  this case, Enduro) and the `sudo` superuser can can view, access, or modify
  the directory and its contents, and that anyone else is locked out.
* `fileMode`: Defines permissions assigned to files in extracted packages.The
  default value is `0o600`, which grants the owner (now Enduro) permissions to
  read and write to the files, but not execute them.

### PREMIS validation activity

These settings control whether or not an activity runs during the ingest
workflow to check for and validate any PREMIS files found in the SIP.

[PREMIS] is a widely used open standard to record information about digital
preservation events, objects, agents, and rights. Enduro's supported
preservation engines [a3m] and [Archivematica] can both parse valid PREMIS.xml
files and add the contents to an AIP's [METS] file during preservation
processing.

When this activity is enabled, Enduro will check for the presence of a file
named `premis.xml` in the `metadata` directory of a SIP. If a matching file
is found in the expected location, Enduro will then run the [xmlvalidate]
activity and validate the premis.xml file against the schema file found at
the configured path.

!!! note

    For the PREMIS validation activity to succeed when enabled:

    * The [libxml2](https://en.wikipedia.org/wiki/Libxml2) library must be
      installed. Note that the project's Docker configuration automatically
      installs this dependency in the Enduro Docker image.
    * The file **MUST** be named `premis.xml`
    * The file **MUST** be found in a `metadata` directory
    * The metadata directory **must not** be nested inside any other directories
    * The PREMIS schema file path must be accessible to Enduro

**Default values**:

```toml
[validatePremis]
enabled = true
xsdPath = "/home/enduro/premis.xsd"
```

* `enabled`: Set to `true` by default. To skip PREMIS validation, change this
  value to `false`.
* `xsdPath`: Defines the filesystem path where Enduro can find the PREMIS 3
  XML Schema Definition file to be used for validation.

    Enduro does include a PREMIS 3.0 XSD file at installation, located at
    `hack/xsd/premis.xsd` from the root Enduro installation directory.

### Watched location configuration

These configuration settings, when enabled, allow Enduro to initiate SIP ingest
from a watched location. The configured watched location can be an object store
bucket such as one provided by MinIO, S3, or Azure. Once this section is
configured, the chosen watched location should then be configured to publish
an event to Enduro's message queue (in this case, [Redis], listening for events
in the queue defined by the `redisList` parameter below) any time a new zipped
package is added to the watched location. Enduro's internal watcher will watch
for new deposit events at the configured `redisAddress` and queue, and will
trigger the ingest workflow when it detects a SIP deposit in the configured
watched location. For more information on ingests from a watched location, see:

* [Initiate ingest via a watched location upload][watched-location]

At this time, [Redis] is the only supported messaging queue - as such, the
`redisAddress` and `redisList` fields are required.

All other default parameters are S3-specific, and might not be needed for other
watched location types. The default configuration included uses a [MinIO] bucket
as the example watched location. MinIO uses Amazon [S3] syntax for its
configuration properties. Different object stores may have different parameters
to be configured. Consult the corresponding object store provider's
documentation for more information.

**Default values**:

```toml
[watcher.embedded]
name = "dev-minio"
redisAddress = "redis://redis.enduro-sdps:6379"
redisList = "minio-events"
endpoint = "http://minio.enduro-sdps:9000"
pathStyle = true
key = "minio"
secret = "minio123"
region = "us-west-1"
bucket = "sips"
workflowType = "create aip"
```

* `name`: Defines a name to be used internally for the watched location. Useful
  if you are configuring more than one watched location.
* `redisAddress`: Binds Redis to a specific address and port.
* `redisList`: The name of the queue that Redis should use for the watched
  location SIP deposit events.
* `endpoint`: API endpoint for the target MinIO instance. Used by Enduro to read
  contents of a watched bucket, or for any other MinIO interactions.
* `pathStyle`: Currently Amazon Web Services support two different URL
  construction methods when interacting with an object store bucket via API. The
  "path-style" method constructs the bucket's access URL using the configured
  properties, such as region, bucket name, and object key. For Enduro
  integrations, the second virtual "host-style" method is not currently
  supported so if using an S3-like object store, ensure this is set to `true`.
* `key`: Username for accessing the configured S3-like bucket.
* `secret`: Password for accessing the configured S3-like bucket.
* `region`:  = AWS S3 buckets are created in a specific region. When interacting
  with S3, you can specify the region during the bucket creation process. For
  a full list of available regions and the syntax to specify them, consult the
  [AWS S3 documentation][S3-regions].
* `bucket`: A configured object store may have more than 1 bucket. This
  parameter specifies the target bucket name to be used for the watched
  location.
* `workflowType`: Specifies the name of the Enduro workflow type to be run when
  SIPs are deposited in the watched location. Currently the only supported
  values are "create aip" and "create and review aip". The latter review
  workflow also only works if [a3m] is the configured [preservation engine].

### Storage endpoint

This element configures the Enduro storage service API endpoint. Even when using
[Archivematica] as the [preservation engine] (which includes the AM
[Storage Service] as the AIP store), this endpoint will act as a proxy to the
AMSS. Otherwise, this configures the primary local storage service when using
[a3m].

**Default values**:

```toml
[storage]
enduroAddress = "enduro.enduro-sdps:9002"
defaultPermanentLocationId = "f2cc963f-c14d-4eaa-b950-bd207189a1f1"
```

* `enduroAddress`: Defines the address and port for the storage API endpoint.
* `defaultPermanentLocationId`: The UUID of the storage location used for
  permanent AIP storage in automated workflows. The default value provided
  represents the first permanent location defined in the
  `mysql-create-locations-job.yaml` Kubernetes manifest.

#### Storage database

Even when using the Archivematica [Storage Service] (AMSS), Enduro still
captures and stores some storage metadata for display in the user interface. The
following elements define the database used to store this information and how
Enduro can connect with it for read and write operations.

Enduro uses a MySQL database by default. Note that other MySQL distributions
(like MariaDB or Percona) can be used, but other database types are not
currently supported in Enduro.

**Default values**:

```toml
[storage.database]
driver = "mysql"
dsn = "enduro:enduro123@tcp(mysql.enduro-sdps:3306)/enduro_storage"
migrate = true
```

* `driver`: Defines the database type. At present, `mysql` is the only supported
  value, even if other MySQL distributions are being used.
* `dsn`: Data Source Name. Specifies the database connection information,
  including username, password, Transmission Control Protocol (TCP) information
  for host-to-host communication, and database name.
* `migrate`: Determines whether database schema migrations are run on Enduro
  start up. Set to `true` (enabled) by default.

    This _could_ be set to `false` (disabled) in a production database to
    protect against the risk of data loss from faulty migrations, but an
    administrator would need to remember to **re-enable the setting prior to any
    application upgrades**, to ensure that the database schema matches any new
    application changes or additions post-upgrade.

!!! tip

    For more information on database migrations, see:

    * [Database migrations](../dev-manual/db-migrations.md)

#### Internal location used for storing AIPs (a3m)

These settings are used to configure a local object store bucket or filesystem
directory if the configured [preservation engine] is [a3m]. Archivematica
includes its own [Storage Service] component, but as a lightweight command-line
derivative, a3m does not.

The defaults included below use Amazon [S3] syntax to configure a [MinIO]
bucket. Other third-party object store providers (such as Azure) will have their
own syntax - consult the provider's documentation for more information.

For **local filesystems**, the only other required field is a `url` parameter
pointing to the target location - for example:

```toml
url = "file:///home/enduro/aipstore"
```

**Default values**:

```toml
[storage.internal]
endpoint = "http://minio.enduro-sdps:9000"
pathStyle = true
key = "minio"
secret = "minio123"
region = "us-west-1"
bucket = "aips"
```

* `endpoint`: API endpoint for the target MinIO instance. Used by Enduro to read
  contents of a watched bucket, or for any other MinIO interactions.
* `pathStyle`: Currently Amazon Web Services support two different URL
  construction methods when interacting with an object store bucket via API. The
  "path-style" method constructs the bucket's access URL using the configured
  properties, such as region, bucket name, and object key. For Enduro
  integrations, the second virtual "host-style" method is not currently
  supported so if using an S3-like object store, ensure this is set to `true`.
* `key`: Username for accessing the configured S3-like bucket.
* `secret`: Password for accessing the configured S3-like bucket.
* `region`: AWS S3 buckets are created in a specific region. When interacting
  with S3, you can specify the region during the bucket creation process. For
  a full list of available regions and the syntax to specify them, consult the
  [AWS S3 documentation][S3-regions].
* `bucket`: A configured object store may have more than one bucket. This
  parameter specifies the target bucket name to be used for the internal AIP
  store location.

#### Storage event listener

These settings configure [Redis] to act as an event listener and messaging
queue with the configured storage location, so that changes in the configured
storage location can be reflected in the Enduro user interface.

At this time Redis is the only supported event listener for Enduro.

**Default values**:

```toml
[storage.event]
redisAddress = "redis://redis.enduro-sdps:6379"
redisChannel = "enduro-storage-events"
```

* `redisAddress`: Binds Redis to a specific address and port.
* `redisChannel`: Redis can be configured to use different channels for
  different queues. In this example configuration, we use same Redis
  installation and address for multiple different event listeners (see also:
  [Event queue](#event-queue) above for example), but different channels for
  each to avoid messaging conflicts.

    If you have multiple Enduro instances each talking to different event
    queues, then you would need to uniquely namespace each to avoid conflicts,
    using this parameter. Otherwise, there is no need to change the default
    value.

### Preservation engine

This configuration setting tells Enduro which [preservation engine] should be
used - either [a3m] or [Archivematica].

**Default value**:

```toml
[preservation]
taskqueue = "a3m"
```

* `taskQueue`: In Temporal, a [Task Queue] is a lightweight, dynamically
  allocated queue that one or more workers can poll for tasks. Temporal supports
  many kinds of parallelization for a distributed application architecture, and
  and task queues can be namespaced to support such a set-up.

    In this case, the setting tells Enduro the name of the task queue to expect
    for workflow tasks provided by the configured preservation engine. Supported
    values are `a3m` (the default), or `am` for Archivematica.

### a3m configuration

The next two configuration blocks must be set if you intend to use [a3m] as the
[preservation engine] for ingest workflows.

The settings in this first block define basic connection information.

**Default values**:

```toml
address = "127.0.0.1:7000"
shareDir = "/home/a3m/.local/share/a3m/share"
capacity = 1
```

* `address`: Binds a3m to a specific address and port for communication.
* `shareDir`: a3m uses a `share` directory with a number of predefined
  subdirectories for processing and package management. This parameter defines
  the location the `share` directory and its contents will be created during
  installation. Below is an example of the default directory contents following
  a3m's installation:

    ```bash
    └── share
        ├── completed
        │   └── Test-fa1d6cb3-c1fd-4618-ba55-32f01fda8198.7z
        ├── currentlyProcessing
        │   ├── ingest
        │   └── transfer
        ├── failed
        │   ├── 0d117bed-2124-48a2-b9d7-f32514d39c1e
        ├── policies
        └── tmp
    ```

* `capacity`: Limits the number of SIPs a worker can process at one time.

#### a3m processing configuration

As a lightweight headless derivative of [Archivematica], [a3m] abandons the
XML-based processing configuration document used by Archivematica. Instead,
users are asked to submit the configuration as part of each SIP submission.

To avoid that redundancy, these settings define the default processing
configuration settings that a3m will use when configured as a
[preservation engine] for Enduro.

**Default values**:

```toml
[a3m.processing]
AssignUuidsToDirectories                     = true
ExamineContents                              = true
GenerateTransferStructureReport              = true
DocumentEmptyDirectories                     = true
ExtractPackages                              = true
DeletePackagesAfterExtraction                = true
IdentifyTransfer                             = true
IdentifySubmissionAndMetadata                = true
IdentifyBeforeNormalization                  = true
Normalize                                    = true
TranscribeFiles                              = true
PerformPolicyChecksOnOriginals               = true
PerformPolicyChecksOnPreservationDerivatives = true
AipCompressionLevel                          = 1
AipCompressionAlgorithm                      = 6
```

For more detailed information on each of these configuration options, please
consult the Archivematica documentation:

* [Archivematica processing configuration fields]

!!! important

    Some of the options in the Archivematica documentation indicate that a given
    setting can be left blank to prompt a user for a decision. **This will not
    work with Enduro** - instead, the ingest workflow will eventually time out
    while polling the preservation engine for updates, causing the workflow to
    fail. For all but the last two configuration options listed above,
    acceptable values are either `true` (i.e. "Yes" in the AM documentation) or
    `false` (i.e. "No").

    In most cases this should be clear, but there are a few configuration
    options with more options supported in Archivematica than a3m. In such
    cases, remember that currently in Enduro the preservation engine only
    handles basic AIP creation - choose the options that align with that or else
    those that involve skipping the processing step described.

    For example, the **Normalize** processing configuration option in the
    Archivematica docs includes a number of options, such as "Normalize for
    preservation and access," "Normalize for access," "Normalize service files
    for access," etc. A3m cannot at present generate
    [DIPs](../user-manual/glossary.md#dissemination-information-package-dip) or
    wait for manual normalization input - it can only normalize for
    preservation. Consequently, `true` in this configuration file equates to
    option 3 in the Archivematica docs (Normalize for preservation), while
    `false` equates to option 7 (Do not normalize).

### Archivematica configuration

The next three configuration blocks must be set if you intend to use
[Archivematica] as the [preservation engine] for ingest workflows.

The settings in this first block define basic connection information.

**Default values**:

```toml
[am]
address = ""
user = ""
apiKey = ""
processingConfig = "automated"
capacity = 1
pollInterval = "10s"
transferDeadline = "1h"
transferSourcePath = ""
zipPIP = false
```

* `address`: Binds Archivematica to a specific address and port for
  communication. Because [a3m] is the installation default preservation engine,
  this is blank until configured.
* `user`: Defines the username associated with the Archivematica instance. Blank
  by default.
* `apiKey`: Defines the API key associated with the Archivematica instance.
  Blank by default.
* `processingConfig`: Set the name of the processing configuration file in
  Archivematica to use for the preservation workflow. Archivematica can have
  several different processing configuration files defined, and users can create
  new ones as needed. This value must match exactly the defined "Name" of the
  corresponding processing configuration file, which can contain only
  alphanumeric characters (a-z or A-Z and 0-9) and underscores. See the
  [AM documentation][Archivematica processing configuration fields] for more
  information.
* `capacity`: Limits the number of SIPs an Archivematica worker can process at
  one time. The default value is 1.
* `pollInterval`: When a submitted SIP in Enduro has passed any initial
  validation checks and been transformed into a [PIP] and sent to the
  [preservation engine], Enduro will then regularly poll the preservation engine
  for updates. This setting determines how frequently Enduro will poll for
  updates, with the default value being 10 seconds (`10s`). Interval values
  must be compatible with the [GoLang] [ParseDuration] function — valid values
  are "ns", "us" (or "µs"), "ms", "s", "m", and "h".
* `transferDeadline`: The maximum amount of time that Enduro should wait for
  Archivematica to finish processing a submitted package. Interval values must
  be compatible with the [GoLang] [ParseDuration] function — valid values are
  "ns", "us" (or "µs"), "ms", "s", "m", and "h". The default value is one hour
  (`1h`).
* `transferSourcePath`: The path to an Archivematica transfer source directory.
  Used in the API call to Archivematica to start processing the submitted [PIP].
  transferSourcePath must be prefixed with the UUID of an AM [Storage Service]
  transfer source directory, optionally followed by a relative path from the
  specified source directory (e.g.
  `749ef452-fbed-4d50-9072-5f98bc01e52e:sftp_upload`). If no
  `transferSourcePath` is specified, the default transfer source path will be
  used. See the [AMSS documentation] for more information.
* `zipPIP`: This boolean setting specifies whether or not a [PIP] should be
  zipped before being sent from Enduro to Archivematica. In either case, the
  package will be placed in a [BagIt] conformant bag before being transferred,
  so the preservation engine can verify the integrity of the package after
  receipt before beginning preservation processing. Default value is `false`,
  but can be changed to `true` to zip the bag before delivery.

#### Archivematica SFTP settings

These settings allow Enduro and Archivematica to use the Secure Shell File
Transfer Protocol (SFTP) for the transfer of [PIPs][PIP] to Archivematica for
preservation processing. This requires setting up an SFTP server on the target
Archivematica host first with permission to write to an
[AMSS][AMSS documentation] transfer directory, and then configuring the
connection information in the settings described here.

**Default values**:

```toml
host = ""
port = ""
user = ""
knownHostsFile = ""
remoteDir = ""
```

* `host`: The host name and address for the SFTP deposit location.
* `port`: The port to use for the SFTP deposit location.
* `user`: The username associated with the SFTP connection.
* `knownHostsFile`: The absolute path to a local SSH "known_hosts" file that
  includes a public host key for the Archivematica SFTP server. Enduro can use
  this key to verify future connection requests.

    An SFTP "known_hosts" file will store the public keys of any other server
    your SFTP server connects to when accepted, and then will use those public
    key fingerprints during future connection requests to verify the identity of
    the connecting server.
* `remoteDir`: the directory path, relative from the SFTP root directory, where
  Enduro should upload any [PIPs][PIP] for preservation processing during a
  workflow.

##### Archivematica SFTP private key connection information

These settings tell Enduro where to find the necessary information to securely
connect with the SFTP server for [PIP] transfer to Archivematica for
preservation processing.

**Default values**:

```toml
[am.sftp.privateKey]
path = ""
passphrase = ""
```

* `path`: Filesystem path to where the private key file is stored.
* `passphrase`: Pass phrase to unlock the private key. Note: while it is
  possible to store the passphrase here as plain text, we **do not recommend
  this**.

### User interface SIP upload filesize limit

These settings define the maximum size of a SIP that Enduro will allow
when uploading via the user interface. The setting is configured in bytes, but
will be shown in the user interface using the largest [SI prefix] (up to TiB)
that will display a unit value larger than one. Possible units values are:
bytes, KiB, MiB, GiB, and TiB. For more information, see:

* [Upload SIPs via the user interface][upload-ui]

**Default value**:

```toml
[upload]
maxSize = 4294967296
```

* `maxSize`: The maximum SIP size allowed for upload via the user interface,
  configured in bytes. Default value is 4294967296, i.e. 4 [Gibibytes] (GiB).

### Internal storage configuration

The following two sections are to configure an internal upload and failed
package storage space, used by Enduro for [uploads via the user
interface][upload-ui] as well as SIPs or [PIPs][PIP] that encounter [content
failures] or [system errors] during ingest, so they can be downloaded for
inspection via the user interface by operators if desired. At least one of the
sections below must be configured.

#### Internal storage configuration - S3-like bucket or filesystem

This subsection allows an object store bucket or local filesystem location to be
configured for UI uploads and failed packages. The default configuration
included uses a [MinIO] bucket as the example location. MinIO uses Amazon [S3]
syntax for its configuration properties. Different object stores may have
different parameters to be configured. Consult the corresponding object store
provider's documentation for more information, or, if using [Azure] blob
storage, use the subsection
[below](#internal-storage-configuration---azure-blob-storage) instead.

For **local filesystems**, the only required field is a `url` parameter
pointing to the target location - for example:

```toml
[sipsource.bucket]
url = "file:///home/enduro/internalstore"
```

**Default values**:

```toml
[internalStorage.bucket]
endpoint = "http://minio.enduro-sdps:9000"
pathStyle = true
accessKey = "minio"
secretKey = "minio123"
region = "us-west-1"
bucket = "internal"
```

* `endpoint`: API endpoint for the target MinIO instance. Used by Enduro to read
  contents of a bucket, or for any other MinIO interactions.
* `pathStyle`: Currently Amazon Web Services support two different URL
  construction methods when interacting with an object store bucket via API. The
  "path-style" method constructs the bucket's access URL using the configured
  properties, such as region, bucket name, and object key. For Enduro
  integrations, the second virtual "host-style" method is not currently
  supported so if using an S3-like object store, ensure this is set to `true`.
* `accessKey`: Username for accessing the configured S3-like bucket.
* `secretKey`: Password for accessing the configured S3-like bucket.
* `region`:  = AWS S3 buckets are created in a specific region. When interacting
  with S3, you can specify the region during the bucket creation process. For
  a full list of available regions and the syntax to specify them, consult the
  [AWS S3 documentation][S3-regions].
* `bucket`: A configured object store may have more than one bucket. This
  parameter specifies the target bucket name to be used for the uploaded SIPs
  and failed packages location. Because the default configuration uses MinIO
  buckets elsewhere as well, ensure that this bucket name is unique.

#### Internal storage configuration - Azure blob storage

This subsection allows an [Azure] blob storage location to be configured for UI
uploads and failed packages, instead of the default S3-like object store
configured above. At installation the values in this subsection are left blank.

If you intend to use Azure for your internal upload space, be sure to remove or
comment out the S3-like object store parameters in the section
[above](#internal-storage-configuration---s3-like-bucket-or-filesystem), and
then add the Azure blob URL connection information there instead. For example,
if your target Azure storage blob is named "sips", your configuration of the
section above might look like this when configured properly:

```toml
[internalStorage.bucket]
url = "azblob://sips"
endpoint = ""
pathStyle = ""
accessKey = ""
secretKey = ""
region = ""
bucket = ""
```

The following subsection provides Enduro with the connection information needed
to access packages in the target Azure blob store.

**Default values**:

```toml
[internalStorage.azure]
storageAccount = ""
storageKey = ""
```

* `storageAccount`: The name of the Azure storage account associated with the
  target blob.
* `storageKey`: The private key that Enduro should use to authenticate and gain
  access to the target blob.

### SIP source location configuration

The following sections configure a location to be used as a SIP source location
for SIP selection and ingest — for more information, see:
[Add SIPs via a source location][sip-source].

The first set of parameters uniquely identify the source location in Enduro,
while the bucket subsection links the specified location.

!!! tip

    Once configured, ingests from a package in a configured source location can
    also be initiated via the [API].

**Default values**:

```toml
[sipsource]
id = "e6ddb29a-66d1-480e-82eb-fcfef1c825c5"
name = "MinIO SIP Source"
```

* `id`: A UUID that unique identifies the SIP source location. Must be a valid
  [version 4 UUID].
* `name`: A human-readable name for the SIP source.

#### SIP source location bucket

This subsection allows either an S3-like object store bucket or a local
filesystem location to be configured as a SIP source location. The
default configuration included uses a [MinIO] bucket as the example location.
MinIO uses Amazon [S3] syntax for its configuration properties. Different object
stores may have different parameters to be configured. Consult the corresponding
object store provider's documentation for more information.

For **local filesystems**, the only required field is a `url` parameter
pointing to the target location - for example:

```toml
[sipsource.bucket]
url = "file:///home/enduro/sipsource"
```

**Default values**:

```toml
[sipsource.bucket]
endpoint = "http://minio.enduro-sdps:9000"
pathStyle = true
accessKey = "minio"
secretKey = "minio123"
region = "us-west-1"
bucket = "sipsource"
```

* `endpoint`: API endpoint for the target MinIO instance. Used by Enduro to read
  contents of a bucket, or for any other MinIO interactions.
* `pathStyle`: Currently Amazon Web Services support two different URL
  construction methods when interacting with an object store bucket via API. The
  "path-style" method constructs the bucket's access URL using the configured
  properties, such as region, bucket name, and object key. For Enduro
  integrations, the second virtual "host-style" method is not currently
  supported so if using an S3-like object store, ensure this is set to `true`.
* `accessKey`: Username for accessing the configured S3-like bucket.
* `secretKey`: Password for accessing the configured S3-like bucket.
* `region`:  = AWS S3 buckets are created in a specific region. When interacting
  with S3, you can specify the region during the bucket creation process. For
  a full list of available regions and the syntax to specify them, consult the
  [AWS S3 documentation][S3-regions].
* `bucket`: A configured object store may have more than one bucket. This
  parameter specifies the target bucket name to be used for the SIP source
  location. Because the default configuration uses MinIO buckets elsewhere as
  well, ensure that this bucket name is unique.

### Telemetry configuration

Telemetry is the process of collecting and analyzing application data to
gain insights into system performance, user behavior, resource usage, and more.

Enduro includes a [GoLang] [OpenTelemetry] package, which when enabled and
properly configured, can support the generation, collection, and export of
telemetry data from Enduro for debugging, optimization, and development purposes.
At this time, no user data is reviewed or collected by the Enduro telemetry
package.

!!! tip

    For additional information on logging, see:

    * [Logging](../dev-manual/logging.md)

**Default values**:

```toml
[telemetry.traces]
enabled = false
address = ""
samplingRatio = 1.0
```

* `enabled`: Boolean value that enables or disables telemetry in Enduro. Set to
  `false` by default - change to `true` to enable.
* `address`: the gRPC address and port of the related observability tool used
  for parsing the collected telemetry data. There are many different tools (both
  proprietary and open source) that can be used with OpenTelemetry data - see
  the [OpenTelemetry docs] for a list of vendor tools known to be compatible.
* `samplingRatio`: This ratio defines the frequency and granularity of data
  collection, with 1.0 representing a 100% capture of all events regardless of
  repetition. In many cases an internal event or call may be repeated dozens or
  even hundreds of times, leading to redundancy in the captured logs. A
  developer or system administrator may then wish to reduce the logging of
  similar events by lowering the sampling ratio - for example, if the ratio is
  reduced to 0.5 (i.e. 50% sampling ratio), then a single event that occurs 10
  times would only be logged 5 times in the resulting trace data.

### Preprocessing child workflow configuration

The following two sections are for configuring a [child workflow] as part of the
ingest workflow run by Enduro. A child workflow is a concept borrowed from
Enduro's [workflow engine], [Temporal] where one workflow spawns an ancillary
workflow. We use this model in Enduro to keep custom activities with
organization-specific business rules or policy requirements separated from the
underlying general Enduro code. For more information, see:

* [Preprocessing child workflow](../dev-manual/preprocessing.md)

**Default values**:

```toml
[preprocessing]
enabled = false
extract = false
sharedPath = "/home/enduro/preprocessing"
```

* `enabled`: Boolean value, set to `false` by default. To enable and configure a
  custom child workflow, change to `true`.
* `extract`: Boolean value, set to `false` by default. This setting determines
  whether SIP extraction happens as part of the child workflow or not. When set
  to false, SIP extraction will occur in the parent Enduro workflow before the
  child workflow is run. In some cases, you may wish to design custom child
  workflow activities that perform operations on the SIP before it is extracted:
  for example, calculating a checksum of the zipped package to check for prior
  duplicate ingests before proceeding. When this value is set to `true` Enduro
  will skip the extraction task at the beginning of the ingest workflow,
  allowing you to define when and how extraction occurs in the child workflow.
* `sharedPath`: The absolute path to the directory that Enduro should use to
  share the SIP between the primary ingest workflow and the configured
  preprocessing child workflow. This path is required when `enabled` is set to
  true.

#### Temporal workflow configuration for preprocessing workflow

These settings ensure that Enduro can connect with [Temporal] and trigger the
configured [child workflow].

**Default values**:

```toml
[preprocessing.temporal]
namespace = "default"
taskQueue = "preprocessing"
workflowName = "preprocessing"
```

* `namespace`: An internal namespace label for the Temporal workflows. This only
  needs to be changed if you expect to be running multiple different workflows
  running at once, so they don't clash.
* `taskQueue`: In Temporal, a [Task Queue] is a lightweight, dynamically
  allocated queue that one or more workers can poll for tasks. Temporal supports
  many kinds of parallelization for a distributed application architecture, and
  and task queues can be namespaced to support such a set-up. The default value
  of `preprocessing` here differentiates it from the general `global` task queue
  used by Enduro, and ensures that Enduro looks for child workflow tasks in a
  dedicated queue.
* `workflowName`: The name of the configured [child workflow] in Temporal.

### Post-storage workflow configuration

In addition to configuring a [child workflow] for custom ingest processing
tasks, Enduro can also run additional child workflows for [post-storage]
processing when properly configured.

Post-storage is a phase in an ingest or preservation workflow describing all the
preservation policy-defined tasks performed on an [AIP] following
preservation processing and AIP storage. Post-storage task examples might
include metadata extraction and delivery to an external system (such as an
archival management system), AIP encryption or replication, and more.

These settings can be used to enable and configure post-storage child workflows
for Enduro, which will be run after AIP creation and storage following a
successful ingest. At installation this section is commented out to render it
inactive - remove the `#` hash symbol from the start of each line to enable this
section as you configure the values.

!!! note

    The post-storage configuration block can be repeated more than once to add
    multiple post-storage child workflows. This is why the `poststorage` header
    is in double brackets rather than single brackets like other configuration
    blocks.

**Default values**:

```toml
# [[poststorage]]
# namespace = "default"
# taskQueue = "poststorage"
# workflowName = "poststorage"
```

* `namespace`: An internal namespace label for the Temporal workflows. This only
  needs to be changed if you expect to be running multiple different workflows
  running at once, so they don't clash.
* `taskQueue`: In Temporal, a [Task Queue] is a lightweight, dynamically
  allocated queue that one or more workers can poll for tasks. Temporal supports
  many kinds of parallelization for a distributed application architecture, and
  and task queues can be namespaced to support such a set-up. The default value
  of `poststorage` here differentiates it from the general `global` task queue
  used by Enduro and the `preprocessing` task queue default used for custom
  ingest child workflows (see
  [above](#preprocessing-child-workflow-configuration)), and ensures that Enduro
  looks for post-storage child workflow tasks in a dedicated queue.
* `workflowName`: The name of the configured [child workflow] in Temporal.

[a3m]: https://github.com/artefactual-labs/a3m
[ABAC]: https://en.wikipedia.org/wiki/Attribute-based_access_control
[AIP]: ../user-manual/glossary.md#archival-information-package-aip
[AMSS documentation]: https://www.archivematica.org/docs/storage-service-latest/administrators
[API]: ../dev-manual/api.md
[Archivematica]: https://archivematica.org/
[Archivematica processing configuration fields]: https://archivematica.org/docs/latest/user-manual/administer/dashboard-admin/#processing-configuration-fields
[Azure]: https://azure.microsoft.com/en-us/products/storage/blobs/
[BagIt]: https://www.rfc-editor.org/rfc/rfc8493
[child workflow]: ../user-manual/glossary.md#child-workflow
[components]: ../user-manual/components.md
[content failures]: ../user-manual/glossary.md#content-failure
[CORS]: https://en.wikipedia.org/wiki/Cross-origin_resource_sharing
[Gibibytes]: https://www.difference.wiki/gigabyte-vs-gibibyte/
[GoLang]: https://go.dev/
[Keycloak]: https://www.keycloak.org/
[METS]: https://www.loc.gov/standards/mets/
[MinIO]: https://www.min.io
[OIDC]: https://en.wikipedia.org/wiki/OpenID#OpenID_Connect_(OIDC)
[OIDC specification]: https://openid.net/specs/openid-connect-core-1_0.html
[OpenTelemetry]: https://opentelemetry.io/docs/what-is-opentelemetry/
[OpenTelemetry docs]: https://opentelemetry.io/ecosystem/vendors/
[ParseDuration]: https://pkg.go.dev/time#ParseDuration
[PIP]: ../user-manual/glossary.md#processing-information-package-pip
[post-storage]: ../user-manual/glossary.md#post-storage
[PREMIS]: https://www.loc.gov/standards/premis/
[preservation engine]: ../user-manual/glossary.md#preservation-engine
[Redis]: https://redis.io/
[S3]: https://aws.amazon.com/s3/
[S3-regions]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/s3-tables-regions-quotas.html#s3-tables-regions
[SI prefix]: https://en.wikipedia.org/wiki/Metric_prefix
[sip-source]: ../user-manual/ingest/submitting-content.md#add-sips-via-a-source-location
[Storage Service]: https://archivematica.org/docs/storage-service-latest/
[system errors]: ../user-manual/glossary.md#system-error
[Task Queue]: https://docs.temporal.io/task-queue
[Temporal]: https://temporal.io
[upload-ui]: ../user-manual/ingest/submitting-content.md#upload-sips-via-the-user-interface
[version 4 UUID]: https://www.rfc-editor.org/rfc/rfc9562.html#name-uuid-version-4
[watched-location]: ../user-manual/ingest/submitting-content.md#initiate-ingest-via-a-watched-location-upload
[workflow engine]: ../user-manual/glossary.md#workflow-engine
