$(call _assert_var,MAKEDIR)
$(call _conditional_include,$(MAKEDIR)/base.mk)
$(call _assert_var,UNAME_OS2)
$(call _assert_var,UNAME_ARCH2)
$(call _assert_var,CACHE_VERSIONS)
$(call _assert_var,CACHE_BIN)

SHFMT_VERSION ?= 3.7.0

SHFMT := $(CACHE_VERSIONS)/shfmt/$(SHFMT_VERSION)
$(SHFMT):
	rm -f $(CACHE_BIN)/shfmt
	mkdir -p $(CACHE_BIN)
	curl -sSL https://github.com/mvdan/sh/releases/download/v$(SHFMT_VERSION)/shfmt_v$(SHFMT_VERSION)_$(UNAME_OS2)_$(UNAME_ARCH2) > $(CACHE_BIN)/shfmt
	chmod +x $(CACHE_BIN)/shfmt
	rm -rf $(dir $(SHFMT))
	mkdir -p $(dir $(SHFMT))
	touch $(SHFMT)
