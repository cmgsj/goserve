SHELL := /bin/bash
LIB := $(CURDIR)/hack/lib.sh

.PHONY: default
default: build install

.PHONY: build
build:
	@$(call exec, build)

.PHONY: install
install:
	@$(call exec, install)

define exec
source $(LIB) && $(@)
endef