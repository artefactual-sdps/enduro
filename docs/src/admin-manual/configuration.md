# Configuration

This page describes the various configuration files and settings that Enduro
supports. See also:

* [Dashboard configuration]

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

    Note that some [child workflow] activities may have their own configuration
    files.

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

**Example configuration**:

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

**Example configuration**:

```toml
[temporal]
address = "temporal.enduro-sdps:7233"
namespace = "default"
taskQueue = "global"
```

* `address`: Tells Enduro where to find Temporal - address and port connection
  information.
* `namespace`: The Temporal namespace in which Enduro workflows run. Custom
  child workflows inherit this namespace. Workers that execute custom child
  workflows must connect to the same namespace.
* `taskQueue`: In Temporal, a [Task Queue] is a lightweight, dynamically
  allocated queue that one or more workers can poll for tasks. Temporal supports
  many kinds of parallelization for a distributed application architecture, and
  and task queues can be namespaced to support such a set-up.

### Internal API

Configuration information for Enduro's internal API calls across different
[components]. Unlike the other API configuration section [below](#api), the
internal API does not use or support authentication. As such, the configured
port should **not be exposed publicly**.

**Example configuration**:

```toml
[internalapi]
listen = "0.0.0.0:9002"

[internalapi.log]
path = "stdout"
level = "warn"
format = "json"
```

* `listen`: tells Enduro what IP address and port to use for internal API
  communication across components.

See [API Logs](#api-logs) for logging configuration details.

### API

The next set of four configuration blocks all relate to Enduro's primary
external facing [API]. This first block sets the basic configuration of the API.

!!! tip

    See also:

    * [Enduro API documentation](../dev-manual/api.md)

**Example configuration**:

```toml
[api]
listen = "0.0.0.0:9000"
corsOrigin = "http://localhost"

[api.log]
path = "/var/log/enduro/api.log"
level = "warn"
format = "json"
```

* `listen`: Specifies the address and port the API will bind to for
  communication.
* `corsOrigin`: [CORS] (short for Cross-Origin Resource Sharing) is a security
  mechanism implemented by web browsers that controls how web pages can access
  resources from different domains, ports, or protocols. This setting defines a
  policy for Enduro to use the specified value as the primary origin domain.

See [API Logs](#api-logs) for logging configuration details.

#### API timeout behavior

Enduro applies a service-level operation timeout to normal Goa API request
handlers. This timeout bounds backend work after the request has been decoded
and before the response is written.

Normal API request handlers have a 5 second operation timeout. Operations that
exceed this budget are canceled through the request context, logged as timed out,
recorded on the active OpenTelemetry span, and returned as an internal API
error when the transport is still open. Operations slower than 2 seconds are
recorded as slow operations.

This service-level timeout is different from HTTP transport and proxy timeouts:

* The service-level timeout controls backend work owned by Enduro.
* The API server write timeout controls how long the HTTP connection can remain
  open while Enduro handles and writes the response. Enduro sets this to 7
  seconds so normal API handlers have time to return a timeout response after
  the 5 second service budget expires.
* Reverse proxy idle timeouts, such as NGINX `proxy_read_timeout`, should be
  higher than Enduro's API write timeout for normal API routes. The NGINX
  defaults are sufficient for this, and these settings measure idle gaps between
  read or write operations rather than total upload or download duration.

The generic service-level timeout is not applied to streaming, upload, download,
or SSE endpoints. These endpoints have different timeout requirements and
must be handled by transport, proxy, and endpoint-specific setup timeouts. This
includes:

* SIP upload: `/api/ingest/sips/upload`
* SIP download: `/api/ingest/sips/{uuid}/download`
* AIP download: `/api/storage/aips/{uuid}/download`
* AIP deletion report download:
  `/api/storage/aips/{uuid}/deletion-report`
* Ingest monitor SSE stream: `/api/ingest/monitor`
* Storage monitor SSE stream: `/api/storage/monitor`

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
enabled = false
```

#### OIDC authentication providers configuration

These setting blocks are used to configure the OpenID Connect ([OIDC])
providers used for access token verification when API authentication is
[enabled](#enable-api-authentication). Multiple OIDC providers can be
configured with a TOML array-of-tables syntax with the same structure. When
API authentication is enabled, at least one OIDC provider must be configured to
verify tokens.

For more details on OIDC configuration, consult the [OIDC specification].

**Example configuration**:

```toml
[[api.auth.oidc]]
providerURL = "https://idp-public.example.com/realms/enduro"
clientID = "enduro"

[[api.auth.oidc]]
providerURL = "https://idp-public.example.com/realms/enduro"
clientID = "enduro-s2s"
skipEmailVerifiedCheck = true
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

##### Enable Attribute Based Access Control for the API OIDC authentication

Enduro uses Attribute Based Access Control ([ABAC]) to manage permissions and
access. When ABAC is enabled for an OIDC verifier, it will check a configurable
multivalue claim against the defined required attributes based on each
endpoint's configuration.

For each verifier, add a matching `[api.auth.oidc.abac]` section immediately
after the verifier configuration if ABAC is needed.

**Example configuration**:

```toml
[[api.auth.oidc]]
providerURL = "https://idp-public.example.com/realms/enduro"
clientID = "enduro"

[api.auth.oidc.abac]
enabled = true
claimPath = "enduro"
claimPathSeparator = ""
claimValuePrefix = ""
useRoles = false
rolesMapping =

[[api.auth.oidc]]
providerURL = "https://idp-public.example.com/realms/enduro"
clientID = "enduro-s2s"
skipEmailVerifiedCheck = true
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

#### Redis ticket storage for browser downloads

Enduro can use [Redis] to store the short-lived tickets used by browser
download handoffs. When Redis ticket storage is not configured, Enduro uses
in-memory tickets instead.

**Example configuration**:

```toml
[api.auth.ticket.redis]
address = "redis://redis.enduro-sdps:6379"
prefix = "enduro"
```

* `address`: Binds Redis to a specific address and port.
* `prefix`: Defines a namespace that can be used as a prefix to differentiate
  instances if you have multiple Enduro installations running. Otherwise, there
  is no reason to change the default "enduro" value.

#### API Logs

Enduro API logs are handled separately from the application logs due to
different volume, retention, routing, and alerting needs. API logs include HTTP
request and response data and can be written to a file, standard output
(stdout), or standard error (stderr). API log messages are encoded in text or
or JSON format.

The API and Internal API logs are configured independently and can log to
separate files or the same file.

**Example configurations**:

```toml
[api.log]
path = "/var/log/enduro/api.log"
level = "info"
format = "json"

[internalapi.log]
path = "stdout"
level = "warn"
format = "text"
```

* `path`: sets the path of the API log file. If `path` is set to a regular file
  path, logs messages are appended to that file. If `path` is set to "stdout" or
  "stderr" log messages are written to standard output or standard error
  respectively. If `path` is not set, or set to an empty string, API logging is
  disabled.
* `level` defines the verbosity of the API logs:
    * "debug" - log all responses (incl. OPTIONS)
    * "info"  - log responses (excl. OPTIONS)
    * "warn"  - log 4xx and 5xx responses only (except for 429)
    * "error" - log 5xx responses only
  The default `level` is "info", and level names are case-insensitive.
* `format`: sets the output encoding of the log messages. Supported formats are
  "json" (the default) and "text". If set to "json" log messages are encoded as
  line delimited JSON, intended for machine analysis or aggregation. If set to
  "text" log messages are encoded as plain text with "key=value" data pairs. The
  "text" format is intended for human readability.

**N.B.** If `path` is a normal file (not stdout or stderr) the log file will
be created by Enduro if it does not exist already, but the full directory path
to the file must exist when the Enduro worker is started or the worker will exit
with an error.

### BagIt bag creation

Enduro makes each PIP delivered to [Archivematica] a [BagIt] bag, providing
file integrity checksums to ensure the PIP contents are not corrupted or
truncated in transmission. The BagIt settings configure how these bags are
created.

[a3m] does not support ingest of BagIt bags, so PIPs are structured as
an [Archivematica transfer with existing checksums] when [a3m] is used for
preservation and these settings are not relevant.

**Example configuration**:

```toml
[bagit]
chechecksumAlgorithm = "sha512"
```

* `checksumAlgorithm` sets the hashing algorithm used to generate file checksums
  in created BagIt bag manifests. Valid values are "md5", "sha1", "sha256", and
  "sha512" (the default).

### BagIt validation

Enduro validates any [BagIt] bags submitted for ingest to ensure the contents
have not been altered or corrupted. In addition Enduro requires a
[Preprocessing child workflow](#preprocessing-child-workflow) to deliver a
[BagIt] bag to Enduro, and these bags are also validated by Enduro to ensure
file integrity.

**Example configuration**:

```toml
[bagitValidator]
cacheDir = "/home/enduro/.cache/bagit-gython"
poolSize = 2
```

* `cacheDir` sets the cache directory used by the [bagit-gython] validator's
  runtime and runners. If `cacheDir` is an empty string or omitted, the
  validator will attempt to create a cache directory in the process user's home
  directory (e.g. /home/enduro/.cache/bagit-gython). If the user's home
  directory is not available (e.g. because the process has no home
  directory), the validator will fall back to using a unique temporary directory
  (e.g. /tmp/bagit-gython-12345) that will be deleted at shutdown.
* `poolSize` sets the number of available concurrent bag validation runners.
  `poolSize` must be 1 (the default value) or greater. If the number of
  requested validation jobs exceeds the available runners, the extra jobs will
  be queued and run when a runner becomes available. See the
  [bagit-gython README] for more details on how the validator pool works.

### Database connection

These settings configure the connection information for Enduro's MySQL database.
Note that other MySQL distributions (like MariaDB or Percona) can be used, but
other database types are not currently supported in Enduro.

**Example configuration**:

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

**Example configuration**:

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

**Example configuration**:

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
from a watched location. Enduro supports two watched-location implementations:

* Filesystem watchers monitor a local directory for new files or directories.
  They do not use a message queue.
* Legacy object-store watchers consume MinIO Redis notification events. This
  path is still available for deployments with MinIO-compatible event
  publishers.

For filesystem watchers, make the deposit appear at its final name only after it
is complete. The watcher reacts to top-level create or rename events; it does
not wait for a file copy or directory tree copy to become stable. Copy or build
the package in a staging location outside the watched path, then move or rename
the completed file or directory into the watched path. If staging inside the
watched path is required, configure `ignore` to match the temporary name, then
rename the completed deposit to an unignored final name. The process running the
watcher must be able to read the watched path and must have permission to delete
or move deposits if retention or `completedDir` cleanup is enabled.

For legacy object-store watchers, the watched location must publish an event to
Enduro's message queue ([Redis], listening for events in the queue defined by
the `redisList` parameter below) any time a new zipped package is added.
Enduro's internal watcher will trigger the ingest workflow when it detects a SIP
deposit in the configured watched location. Enduro does not currently support
Azure Blob Storage or generic S3-compatible buckets as event-driven watched
locations. For more information on ingests from a watched location, see:

* [Initiate ingest via a watched location upload][watched-location]

#### Filesystem watcher

Use a repeated `[[watcher.filesystem]]` table for each filesystem watched
location.

**Example configuration**:

```toml
[[watcher.filesystem]]
name = "filesystem-dropbox"
path = "/home/enduro/watched"
ignore = "^\\."
inotify = false
pollInterval = "200ms"
retentionPeriod = "-1s"
completedDir = "/home/enduro/watched-complete"
workflowType = "create aip"
```

* `name`: Defines a name to be used internally for the watched location.
  This must be unique across all configured watchers.
* `path`: Filesystem directory that Enduro watches for new deposits. The
  directory must exist before Enduro starts.
* `ignore`: Optional regular expression matched against the base name of a
  created file or directory. Matching deposits are ignored. Use this when an
  incomplete deposit must be staged inside the watched path under a temporary
  name before being renamed to its final name.
* `inotify`: Set to `true` to prefer filesystem event notifications on
  platforms where they are available. The default `false` uses polling.
* `pollInterval`: Time between filesystem polls when polling is used. The
  default is `200ms`; use a string format compatible with [ParseDuration].
* `retentionPeriod`: Duration to retain the original SIP before deleting it
  after a successful ingest. Set to a negative value to disable automatic
  deletion. Set to `"0"` to delete immediately. Use a string format compatible
  with [ParseDuration].
* `completedDir`: Directory where Enduro moves the original SIP after a
  successful ingest. This setting can only be used when `retentionPeriod` is
  negative.
* `workflowType`: Specifies the name of the Enduro workflow type to be run when
  SIPs are deposited in the watched location. Currently the only supported
  values are "create aip" and "create and review aip". The latter review
  workflow also only works if [a3m] is the configured [preservation engine].

#### Legacy MinIO Redis watcher

At this time, [Redis] is the only supported messaging queue for legacy
object-store watchers, so `redisAddress` and `redisList` are required.
All other default parameters are MinIO bucket access settings.

**Example configuration**:

```toml
# The legacy watched-location watcher consumes MinIO Redis notification events.
# Keep this block commented unless your deployment provides MinIO-compatible
# Redis notification events.
# [watcher.embedded]
# name = "legacy-minio-events"
# redisAddress = "redis://redis.enduro-sdps:6379"
# redisList = "minio-events"
# endpoint = "http://minio.example:9000"
# pathStyle = true
# key = "example-access-key"
# secret = "example-secret-key"
# region = "us-west-1"
# bucket = "sips"
# workflowType = "create aip"
```

* `name`: Defines a name to be used internally for the watched location. Useful
  if you are configuring more than one watched location.
* `redisAddress`: Binds Redis to a specific address and port.
* `redisList`: The name of the queue that Redis should use for the watched
  location SIP deposit events.
* `endpoint`: API endpoint for the MinIO S3 API.
* `pathStyle`: Currently Amazon Web Services support two different URL
  construction methods when interacting with an object store bucket via API. The
  "path-style" method constructs the bucket's access URL using the configured
  properties, such as region, bucket name, and object key. For Enduro
  integrations, the second virtual "host-style" method is not currently
  supported. MinIO deployments typically use `pathStyle = true`.
* `key`: Access key for the configured bucket.
* `secret`: Secret key for the configured bucket.
* `region`: AWS S3 buckets are created in a specific region. When interacting
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

### Bucket configuration options

Several Enduro settings use the same bucket configuration shape for filesystem
or object store locations:

* `[storage.internal]`
* `[a3m.aipStaging]`
* `[internalStorage]`
* `[sipsource.bucket]`

Each bucket can be configured with either a URL or provider-specific fields.
When `url` is set, it takes precedence over the S3-compatible fields listed
below. The URL form supports local filesystem locations, S3-compatible buckets,
and [Azure Blob Storage][Azure] containers. S3-compatible buckets can
alternatively be configured with the endpoint, credentials, region, and bucket
fields below.

Use the examples below with the section you are configuring. Replace
`[example.bucket]` with `[storage.internal]`, `[internalStorage]`,
`[a3m.aipStaging]`, or `[sipsource.bucket]` as appropriate.

**Example filesystem URL configuration**:

```toml
[example.bucket]
url = "file:///home/enduro/example-bucket?metadata=skip&no_tmp_dir=true"
```

For filesystem-backed storage, the directory referenced by `url` must exist
before Enduro starts, and the process user running Enduro must have the read,
write, or delete permissions required by that bucket's role. When Enduro
processes run on different hosts or containers, the filesystem path must refer
to shared storage that is mounted at the same path for every process that needs
to access the bucket.

Enduro implements `file://` bucket URLs with the Go CDK `fileblob` package.
By default, `fileblob` stores blob metadata in `.attrs` sidecar files, and
writes objects to the system temporary directory before renaming them to the
final path to avoid partial writes. These defaults are useful for general
filesystem-backed buckets, but they can be surprising in Enduro deployments:
metadata sidecars appear beside package files, and renames can fail when the
system temporary directory and bucket directory are on different filesystems or
container mounts.

For this reason, Enduro filesystem bucket examples use `metadata=skip` to
prevent `.attrs` sidecar files. Use `no_tmp_dir=true` for filesystem buckets
that are written through `fileblob`; the bucket-specific sections below note
when this applies. This keeps temporary write files next to the final object,
avoiding cross-device rename failures. Both parameters are documented in the Go
CDK [fileblob URL opener] documentation.

**Example S3-compatible URL configuration**:

```toml
[example.bucket]
url = "s3://example-bucket?region=us-west-1"
```

**Example S3-compatible field configuration**:

```toml
[example.bucket]
url = ""
endpoint = "https://s3.example.com"
pathStyle = true
accessKey = "example-access-key"
secretKey = "example-secret-key"
region = "us-west-1"
bucket = "example-bucket"
```

* `url`: URL for the target bucket or filesystem location. This can be used for
  local filesystems, S3-compatible object storage, or [Azure Blob
  Storage][Azure]. When `url` is set, it takes precedence over the individual
  S3-compatible settings below.
* `endpoint`: API endpoint for the target S3-compatible object store.
* `pathStyle`: Currently Amazon Web Services support two different URL
  construction methods when interacting with an object store bucket via API. The
  "path-style" method constructs the bucket's access URL using the configured
  properties, such as region, bucket name, and object key. For Enduro
  integrations, the second virtual "host-style" method is not currently
  supported so if using an S3-compatible object store, ensure this is set to
  `true`.
* `accessKey`: Username for accessing the configured S3-compatible bucket.
* `secretKey`: Password for accessing the configured S3-compatible bucket.
* `region`: AWS S3 buckets are created in a specific region. When interacting
  with S3, you can specify the region during the bucket creation process. For
  a full list of available regions and the syntax to specify them, consult the
  [AWS S3 documentation][S3-regions].
* `bucket`: A configured object store may have more than one bucket. This
  parameter specifies the target bucket name to be used by the configured
  section. Because the default configuration uses multiple buckets, ensure that
  each bucket name is unique.

For [Azure Blob Storage][Azure], set an Azure bucket `url` using the
`azblob://` scheme and leave the S3-compatible fields empty. The Azure bucket
name is included in the URL, so there is no need for a separate bucket config
value. Then add the matching Azure authentication section:

* `[storage.internal]` uses `[storage.internal.azure]`.
* `[internalStorage]` uses `[internalStorage.azure]`.
* `[a3m.aipStaging]` uses `[a3m.aipStaging.azure]`.
* `[sipsource.bucket]` uses `[sipsource.bucket.azure]`.

**Example Azure shared key configuration**:

```toml
[example.bucket]
url = "azblob://example-bucket"

[example.bucket.azure]
storageAccount = "my-account"
storageKey = "dGVzdA==" # Must be base64 encoded
```

**Example Azure client secret configuration**:

```toml
[example.bucket]
url = "azblob://example-bucket"

[example.bucket.azure]
storageAccount = "my-account"
tenantID = "my-tenant-id"
clientID = "my-client-id"
clientSecret = "my-secret"
```

* `storageAccount`: The name of the Azure storage account.
* `storageKey`: The shared key that Enduro should use to authenticate.
* `tenantID`: Azure tenant ID for client secret authentication.
* `clientID`: Azure client ID for client secret authentication.
* `clientSecret`: Azure client secret for client secret authentication.

When using Azure Blob Storage, Enduro supports two authentication modes:

* Shared key authentication using `storageKey`.
* Client secret authentication using `tenantID`, `clientID`, and
  `clientSecret`.

If both authentication modes are configured, the client secret settings take
precedence.

### Ingest settings

The ingest configuration section configures the SIP ingest workflow.

**Example configuration**:

```toml
[ingest]
allowDuplicates = false
```

#### allowDuplicates

The `allowDuplicates` setting toggles whether a SIP can be ingested more than
once by Enduro. The default value (false) will stop ingest with a content error
when a SIP archive file (e.g. a zip) is submitted that has the same checksum as
a previously ingested SIP. SIPs submitted as a directory are not checked for
duplicate contents.

A SIP is only considered a duplicate if the checksum matches an existing
SIP with a status of: "ingested", "pending", "processing", "queued", or
"validated". If the SIP status is "error", "failed" or "canceled" the
SIP will be ignored when checking for duplicates.

A checksum is calculated and stored for every SIP archive ingested by Enduro,
regardless of this setting. When `allowDuplicates` is false, a new ingest's
checksum will be checked against all previously ingested SIP checksums, even if
`allowDuplicates` was true when the old SIPs were ingested.

### Ingest storage settings

This element configures the Enduro storage service API endpoint. Even when using
[Archivematica] as the [preservation engine] (which includes the AM
[Storage Service] as the AIP store), this endpoint will act as a proxy to the
AMSS. Otherwise, this configures the primary local storage service when using
[a3m].

**Example configuration**:

```toml
[ingest.storage]
address = "enduro.enduro-sdps:9002"
defaultPermanentLocationId = "f2cc963f-c14d-4eaa-b950-bd207189a1f1"
```

* `address` **required**: Defines the address and port for the storage API
  endpoint.
* `defaultPermanentLocationId` **required**: The UUID of the storage location
  used for permanent AIP storage in automated workflows.

#### Ingest storage OIDC settings

These settings configure client credentials authentication for ingest requests
to the storage API. When enabled, ingest requests an access token from the
configured OIDC provider and sends it as a bearer token. Tokens generated by
this OIDC provider must be verified by at least one of the providers from the
full [API OIDC configuration](#oidc-authentication-providers-configuration).

**Default values**:

```toml
[ingest.storage.oidc]
enabled = false
providerURL = ""
tokenURL = ""
clientID = ""
clientSecret = ""
audience = ""
scopes = ""
retryMaxAttempts = 3
retryInitialInterval = "500ms"
retryMaxInterval = "2s"
retryBackoffCoefficient = 2.0
tokenExpiryLeeway = "30s"
```

* `enabled`: Enables OIDC client credentials for ingest to storage API calls.
* `providerURL`: OIDC provider URL used for token endpoint discovery.
* `tokenURL`: Optional token endpoint URL. If set, discovery is skipped.
* `clientID`: OIDC client ID used for token requests.
* `clientSecret`: OIDC client secret used for token requests.
* `scopes`: Optional comma-separated scopes requested during token retrieval.
* `audience`: Optional audience sent in token endpoint parameters.
* `tokenExpiryLeeway`: Time before expiry when cached tokens are refreshed.
* `retryMaxAttempts`: Maximum attempts when token retrieval fails transiently.
* `retryInitialInterval`: Initial retry backoff interval.
* `retryMaxInterval`: Maximum retry backoff interval.
* `retryBackoffCoefficient`: Exponential multiplier applied between retry
  attempts (must be >= 1).

#### Storage database

Even when using the Archivematica [Storage Service] (AMSS), Enduro still
captures and stores some storage metadata for display in the user interface. The
following elements define the database used to store this information and how
Enduro can connect with it for read and write operations.

Enduro uses a MySQL database by default. Note that other MySQL distributions
(like MariaDB or Percona) can be used, but other database types are not
currently supported in Enduro.

**Example configuration**:

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

#### Internal location used for storing staging AIPs and reports

These settings are used to configure an internal filesystem directory or object
store bucket where the storage service places reports, and stages AIPs when
[a3m] is the configured [preservation engine]. Archivematica includes its own
[Storage Service] component, but as a lightweight command-line derivative, a3m
does not.

Configure this bucket with the shared
[bucket configuration options](#bucket-configuration-options), using
`[storage.internal]` as the bucket section. For Azure Blob Storage, use
`[storage.internal.azure]` as the matching authentication section.

**Example filesystem configuration**:

```toml
[storage.internal]
url = "file:///home/enduro/internal-storage/storage?metadata=skip&no_tmp_dir=true"
```

For filesystem-backed storage, Enduro writes to this bucket, so the process
user must have write access. If multiple Enduro processes need to access this
bucket, the filesystem path must refer to storage shared by those processes.
Because Enduro writes to this bucket through `fileblob`, include
`no_tmp_dir=true` in the `file://` URL when the system temporary directory may
be on another filesystem.

For Archivematica deployments, this internal location is used for AIP deletion
reports while final AIPs are stored in Archivematica Storage Service.

For a3m deployments, the a3m worker writes completed AIPs directly to this
bucket through its own `[a3m.aipStaging]` configuration, and the storage
service later reads staged AIPs from this bucket. Both sections must point to
the same bucket root. For filesystem-backed storage, this means the configured
path must be mounted at the same path for both processes.

Staged a3m AIPs are written under the fixed `aips/` prefix inside
`[storage.internal]`. Configure the bucket URL as the root of the internal
storage bucket, not as the `aips/` directory itself.

#### Storage event listener

These settings configure [Redis] to act as an event listener and messaging
queue with the configured storage location, so that changes in the configured
storage location can be reflected in the Enduro user interface.

At this time Redis is the only supported event listener for Enduro.

**Example configuration**:

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
taskqueue = "am"
```

* `taskQueue`: In Temporal, a [Task Queue] is a lightweight, dynamically
  allocated queue that one or more workers can poll for tasks. Temporal supports
  many kinds of parallelization for a distributed application architecture, and
  and task queues can be namespaced to support such a set-up.

    In this case, the setting tells Enduro the name of the task queue to expect
    for workflow tasks provided by the configured preservation engine. Supported
    values are `am` for Archivematica (the default), or `a3m`.

### a3m configuration

The next two configuration blocks must be set if you intend to use [a3m] as the
[preservation engine] for ingest workflows.

The settings in this first block define basic connection information.

**Example configuration**:

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

#### a3m AIP staging bucket

These settings configure the ingest-domain view of the bucket where the a3m
worker writes completed AIPs before the storage domain reads them. Configure
this bucket with the shared
[bucket configuration options](#bucket-configuration-options), using
`[a3m.aipStaging]` as the bucket section. For Azure Blob Storage, use
`[a3m.aipStaging.azure]` as the matching authentication section.

For a3m deployments, this bucket must point to the same bucket root as
`[storage.internal]`. Staged AIPs are written under the fixed `aips/` prefix
inside the bucket.

For Archivematica deployments this bucket configuration is ignored.

**Example filesystem configuration**:

```toml
[a3m.aipStaging]
url = "file:///home/enduro/internal-storage/storage?metadata=skip&no_tmp_dir=true"
```

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

This section configures an internal upload and failed package storage space,
used by Enduro for [uploads via the user interface][upload-ui] as well as SIPs
or [PIPs][PIP] that encounter [content failures] or [system errors] during
ingest, so they can be downloaded for inspection via the user interface by
operators if desired.

Configure this bucket with the shared
[bucket configuration options](#bucket-configuration-options), using
`[internalStorage]` as the bucket section. For Azure Blob Storage, use
`[internalStorage.azure]` as the matching authentication section.

**Example filesystem configuration**:

```toml
[internalStorage]
url = "file:///home/enduro/internal-storage/ingest?metadata=skip&no_tmp_dir=true"
```

For filesystem-backed storage, Enduro writes API-uploaded SIPs and failed
packages to this bucket, and later reads them for workflow processing or
operator downloads. The process user must have read and write access. When the
API and worker processes run on different hosts or containers, the filesystem
path must refer to shared storage that is mounted at the same path for every
process that needs to read from or write to the bucket. Because Enduro writes
to this bucket through `fileblob`, include `no_tmp_dir=true` in the `file://`
URL when the system temporary directory may be on another filesystem.

### SIP source location configuration

The following settings configure a location to be used as a SIP source location
for SIP selection and ingest. For more information, see
[Add SIPs via a source location][sip-source].

The first set of parameters uniquely identify the source location in Enduro,
while the bucket subsection links the specified location.

!!! tip

    Once configured, ingests from a package in a configured source location can
    also be initiated via the [API].

**Example configuration**:

```toml
[sipsource]
id = "e6ddb29a-66d1-480e-82eb-fcfef1c825c5"
name = "Filesystem SIP Source"
```

* `id`: A UUID that unique identifies the SIP source location. Must be a valid
  [version 4 UUID].
* `name`: A human-readable name for the SIP source.

#### SIP source location bucket

Configure this bucket with the shared
[bucket configuration options](#bucket-configuration-options), using
`[sipsource.bucket]` as the bucket section. For Azure Blob Storage, use
`[sipsource.bucket.azure]` as the matching authentication section.

**Example filesystem configuration**:

```toml
[sipsource.bucket]
url = "file:///home/enduro/internal-storage/sip-source?metadata=skip"
```

For a filesystem-backed SIP source, the API lists objects from this bucket, and
workers download selected objects for ingest. Depending on SIP retention
settings, workers may also delete source objects after ingest, so the process
user must be able to list the directory contents, and read and delete files.
When the API and worker processes run on different hosts or containers, the
filesystem path must refer to shared storage that is mounted at the same path
for every process that needs to access the bucket.

Enduro does not write new source SIPs to this bucket through `fileblob`, so
`no_tmp_dir=true` is not needed for Enduro's SIP source runtime behavior.

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

### Child workflows

A child workflow is a concept borrowed from Enduro's [workflow engine],
[Temporal] where one workflow spawns an ancillary workflow. We use this model in
Enduro to keep custom activities with organization specific business rules or
policy requirements separated from the underlying general Enduro code.

Enduro schedules custom child workflows in the same Temporal namespace as their
parent workflows, as configured by `[temporal].namespace`. Workers that execute
these child workflows must connect to the same namespace.

All child workflow configurations share three common settings: `type`,
`taskQueue`, and `workflowName`.  An example configuration for the poststorage
child workflow is shown below.

!!! note

    The `childWorkflows` configuration block can be repeated more than once to
    add multiple child workflows. This is why the `childWorkflows` header is in
    double brackets rather than single brackets like other configuration blocks.

**Example configuration**:

```toml
[[childWorkflows]]
type = "poststorage"
taskQueue = "custom-queue"
workflowName = "custom-poststorage-workflow"
```

* `type`: The type of child workflow. Type is used to load the correct child
  workflow configuration in Enduro so only one child workflow of each type may
  be configured for any instance of Enduro. If multiple child workflows have
  the same type, an error will be thrown when the configuration is validated.
* `taskQueue`: In Temporal, a [Task Queue] is a lightweight, dynamically
  allocated queue that one or more workers can poll for tasks. The example
  value `custom-queue` differentiates this queue from Enduro's general `global`
  queue. The custom workers must poll the configured queue.
* `workflowName`: The name of the configured [child workflow] in Temporal. It
  must exactly match the custom workers registration.

All `[[childWorkflows]]` configuration sections are commented out at Enduro
installation, disabling the child workflows. Uncomment the relevant
configuration section to enable a child workflow, adjusting the settings as
necessary for your environment.

If you are running child workflows in development see the [Tilt environment
configuration] documentation for instructions on configuring Tilt to run the
child workflow workers and load any other required resources.

#### Preprocessing child workflow

The preprocessing child workflow is run before preservation processing by a
preservation engine (Archivematica or a3m). A preprocessing workflow can support
automated tasks such as validation and transformation of the SIP to meet
organization specific expectations and standards before preservation.

The preprocessing configuration supports two parameters in addition
to the common child workflow settings: `extract` and `sharedPath`.

**Example configuration**:

```toml
[[childWorkflows]]
type = "preprocessing"
taskQueue = "custom-queue"
workflowName = "custom-preprocessing-workflow"
extract = true
sharedPath = "/home/enduro/preprocessing"
```

* `extract`: Boolean value, set to `false` by default. This setting determines
  whether SIP extraction happens as part of the preprocessing child workflow or
  not. When set to false, SIP extraction will occur in the parent Enduro
  workflow before the child workflow is run. In some cases, you may wish to
  design custom child workflow activities that perform operations on the SIP
  before it is extracted: for example, calculating a checksum of the zipped
  package to check for prior duplicate ingests before proceeding. When this
  value is set to `true` Enduro will skip the extraction task at the beginning
  of the ingest workflow, allowing you to define when and how extraction occurs
  in the child workflow.
* `sharedPath`: The absolute path to the directory that Enduro should use to
  share the SIP between the primary ingest workflow and the configured
  preprocessing child workflow. A `sharedPath` is required for the preprocessing
  workflow because it requires access to the SIP for validation and
  transformation activities.

See the [Tilt environment configuration] documentation for instructions on
configuring Tilt to run the preprocessing worker and provision the shared path
volume in the Enduro development environment.

#### Poststorage child workflow

[Post-storage] is a phase in an ingest or preservation workflow describing all
the preservation policy-defined tasks performed on an [AIP] following
preservation processing and AIP storage. Post-storage task examples might
include metadata extraction and delivery to an external system (such as an
archival management system), AIP encryption or replication, and more.

**Example configuration**:

```toml
[[childWorkflows]]
type = "poststorage"
taskQueue = "custom-queue"
workflowName = "custom-poststorage-workflow"
```

#### Postbatch child workflow

A postbatch child workflow runs after multiple SIPs are ingested and stored
using the batch ingest functionality of Enduro. The postbatch workflow can be
used to report the results of a batch ingest, for instance creating a report
that collates SIP and AIP data for all of the SIPs in a batch.

**Example configuration**:

```toml
[[childWorkflows]]
type = "postbatch"
taskQueue = "custom-queue"
workflowName = "custom-postbatch-workflow"
```

[a3m]: https://github.com/artefactual-labs/a3m
[ABAC]: https://en.wikipedia.org/wiki/Attribute-based_access_control
[AIP]: ../user-manual/glossary.md#archival-information-package-aip
[AMSS documentation]: https://www.archivematica.org/docs/storage-service-latest/administrators
[API]: ../dev-manual/api.md
[Archivematica]: https://archivematica.org/
[Archivematica processing configuration fields]: https://archivematica.org/docs/latest/user-manual/administer/dashboard-admin/#processing-configuration-fields
[Archivematica transfer with existing checksums]: https://www.archivematica.org/en/docs/latest/user-manual/transfer/transfer/#create-a-transfer-with-existing-checksums
[Azure]: https://azure.microsoft.com/en-us/products/storage/blobs/
[BagIt]: https://www.rfc-editor.org/rfc/rfc8493
[bagit-gython]: https://github.com/artefactual-labs/bagit-gython
[bagit-gython README]: https://github.com/artefactual-labs/bagit-gython/blob/main/README.md
[child workflow]: ../user-manual/glossary.md#child-workflow
[components]: ../user-manual/components.md
[CORS]: https://en.wikipedia.org/wiki/Cross-origin_resource_sharing
[Dashboard configuration]: ../admin-manual/dashboard-config.md
[fileblob URL opener]: https://pkg.go.dev/gocloud.dev/blob/fileblob#URLOpener
[Gibibytes]: https://www.difference.wiki/gigabyte-vs-gibibyte/
[GoLang]: https://go.dev/
[METS]: https://www.loc.gov/standards/mets/
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
[S3-regions]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/s3-tables-regions-quotas.html#s3-tables-regions
[SI prefix]: https://en.wikipedia.org/wiki/Metric_prefix
[sip-source]: ../user-manual/ingest/submitting-content.md#initiate-ingest-using-sips-uploaded-to-a-source-location
[Storage Service]: https://archivematica.org/docs/storage-service-latest/
[system errors]: ../user-manual/glossary.md#system-error
[Task Queue]: https://docs.temporal.io/task-queue
[Temporal]: https://temporal.io
[tilt environment configuration]: ../dev-manual/devel.md#tilt-environment-configuration
[upload-ui]: ../user-manual/ingest/submitting-content.md#upload-sips-via-the-user-interface
[version 4 UUID]: https://www.rfc-editor.org/rfc/rfc9562.html#name-uuid-version-4
[watched-location]: ../user-manual/ingest/submitting-content.md#initiate-ingest-via-a-watched-location-upload
[workflow engine]: ../user-manual/glossary.md#workflow-engine
