$(call _assert_var,MAKEDIR)
$(call _conditional_include,$(MAKEDIR)/base.mk)
$(call _assert_var,UNAME_OS2)
$(call _assert_var,UNAME_ARCH2)
$(call _assert_var,CACHE_VERSIONS)
$(call _assert_var,CACHE_BIN)

MIGRATE_VERSION ?= 4.16.2

MIGRATE := $(CACHE_VERSIONS)/migrate/$(MIGRATE_VERSION)
$(MIGRATE):
	@rm -f $(CACHE_BIN)/migrate
	@mkdir -p $(CACHE_BIN)
	@$(eval TMP := $(shell mktemp -d))
	@curl -sSL \
		https://github.com/golang-migrate/migrate/releases/download/v$(MIGRATE_VERSION)/migrate.$(UNAME_OS2)-$(UNAME_ARCH2).tar.gz \
		| tar xz -C $(TMP)
	@mv $(TMP)/migrate $(CACHE_BIN)/
	@chmod +x $(CACHE_BIN)/migrate
	@rm -rf $(dir $(MIGRATE))
	@mkdir -p $(dir $(MIGRATE))
	@touch $(MIGRATE)
