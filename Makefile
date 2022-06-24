include .bingo/Variables.mk

SHELL=/bin/bash
BUILD_TIME=$(shell date -u +%Y-%m-%dT%T%z)
GIT_COMMIT=$(shell git rev-parse --short HEAD)
LD_FLAGS= '-X "github.com/artefactual-labs/enduro/internal/version.BuildTime=$(BUILD_TIME)" -X github.com/artefactual-labs/enduro/internal/version.GitCommit=$(GIT_COMMIT)'
GO_FLAGS= -ldflags=$(LD_FLAGS)

define NEWLINE


endef

IGNORED_PACKAGES := \
	github.com/artefactual-labs/enduro/hack/genpkgs \
	github.com/artefactual-labs/enduro/internal/amclient/fake \
	github.com/artefactual-labs/enduro/internal/api/design \
	github.com/artefactual-labs/enduro/internal/api/gen/batch \
	github.com/artefactual-labs/enduro/internal/api/gen/package_ \
	github.com/artefactual-labs/enduro/internal/api/gen/package_/views \
	github.com/artefactual-labs/enduro/internal/api/gen/http/batch/client \
	github.com/artefactual-labs/enduro/internal/api/gen/http/batch/server \
	github.com/artefactual-labs/enduro/internal/api/gen/http/cli/enduro \
	github.com/artefactual-labs/enduro/internal/api/gen/http/package_/client \
	github.com/artefactual-labs/enduro/internal/api/gen/http/package_/server \
	github.com/artefactual-labs/enduro/internal/api/gen/http/swagger/client \
	github.com/artefactual-labs/enduro/internal/api/gen/http/swagger/server \
	github.com/artefactual-labs/enduro/internal/api/gen/swagger \
	github.com/artefactual-labs/enduro/internal/batch/fake \
	github.com/artefactual-labs/enduro/internal/package_/fake \
	github.com/artefactual-labs/enduro/internal/temporal/testutil \
	github.com/artefactual-labs/enduro/internal/watcher/fake
PACKAGES		:= $(shell go list ./...)
TEST_PACKAGES	:= $(filter-out $(IGNORED_PACKAGES),$(PACKAGES))

export PATH:=$(GOBIN):$(PATH)

.DEFAULT_GOAL := run

$(GOBIN)/bingo:
	$(GO) install github.com/bwplotka/bingo@latest

bingo: $(GOBIN)/bingo

tools: bingo
	bingo get
	bingo list

run: enduro-dev
	./build/enduro

build: enduro-dev enduro-a3m-worker-dev

enduro-dev:
	mkdir -p ./build
	$(GO) build -trimpath -o build/enduro $(GO_FLAGS) -v .

enduro-a3m-worker-dev:
	mkdir -p ./build
	$(GO) build -trimpath -o build/enduro-a3m-worker $(GO_FLAGS) -v ./cmd/enduro-a3m-worker

test:
	@$(GOTESTSUM) $(TEST_PACKAGES)

test-race:
	@$(GOTESTSUM) $(TEST_PACKAGES) -- -race

ignored:
	$(foreach PACKAGE,$(IGNORED_PACKAGES),@echo $(PACKAGE)$(NEWLINE))

lint:
	$(GOLANGCI_LINT) run -v --timeout=5m --fix

gen-goa:
	$(GOA) gen github.com/artefactual-labs/enduro/internal/api/design -o internal/api

clean:
	rm -rf ./build ./dist

release-test-config:
	$(GORELEASER) --snapshot --skip-publish --rm-dist

release-test:
	$(GORELEASER) --skip-publish

PROJECT := enduro
UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)
CACHE_BASE := $(HOME)/.cache/$(PROJECT)
CACHE := $(CACHE_BASE)/$(UNAME_OS)/$(UNAME_ARCH)
CACHE_BIN := $(CACHE)/bin
export PATH := $(abspath $(CACHE_BIN)):$(PATH)
CACHE_VERSIONS := $(CACHE)/versions
HUGO_VERSION := 0.90.0
HUGO := $(CACHE_VERSIONS)/hugo/$(HUGO_VERSION)
$(HUGO):
	@rm -f $(CACHE_BIN)/hugo
	@mkdir -p $(CACHE_BIN)
	curl -sSL "https://github.com/gohugoio/hugo/releases/download/v$(HUGO_VERSION)/hugo_extended_$(HUGO_VERSION)_Linux-64bit.tar.gz" | tar xzf - -C "$(CACHE_BIN)"
	chmod +x "$(CACHE_BIN)/hugo"
	@rm -rf $(dir $(HUGO))
	@mkdir -p $(dir $(HUGO))
	@touch $(HUGO)

website: $(HUGO)
	hugo serve --source=website/

gen-dashboard-client:
	@rm -rf $(CURDIR)/dashboard/src/openapi-generator
	@docker container run --rm --user $(shell id -u):$(shell id -g) --volume $(CURDIR):/local openapitools/openapi-generator-cli:v6.0.0 \
		generate \
			--input-spec /local/internal/api/gen/http/openapi.json \
			--generator-name typescript-fetch \
			--output /local/dashboard/src/openapi-generator/ \
			--skip-validate-spec \
			-p "generateAliasAsModel=false" \
			-p "withInterfaces=true" \
			-p "supportsES6=true"
	@echo "@@@@ Please, review all warnings generated by openapi-generator-cli above!"
	@echo "@@@@ We're using \`--skip-validate-spec\` to deal with Goa spec generation issues."

gen-mock:
	$(MOCKGEN) -destination=./internal/batch/fake/mock_batch.go -package=fake github.com/artefactual-labs/enduro/internal/batch Service
	$(MOCKGEN) -destination=./internal/package_/fake/mock_package_.go -package=fake github.com/artefactual-labs/enduro/internal/package_ Service
	$(MOCKGEN) -destination=./internal/storage/fake/mock_storage.go -package=fake github.com/artefactual-labs/enduro/internal/storage Service
	$(MOCKGEN) -destination=./internal/watcher/fake/mock_watcher.go -package=fake github.com/artefactual-labs/enduro/internal/watcher Service

.PHONY: *
