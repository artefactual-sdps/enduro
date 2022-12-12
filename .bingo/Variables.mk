# Auto generated binary variables helper managed by https://github.com/bwplotka/bingo v0.7. DO NOT EDIT.
# All tools are designed to be build inside $GOBIN.
BINGO_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
GOPATH ?= $(shell go env GOPATH)
GOBIN  ?= $(firstword $(subst :, ,${GOPATH}))/bin
GO     ?= $(shell which go)

# Below generated variables ensure that every time a tool under each variable is invoked, the correct version
# will be used; reinstalling only if needed.
# For example for bingo variable:
#
# In your main Makefile (for non array binaries):
#
#include .bingo/Variables.mk # Assuming -dir was set to .bingo .
#
#command: $(BINGO)
#	@echo "Running bingo"
#	@$(BINGO) <flags/args..>
#
BINGO := $(GOBIN)/bingo-v0.6.0
$(BINGO): $(BINGO_DIR)/bingo.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/bingo-v0.6.0"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=bingo.mod -o=$(GOBIN)/bingo-v0.6.0 "github.com/bwplotka/bingo"

ENT := $(GOBIN)/ent-v0.11.0
$(ENT): $(BINGO_DIR)/ent.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/ent-v0.11.0"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=ent.mod -o=$(GOBIN)/ent-v0.11.0 "entgo.io/ent/cmd/ent"

GO_MOD_UPGRADE := $(GOBIN)/go-mod-upgrade-v0.9.1
$(GO_MOD_UPGRADE): $(BINGO_DIR)/go-mod-upgrade.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/go-mod-upgrade-v0.9.1"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=go-mod-upgrade.mod -o=$(GOBIN)/go-mod-upgrade-v0.9.1 "github.com/oligot/go-mod-upgrade"

GOA := $(GOBIN)/goa-v3.10.2
$(GOA): $(BINGO_DIR)/goa.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/goa-v3.10.2"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=goa.mod -o=$(GOBIN)/goa-v3.10.2 "goa.design/goa/v3/cmd/goa"

GOTESTSUM := $(GOBIN)/gotestsum-v1.8.2
$(GOTESTSUM): $(BINGO_DIR)/gotestsum.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/gotestsum-v1.8.2"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=gotestsum.mod -o=$(GOBIN)/gotestsum-v1.8.2 "gotest.tools/gotestsum"

MIGRATE := $(GOBIN)/migrate-v4.15.2
$(MIGRATE): $(BINGO_DIR)/migrate.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/migrate-v4.15.2"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -tags=mysql -mod=mod -modfile=migrate.mod -o=$(GOBIN)/migrate-v4.15.2 "github.com/golang-migrate/migrate/v4/cmd/migrate"

MOCKGEN := $(GOBIN)/mockgen-v1.6.0
$(MOCKGEN): $(BINGO_DIR)/mockgen.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/mockgen-v1.6.0"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=mockgen.mod -o=$(GOBIN)/mockgen-v1.6.0 "github.com/golang/mock/mockgen"

TPARSE := $(GOBIN)/tparse-v0.11.1
$(TPARSE): $(BINGO_DIR)/tparse.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/tparse-v0.11.1"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=tparse.mod -o=$(GOBIN)/tparse-v0.11.1 "github.com/mfridman/tparse"

