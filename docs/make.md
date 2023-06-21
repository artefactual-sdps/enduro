# Makefile

The Makefile provides developer utility scripts via command line `make` tasks.

Running `make` with no arguments (or `make help`) prints the help message. Some
targets are described below with greater detail.

## make tools

`make tools` installs all developer binaries
managed by [bingo](https://github.com/bwplotka/bingo), and outputs a list of
these binaries and their installed versions.

## make gen-mock

`make gen-mock` generates mock interfaces with [mockgen](https://github.com/golang/mock).

New mocks can be added to the project by adding a new `mockgen` line to `Makefile` that specifies the mock details (e.g. source, destination, package).

## make golangci-lint

`make golangci-lint` lints the project Go files with
[golangci-lint](https://github.com/golangci/golangci-lint).

If `golangci-lint` is not installed with the expected version then the binary
will be downloaded and installed before it is run.

## make lint

`make lint` is an alias for `make golangci-lint`.

## make test

`make test` runs all tests and outputs a summary of the results using
[gotestsum](https://pkg.go.dev/gotest.tools/gotestsum).

## make test-race

`make test-race` does `make test` but with the
[Go Race Detector](https://go.dev/blog/race-detector) enabled to flag race
conditions.

## make tparse

`make tparse` runs all tests and outputs detailed results using
[tparse](https://github.com/mfridman/tparse).
