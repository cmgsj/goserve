BIN := $(CURDIR)/bin
GOBIN ?= $(shell go env GOPATH)/bin
VERSION := $(shell git describe --tags --abbrev=0 | sed 's/^v//')

.PHONY: default
default: build install

.PHONY: build
build:
	@go build -trimpath -ldflags "-s -w -extldflags '-static' -X github.com/cmgsj/goserve/internal/version.version=$(VERSION)" -o $(BIN)/goserve ./cmd/goserve

.PHONY: install
install:
	@go install -trimpath -ldflags "-s -w -extldflags '-static' -X github.com/cmgsj/goserve/internal/version.version=$(VERSION)" ./cmd/goserve

.PHONY: uninstall
uninstall:
	@rm -f $(GOBIN)/goserve

.PHONY: clean
clean:
	@rm -rf $(BIN)
