$(call _assert_var,MAKEDIR)
$(call _conditional_include,$(MAKEDIR)/base.mk)
$(call _assert_var,UNAME_OS)
$(call _assert_var,UNAME_ARCH2)
$(call _assert_var,CACHE_VERSIONS)
$(call _assert_var,CACHE_BIN)

ATLAS_VERSION ?= 0.31.0

ATLAS := $(CACHE_VERSIONS)/atlas/$(ATLAS_VERSION)
$(ATLAS):
	rm -f $(CACHE_BIN)/atlas
	mkdir -p $(CACHE_BIN)
	$(eval TMP := $(shell mktemp -d))
	$(eval OS := $(shell echo $(UNAME_OS) | tr A-Z a-z))
	curl -sSL \
		https://release.ariga.io/atlas/atlas-$(OS)-$(UNAME_ARCH2)-v$(ATLAS_VERSION) \
		> $(CACHE_BIN)/atlas
	chmod +x $(CACHE_BIN)/atlas
	rm -rf $(dir $(ATLAS))
	mkdir -p $(dir $(ATLAS))
	touch $(ATLAS)
