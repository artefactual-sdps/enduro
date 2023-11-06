# Testing

## `go test`

In Go, tests are executed with `go test`. Use `go help test` to see all the
options provided. A simple example:

    go test -v ./pkg/isr/...

With flag `-v`, `go test` prints the full output include logging.

### Disable caching

Go uses its cache of tests results unless the code was changed. It is possible
to force tests to run regardless the status of the cache with:

    go test -count=1 ./pkg/isr/...

### Race detector

Go comes with a race detector, often used in CI only since it slows down the
tests significantly:

    go test -race ./pkg/isr/...

You can use `make test-race` to achieve the same.

## `gotestsum`

For convenience, we use [gotestsum]. It is a `go test` runner that produces more
readable output and additional options, e.g.:

    gotestsum --format=testname ./pkg/isr

You can still pass `go test` flags to the runner, e.g.:

    gotestsum --format=testname ./pkg/isr -- -count=1

Our `make test` uses `gotestsum`.

## `tparse`

Similarly, we use [tparse] to produce coverage reports.

    $ go test -count=1 -json ./pkg/isr/... | tparse
    ┌─────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
    │  STATUS │ ELAPSED │                            PACKAGE                            │ COVER │ PASS │ FAIL │ SKIP  │
    │─────────┼─────────┼───────────────────────────────────────────────────────────────┼───────┼──────┼──────┼───────│
    │  PASS   │  0.00s  │ gitlab.artefactual.com/clients/ste/sdps_preprocessing/pkg/isr │  --   │  10  │  0   │  0    │
    └─────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

We provide `make tparse` with a set of predefined configuration flags.

[gotestsum]: https://github.com/gotestyourself/gotestsum
[tparse]: https://github.com/mfridman/tparse
