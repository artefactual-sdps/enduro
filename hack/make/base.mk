$(call _assert_var,PROJECT)

UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)
CACHE_BASE ?= $(HOME)/.cache/$(PROJECT)
CACHE := $(CACHE_BASE)/$(UNAME_OS)/$(UNAME_ARCH)
CACHE_BIN := $(CACHE)/bin
CACHE_VERSIONS := $(CACHE)/versions
CACHE_GOBIN := $(CACHE)/gobin
CACHE_GOCACHE := $(CACHE)/gocache

export PATH := $(abspath $(CACHE_BIN)):$(PATH)
