GOBIN ?= $(shell go env GOPATH)/bin

.PHONY: default
default: build install

.PHONY: build
build:
	@go build ./cmd/goserve

.PHONY: install
install:
	@go install ./cmd/goserve

.PHONY: uninstall
uninstall:
	@rm -f $(GOBIN)/goserve

.PHONY: clean
clean:
	@rm -rf ./goserve
