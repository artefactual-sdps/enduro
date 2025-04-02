# Makefile

The Makefile provides developer utility scripts via command line `make` tasks.

Running `make` with no arguments (or `make help`) prints the help message. Some
targets are described below with greater detail.

Dependencies are downloaded automatically.

## Debug mode

The debug mode produces more output, including the commands executed. E.g.:

```shell
$ make env DBG_MAKEFILE=1
Makefile:10: ***** starting Makefile for goal(s) "env"
Makefile:11: ***** Fri 10 Nov 2023 11:16:16 AM CET
go env
GO111MODULE=''
GOARCH='amd64'
...
```

## Targerts

### make gen-mock

`make gen-mock` generates mock interfaces with [mockgen](https://github.com/golang/mock).

New mocks can be added to the project by adding a new `mockgen` line to
`Makefile` that specifies the mock details (e.g. source, destination, package).

### make fmt

`make fmt` lints the project Go files with
[golangci-lint](https://github.com/golangci/golangci-lint).

### make lint

`make lint` lints the project Go files with
[golangci-lint](https://github.com/golangci/golangci-lint).

### make test

`make test` runs all tests and outputs a summary of the results using
[gotestsum](https://pkg.go.dev/gotest.tools/gotestsum).

It accepts a couple of arguments to customize its behavior:

- `TFORMAT` adjusts the output format used by `gotestsum` (default: `short`),
- `GOTEST_FLAGS` adjuts the flags sent to `go test`. Run `go help testflag` to
  know more about the available flags.

Some usage examples:

```shell
# Use default configuration.
$ make test

# Enable coverage reporting.
$ make test GOTEST_FLAGS=-cover

# Run a specific test with the race detector.
$ make test GOTEST_FLAGS="-race -run=^TestSameOriginChecker"

# A bit of everything.
$ make test GOTEST_FLAGS="-cover -race" TFORMAT=standard-verbose
```

### make test-race

`make test-race` does `make test` but with the
[Go Race Detector](https://go.dev/blog/race-detector) enabled to flag race
conditions.

Alternatively, you can run:

```shell
make test GOTEST_FLAGS=-race
```

### make test-ci

`make test-ci` does `make test` but with the
[Go Race Detector](https://go.dev/blog/race-detector) enabled to flag race
conditions and coverage reporting enabled. This is meant to be used in CI.

### make tparse

`make tparse` runs all tests and outputs detailed results using
[tparse](https://github.com/mfridman/tparse).
