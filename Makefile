BIN := $(CURDIR)/bin
CMD := goserve
GOBIN ?= $(shell go env GOPATH)/bin
VERSION := $(shell git describe --tags --abbrev=0 | sed 's/^v//')

.PHONY: default
default: build install

.PHONY: build
build:
	@echo "building $(CMD) v$(VERSION)"
	@go build -trimpath -ldflags "-s -w -extldflags '-static' -X github.com/cmgsj/goserve/internal/version.version=$(VERSION)" -o $(BIN)/$(CMD) .

.PHONY: install
install:
	@echo "installing $(CMD) v$(VERSION)"
	@go install -trimpath -ldflags "-s -w -extldflags '-static' -X github.com/cmgsj/goserve/internal/version.version=$(VERSION)" .

.PHONY: uninstall
uninstall:
	@echo "uninstalling $(GOBIN)/$(CMD)"
	@rm -f $(GOBIN)/$(CMD)

.PHONY: clean
clean:
	@echo "removing $(BIN)/$(CMD)"
	@rm -rf $(BIN)/$(CMD)

.PHONY: mkcert
mkcert:
	@mkcert -cert-file tls/cert.pem -key-file tls/key.pem localhost
