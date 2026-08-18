# Dependency management

## Update dependencies individually

It is safer to update dependencies individually rather than all at once.

Explained in the [Modules wiki page] at great detail, this is the way to list
available minor and patch upgrades only for our direct dependencies:

    go list -u -f '{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}' -m all 2> /dev/null
    entgo.io/ent: v0.11.8 -> v0.12.3
    github.com/aws/aws-sdk-go-v2/config: v1.18.22 -> v1.18.25
    github.com/aws/aws-sdk-go-v2/credentials: v1.13.21 -> v1.13.24
    github.com/aws/aws-sdk-go-v2/service/s3: v1.33.0 -> v1.33.1
    github.com/go-logr/zapr: v1.2.3 -> v1.2.4
    github.com/nyudlts/go-bagit: v0.2.0-alpha -> v0.2.1-alpha
    github.com/redis/go-redis/v9: v9.0.3 -> v9.0.4
    go.temporal.io/sdk: v1.22.1 -> v1.22.2
    golang.org/x/exp: v0.0.0-20230124195608-d38c7dcee874 -> v0.0.0-20230515195305-f3d0a9c9a5cc
    golang.org/x/sync: v0.1.0 -> v0.2.0
    google.golang.org/grpc: v1.54.0 -> v1.55.0

Alternatively, use `make deps` (uses `go-mod-outdated`, managed with bine):

    +---------------------+---------+-------------+--------+------------------+
    |       MODULE        | VERSION | NEW VERSION | DIRECT | VALID TIMESTAMPS |
    +---------------------+---------+-------------+--------+------------------+
    | ariga.io/atlas      | v0.37.0 | v1.3.0      | true   | true             |
    | entgo.io/ent        | v0.14.5 | v0.14.6     | true   | true             |
    | github.com/pkg/sftp | v1.13.6 | v1.13.11    | true   | true             |
    +---------------------+---------+-------------+--------+------------------+

For example, update `golang.org/x/sync` individually to the latest version with:
`go get golang.org/x/sync` or `go get golang.org/x/sync@latest` (`v0.2.0`). This
is the preferred method.

Avoid `go get -u golang.org/x/sync` or `go get -u=patch golang.org/x/sync`
because it gets the latest versions of all the direct and indirect dependencies
of `golang.org/x/sync`.

Avoid `go get -u ./...` or `go get -u=patch ./...` because it gets the latest
versions of all the dependencies of our application.

## Major dependency updates

If a module is released at major version v2 or higher, its path must have a
[major version suffix]. These are some examples from our `go.mod`:

    github.com/alicebob/miniredis/v2 v2.30.2
    github.com/mholt/archiver/v3 v3.5.1
    github.com/redis/go-redis/v9 v9.0.3

Go chose this model to discourage backward-incompatible changes. They make it
comparable to using a different dependency and that is why the module path must
be different.

While dealing with this type of module update requires more care, tools like
[gomajor] can automate some parts of the process.

## Special dependencies

### `entgo.io/ent`

Update the dependency:

    go get entgo.io/ent@vX.Y.Z
    go mod tidy

Update the `ent` binary version in `.bine.toml` to match:

    [[bins]]
    name = "ent"
    go_package = "entgo.io/ent/cmd/ent"
    version = "X.Y.Z"

Now you can generate the code with:

    make gen-ent

### `goa.design/goa/v3`

Update Goa and its plugins:

    go get goa.design/goa/v3@vX.Y.Z
    go get goa.design/plugins/v3@vX.Y.Z
    go mod tidy

Update the `goa` binary version in `.bine.toml` to match:

    [[bins]]
    name = "goa"
    go_package = "goa.design/goa/v3/cmd/goa"
    version = "X.Y.Z"

Now you can generate the code with:

    make gen-goa

## Version updates

Dependabot handles regular version updates in addition to security updates. Its
configuration groups related updates and limits the number of open pull requests
for each package ecosystem and directory to keep the volume manageable.

The following sections describe how to identify version updates in different
areas of the project.

### Automated process

Dependabot checks the package ecosystems configured in
`.github/dependabot.yml`. Most checks run weekly, while the main GitHub Actions
check runs daily. Review the resulting grouped pull requests, their release
notes, and the CI results before merging them.

### Manual process

#### Go dependencies

    make deps

#### Development binaries

Review the versions in `.bine.toml` manually. Bine manages the installation of
development binaries such as Ent, Goa, and `go-mod-outdated`.

#### Dashboard dependencies

Run:

    cd dashboard
    npm ci
    npm run deps

#### Node runtime used by the dashboard

Treat Node as a runtime dependency managed in a few coordinated places:

* `/.node-version`: exact Node version used by local tooling and GitHub Actions
* `/dashboard/Dockerfile`: exact Node version used by the dashboard build image
* `/dashboard/package.json` (`engines.node`): supported Node major version,
  enforced by `/dashboard/.npmrc`
* `/dashboard/package.json` and `/dashboard/package-lock.json`: Node-related
  TypeScript packages such as `@tsconfig/node26` and `@types/node`

Typical Node update:

1. Update `/.node-version`.
2. Update `/dashboard/Dockerfile`.
3. Update `engines.node` if the supported major version changes.
4. Review `@tsconfig/nodeXX` and `@types/node`.
5. Refresh `/dashboard/package-lock.json`.
6. Re-run `npm run type-check`, `npm run build`, `npm test`, and `npm run lint`.

#### GitHub Actions

Review manually, e.g.:

    git grep "uses: " .github/workflows/

#### Dockerfile

Review manually in:

* `/Dockerfile`
* `/dashboard/Dockerfile`

What to look for?

* Base images
* Other dependencies

#### Pulumi

    hack/pulumi/go.mod

#### Services

Review manually, under `hack/kube/...`.

[Modules wiki page]: https://github.com/golang/go/wiki/Modules#how-to-upgrade-and-downgrade-dependencies
[major version suffix]: https://go.dev/ref/mod#major-version-suffixes
[gomajor]: https://github.com/icholy/gomajor
