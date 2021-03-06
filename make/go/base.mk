# Managed by makego. DO NOT EDIT.

# Must be set
$(call _assert_var,MAKEGO)
# Must be set
$(call _assert_var,MAKEGO_REMOTE)
# Must be set
$(call _assert_var,PROJECT)
# Must be set
$(call _assert_var,GO_MODULE)

UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)

ENV_DIR := .env
ENV_SH := $(ENV_DIR)/env.sh
ENV_BACKUP_DIR := $(HOME)/.config/$(PROJECT)/env

TMP := .tmp

EXTRA_MAKEGO_FILES := $(EXTRA_MAKEGO_FILES) scripts/checknodiffgenerated.bash

# Settable
FILE_IGNORES := $(FILE_IGNORES) $(ENV_DIR)/ $(TMP)/
# Settable
CACHE_BASE ?= $(HOME)/.cache/$(PROJECT)

CACHE := $(CACHE_BASE)/$(UNAME_OS)/$(UNAME_ARCH)
CACHE_BIN := $(CACHE)/bin
CACHE_INCLUDE := $(CACHE)/include
CACHE_VERSIONS := $(CACHE)/versions
CACHE_ENV := $(CACHE)/env
CACHE_GO := $(CACHE)/go

# Runtime MAKEGOALL

export GO111MODULE := on
ifdef GOPRIVATE
export GOPRIVATE := $(GOPRIVATE),$(GO_MODULE)
else
export GOPRIVATE := $(GO_MODULE)
endif
export GOPATH := $(abspath $(CACHE_GO))
export GOBIN := $(abspath $(CACHE_BIN))
export PATH := $(GOBIN):$(PATH)

print-%:
	@echo $($*)

.PHONY: envbackup
envbackup:
	rm -rf "$(ENV_BACKUP_DIR)"
	mkdir -p "$(dir $(ENV_BACKUP_DIR))"
	cp -R "$(ENV_DIR)" "$(ENV_BACKUP_DIR)"

.PHONY: envrestore
envrestore:
	@ if [ ! -d "$(ENV_BACKUP_DIR)" ]; then echo "no backup stored in $(ENV_BACKUP_DIR)"; exit 1; fi
	rm -rf "$(ENV_DIR)"
	cp -R "$(ENV_BACKUP_DIR)" "$(ENV_DIR)"

.PHONY: direnv
direnv:
	@mkdir -p $(CACHE_ENV)
	@rm -f $(CACHE_ENV)/env.sh
	@echo 'export CACHE="$(abspath $(CACHE))"' >> $(CACHE_ENV)/env.sh
	@echo 'export GO111MODULE="$(GO111MODULE)"' >> $(CACHE_ENV)/env.sh
	@echo 'export GOPRIVATE="$(GOPRIVATE)"' >> $(CACHE_ENV)/env.sh
	@echo 'export GOPATH="$(GOPATH)"' >> $(CACHE_ENV)/env.sh
	@echo 'export GOBIN="$(GOBIN)"' >> $(CACHE_ENV)/env.sh
	@echo 'export PATH="$(GOBIN):$${PATH}"' >> $(CACHE_ENV)/env.sh
	@echo '[ -f "$(abspath $(ENV_SH))" ] && . "$(abspath $(ENV_SH))"' >> $(CACHE_ENV)/env.sh
	@echo $(CACHE_ENV)/env.sh

.PHONY: clean
clean:
	git clean -xdf -e /$(ENV_DIR)/

.PHONY: cleancache
cleancache:
	rm -rf $(CACHE_BASE)

.PHONY: nuke
nuke: clean cleancache
	sudo rm -rf $(CACHE_GO)/pkg/mod

.PHONY: dockerdeps
dockerdeps::

.PHONY: deps
deps:: dockerdeps

.PHONY: pregenerate
pregenerate::

.PHONY: postgenerate
postgenerate::

.PHONY: licensegenerate
licensegenerate::

.PHONY: generate
generate:
	@$(MAKE) pregenerate
	@$(MAKE) postgenerate
	@$(MAKE) licensegenerate

.PHONY: checknodiffgenerated
checknodiffgenerated:
	@ if [ -d .git ]; then \
			$(MAKE) __checknodiffgeneratedinternal; \
		else \
			echo "skipping make checknodiffgenerated due to no .git repository" >&2; \
		fi

.PHONY: updatemakego
updatemakego:
ifndef DESTRUCTIVE
	$(error Set DESTRUCTIVE=1 to acknowledge this is potentially destructive to your current makego files)
else
	@rm -rf $(TMP)/makego
	@mkdir -p $(TMP)
	git clone $(MAKEGO_REMOTE) $(TMP)/makego
	rm -rf $(MAKEGO)
ifdef MAKEGOALL
	cp -R $(TMP)/makego/make/go $(MAKEGO)
else
	mkdir -p $(MAKEGO)
	$(foreach makego_file,$(subst $(MAKEGO)/,,$(filter $(MAKEGO)/%.mk,$(MAKEFILE_LIST))),cp $(TMP)/makego/make/go/$(makego_file) $(MAKEGO)/$(makego_file); )
	$(foreach extra_makego_file,$(sort $(EXTRA_MAKEGO_FILES)),mkdir -p $(MAKEGO)/$(dir $(extra_makego_file)); cp $(TMP)/makego/make/go/$(extra_makego_file) $(MAKEGO)/$(extra_makego_file); )
endif
	@rm -rf $(TMP)/makego
endif

.PHONY: copytomakego
copytomakego:
	@rm -rf $(TMP)/makego
	@mkdir -p $(TMP)
	git clone $(MAKEGO_REMOTE) $(TMP)/makego
	$(foreach makego_file,$(subst $(MAKEGO)/,,$(shell find $(MAKEGO) -type f)),mkdir -p $(TMP)/makego/make/go/$(dir $(makego_file)); cp $(MAKEGO)/$(makego_file) $(TMP)/makego/make/go/$(makego_file); )
	@cd $(TMP)/makego; git status; git diff
	@echo cd $(TMP)/makego
	@echo git diff
	@echo git push origin master
	@echo cd -

.PHONY: initmakego
initmakego::

.PHONY: updategitignores
updategitignores:
	@rm -f .gitignore
	@echo '# Autogenerated by makego. DO NOT EDIT.' > .gitignore
	@$(foreach file_ignore,$(sort $(FILE_IGNORES)),echo /$(file_ignore) >> .gitignore || exit 1; )

pregenerate:: updategitignores

.PHONY: __checknodiffgeneratedinternal
__checknodiffgeneratedinternal:
	bash $(MAKEGO)/scripts/checknodiffgenerated.bash $(MAKE) generate
