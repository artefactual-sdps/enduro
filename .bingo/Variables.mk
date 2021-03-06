# Auto generated binary variables helper managed by https://github.com/bwplotka/bingo v0.5.2. DO NOT EDIT.
# All tools are designed to be build inside $GOBIN.
BINGO_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
GOPATH ?= $(shell go env GOPATH)
GOBIN  ?= $(firstword $(subst :, ,${GOPATH}))/bin
GO     ?= $(shell which go)

# Below generated variables ensure that every time a tool under each variable is invoked, the correct version
# will be used; reinstalling only if needed.
# For example for go-mod-upgrade variable:
#
# In your main Makefile (for non array binaries):
#
#include .bingo/Variables.mk # Assuming -dir was set to .bingo .
#
#command: $(GO_MOD_UPGRADE)
#	@echo "Running go-mod-upgrade"
#	@$(GO_MOD_UPGRADE) <flags/args..>
#
GO_MOD_UPGRADE := $(GOBIN)/go-mod-upgrade-v0.8.0
$(GO_MOD_UPGRADE): $(BINGO_DIR)/go-mod-upgrade.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/go-mod-upgrade-v0.8.0"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=go-mod-upgrade.mod -o=$(GOBIN)/go-mod-upgrade-v0.8.0 "github.com/oligot/go-mod-upgrade"

GOA := $(GOBIN)/goa-v3.7.10
$(GOA): $(BINGO_DIR)/goa.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/goa-v3.7.10"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=goa.mod -o=$(GOBIN)/goa-v3.7.10 "goa.design/goa/v3/cmd/goa"

GOLANGCI_LINT := $(GOBIN)/golangci-lint-v1.46.2
$(GOLANGCI_LINT): $(BINGO_DIR)/golangci-lint.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/golangci-lint-v1.46.2"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=golangci-lint.mod -o=$(GOBIN)/golangci-lint-v1.46.2 "github.com/golangci/golangci-lint/cmd/golangci-lint"

GOTESTSUM := $(GOBIN)/gotestsum-v1.8.1
$(GOTESTSUM): $(BINGO_DIR)/gotestsum.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/gotestsum-v1.8.1"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=gotestsum.mod -o=$(GOBIN)/gotestsum-v1.8.1 "gotest.tools/gotestsum"

MOCKGEN := $(GOBIN)/mockgen-v1.6.0
$(MOCKGEN): $(BINGO_DIR)/mockgen.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/mockgen-v1.6.0"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=mockgen.mod -o=$(GOBIN)/mockgen-v1.6.0 "github.com/golang/mock/mockgen"

TPARSE := $(GOBIN)/tparse-v0.11.1
$(TPARSE): $(BINGO_DIR)/tparse.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/tparse-v0.11.1"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=tparse.mod -o=$(GOBIN)/tparse-v0.11.1 "github.com/mfridman/tparse"

