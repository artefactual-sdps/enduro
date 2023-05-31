$(call _assert_var,MAKEDIR)
$(call _conditional_include,$(MAKEDIR)/base.mk)
$(call _assert_var,UNAME_OS)
$(call _assert_var,UNAME_ARCH)
$(call _assert_var,CACHE_VERSIONS)
$(call _assert_var,CACHE_BIN)

# Keep in sync with the ent version in go.mod.
# See https://entgo.io/docs/code-gen/#version-compatibility-between-entc-and-ent
ENT_VERSION ?= 0.12.3

ENT := $(CACHE_VERSIONS)/ent/$(ENT_VERSION)
$(ENT):
	@rm -f $(CACHE_BIN)/ent
	@mkdir -p $(CACHE_BIN)
	@env GOBIN=$(CACHE_BIN) go install entgo.io/ent/cmd/ent@v$(ENT_VERSION)
	@chmod +x $(CACHE_BIN)/ent
	@rm -rf $(dir $(ENT))
	@mkdir -p $(dir $(ENT))
	@touch $(ENT)
