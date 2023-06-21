 Dependency management

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

Update `golang.org/x/sync` individually to the latest version with:
`go get golang.org/x/sync` or `go get golang.org/x/sync@latest` (`v0.2.0`). This
is the preferred method.

Avoid `go get -u golang.org/x/sync` or `go get -u=patch golang.org/x/sync`
because it gets the latest versions of all the direct and indirect dependencies
of `golang.org/x/sync`.

Avoid `go get -u ./...` or `go get -u=patch ./...` because it gets the latest
versions of all the dependencies of ou rapplication.

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

    go get entgo.io/ent@v0.11.10
    go mod tidy

Edit `hack/make/dep_ent.mk` to update the binary installation:

    ENT_VERSION ?= 0.11.10

Now you can generate the code with:

    make gen-ent

### `goa.design/goa/v3`

Update the dependency:

    go get goa.design/goa/v3/cmd/goa@v3.11.3
    go mod tidy

Edit `hack/make/dep_goa.mk` to update the binary installation:

    GOA_VERSION ?= 0.11.10

Now you can generate the code with:

    make gen-goa

[Modules wiki page]: https://github.com/golang/go/wiki/Modules#how-to-upgrade-and-downgrade-dependencies
[major version suffix]: https://go.dev/ref/mod#major-version-suffixes
[gomajor]: https://github.com/icholy/gomajor
