$(call _assert_var,MAKEDIR)
$(call _conditional_include,$(MAKEDIR)/base.mk)
$(call _assert_var,UNAME_OS)
$(call _assert_var,UNAME_ARCH)
$(call _assert_var,CACHE_VERSIONS)
$(call _assert_var,CACHE_BIN)

GO_ENUM_VERSION ?= 0.6.0

GO_ENUM := $(CACHE_VERSIONS)/go-enum/$(GO_ENUM_VERSION)
$(GO_ENUM):
	rm -f $(CACHE_BIN)/go-enum
	mkdir -p $(CACHE_BIN)
	curl -sSL \
		https://github.com/abice/go-enum/releases/download/v$(GO_ENUM_VERSION)/go-enum_$(UNAME_OS)_$(UNAME_ARCH) \
		> $(CACHE_BIN)/go-enum
	chmod +x $(CACHE_BIN)/go-enum
	rm -rf $(dir $(GO_ENUM))
	mkdir -p $(dir $(GO_ENUM))
	touch $(GO_ENUM)
