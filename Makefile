SHELL := /bin/bash

MODULE := $$(go list -m)

.PHONY: default
default: tidy fmt generate build

.PHONY: upgrade
upgrade:
	@go list -m -f '{{if and (not .Main) (not .Indirect)}}{{.Path}}{{end}}' all | xargs go get; \
	$(MAKE) tidy

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: fmt
fmt:
	@go fmt ./...; \
	go tool goimports -w -local $(MODULE) .

.PHONY: generate
generate:
	@go generate ./...

.PHONY: lint
lint:
	@go vet ./...; \
	golangci-lint run ./...; \
	govulncheck ./...
 
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
	go mod download; \
	go "$${cmd}" "$${flags[@]}" .
