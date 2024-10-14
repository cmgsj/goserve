SHELL := /bin/bash

MODULE := $$(go list -m)

.PHONY: default
default: fmt install

.PHONY: fmt
fmt:
	@find . -type f -name "*.go" ! -path "./vendor/*" | while read -r file; do \
		go fmt "$${file}" 2>&1 | grep -v "is a program, not an importable package"; \
		goimports -w -local $(MODULE) "$${file}"; \
	done

.PHONY: test
test:
	@go test -v ./...

.PHONY: build
build:
	@$(MAKE) binary cmd=build

.PHONY: install
install:
	@$(MAKE) binary cmd=install

.PRONY: binary
binary:
	@if [[ -z "$${cmd}" ]]; then \
		echo "must set cmd env var"; \
		exit 1; \
	fi; \
	if [[ "$${cmd}" != "build" && "$${cmd}" != "install" ]]; then \
		echo "unknown cmd '$${cmd}'"; \
		exit 1; \
	fi; \
	if [[ -z "$${version}" ]]; then \
		version="$$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//')"; \
	fi; \
	ldflags="-s -w -extldflags='-static'"; \
	if [[ -n "$${version}" ]]; then \
		ldflags+=" -X '$(MODULE)/pkg/cmd/goserve.version=$${version}'"; \
	fi; \
	flags=(-trimpath -ldflags="$${ldflags}"); \
	if [[ "$${cmd}" == "build" ]]; then \
		flags+=(-o "bin/goserve"); \
	fi; \
	echo "$${cmd}ing goserve@$${version} $$(go env GOOS)/$$(go env GOARCH) cgo=$$(go env CGO_ENABLED)"; \
	go "$${cmd}" "$${flags[@]}" .
