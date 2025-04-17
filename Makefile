SHELL := /bin/bash

MODULE := $$(go list -m)

.PHONY: default
default: tidy fmt generate build

.PHONY: tools
tools:
	@go -C tools install tool

.PHONY: update
update:
	@go -C tools get $$(go -C tools mod edit -json | jq -r '.Tool[].Path')
	@go get $$(go mod edit -json | jq -r '.Require[] | select(.Indirect | not) | .Path')
	@$(MAKE) tidy
	@$(MAKE) tools
	@$(MAKE) build

.PHONY: tidy
tidy:
	@go -C tools mod tidy
	@go -C tools mod download
	@go mod tidy
	@go mod download

.PHONY: fmt
fmt:
	@gofumpt -l -w -extra .

.PHONY: generate
generate:
	@go generate ./...

.PHONY: lint
lint:
	@go vet ./...
	@golangci-lint run ./...
	@govulncheck ./...
 
.PHONY: test
test:
	@go test -coverprofile=cover.out -race ./...

.PHONY: cover/html
cover/html: test
	@go tool cover -html=cover.out

.PHONY: cover/func
cover/func: test
	@go tool cover -func=cover.out

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
	echo "$${cmd}ing goserve@$${version}"; \
	CGO_ENABLED=0 go "$${cmd}" "$${flags[@]}" .

.PHONY: clean
clean:
	@rm -f bin/*
