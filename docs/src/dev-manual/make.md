# Makefile

The Makefile provides developer utility scripts via command line `make` tasks.

Running `make` with no arguments (or `make help`) prints the help message. Some
targets are described below with greater detail.

Dependencies are downloaded automatically.

## make gen-mock

`make gen-mock` generates mock interfaces with [mockgen](https://github.com/golang/mock).

New mocks can be added to the project by adding a new `mockgen` line to
`Makefile` that specifies the mock details (e.g. source, destination, package).

## make lint

`make lint` lints the project Go files with
[golangci-lint](https://github.com/golangci/golangci-lint).

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
