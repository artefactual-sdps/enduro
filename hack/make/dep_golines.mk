$(call _assert_var,MAKEDIR)
$(call _conditional_include,$(MAKEDIR)/base.mk)
$(call _assert_var,CACHE_VERSIONS)
$(call _assert_var,CACHE_BIN)

GOLINES_VERSION ?= 0.12.2

GOLINES := $(CACHE_VERSIONS)/golines/$(GOLINES_VERSION)
$(GOLINES):
	rm -f $(CACHE_BIN)/golines
	mkdir -p $(CACHE_BIN)
	env GOBIN=$(CACHE_BIN) go install github.com/segmentio/golines@v$(GOLINES_VERSION)
	chmod +x $(CACHE_BIN)/golines
	rm -rf $(dir $(GOLINES))
	mkdir -p $(dir $(GOLINES))
	touch $(GOLINES)
