$(call _assert_var,MAKEDIR)
$(call _conditional_include,$(MAKEDIR)/base.mk)
$(call _assert_var,CACHE_VERSIONS)
$(call _assert_var,CACHE_BIN)

WORKFLOWCHECK_VERSION ?= 0.2.0

WORKFLOWCHECK := $(CACHE_VERSIONS)/workflowcheck/$(WORKFLOWCHECK_VERSION)
$(WORKFLOWCHECK):
	rm -f $(CACHE_BIN)/workflowcheck
	mkdir -p $(CACHE_BIN)
	env GOBIN=$(CACHE_BIN) go install go.temporal.io/sdk/contrib/tools/workflowcheck@v$(WORKFLOWCHECK_VERSION)
	chmod +x $(CACHE_BIN)/workflowcheck
	rm -rf $(dir $(WORKFLOWCHECK))
	mkdir -p $(dir $(WORKFLOWCHECK))
	touch $(WORKFLOWCHECK)
