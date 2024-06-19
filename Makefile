.PHONY: default
default: fmt build install

.PHONY: fmt
fmt:
	@go fmt ./...
	@goimports -w -local github.com/cmgsj/goserve $$(find . -type f -name "*.go" ! -path "./vendor/*")

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
	version="$$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//')"; \
	ldflags="-s -w -extldflags='-static'"; \
	if [[ -n "$${version}" ]]; then \
		ldflags+=" -X 'github.com/cmgsj/goserve/pkg/cmd/goserve.v=$${version}'"; \
	fi; \
	flags=(-trimpath -ldflags="$${ldflags}"); \
	if [[ "$${cmd}" == "build" ]]; then \
		flags+=(-o "bin/goserve"); \
	fi; \
	echo "$${cmd}ing goserve v$${version} $$(go env GOOS)/$$(go env GOARCH) cgo=$$(go env CGO_ENABLED)"; \
	go "$${cmd}" "$${flags[@]}" .
