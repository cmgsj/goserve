BIN := $(CURDIR)/bin
CGO_ENABLED ?= 0
GOOS ?= darwin linux windows
GOARCH ?= amd64 arm64
GOBIN ?= $(shell go env GOPATH)/bin
VERSION := $(shell git describe --tags --abbrev=0)

.PHONY: default
default: build

.PHONY: build
build:
	@for goos in $(GOOS); do \
		for goarch in $(GOARCH); do \
		out="$(BIN)/$$goos-$$goarch/goserve" ; \
		if [ $$goos = "windows" ]; then out="$$out.exe" ; fi ; \
		GOOS=$$goos GOARCH=$$goarch go build \
			-ldflags "-X github.com/cmgsj/goserve/internal/version.version=$(VERSION)" -o $$out; \
		done \
	done

.PHONY: install
install:
	@go install -ldflags "-X github.com/cmgsj/goserve/internal/version.version=$(VERSION)"

.PHONY: uninstall
uninstall:
	@rm -f $(GOBIN)/goserve

.PHONY: clean
clean:
	@rm -rf $(BIN)
