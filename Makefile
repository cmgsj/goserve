SHELL := /bin/bash
LIB := $(CURDIR)/hack/lib.sh

.PHONY: default
default: build install

.PHONY: fmt
fmt:
	@$(call exec, fmt)

.PHONY: test
test:
	@$(call exec, test)

.PHONY: build
build:
	@$(call exec, build)

.PHONY: install
install:
	@$(call exec, install)

define exec
source $(LIB) && $(@)
endef