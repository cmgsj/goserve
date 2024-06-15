#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

fmt() {
	go fmt ./...
	goimports -w -local github.com/cmgsj/goserve $(find . -type f -name "*.go" ! -path "./vendor/*")
}

test() {
	go test -v ./...
}

build() {
	binary build
}

install() {
	binary install
}

binary() {
	local cmd="$1"
	local version="$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//')"
	local ldflags="-s -w -extldflags='-static'"
	if [[ -n "$version" ]]; then
		ldflags+=" -X 'github.com/cmgsj/goserve/internal/version.version=$version'"
	fi
	local flags=(-trimpath -ldflags="$ldflags")
	if [[ "$cmd" == "build" ]]; then
		flags+=(-o "bin/goserve")
	fi
	echo "${cmd}ing goserve v$version $(go env GOOS)/$(go env GOARCH) cgo=$(go env CGO_ENABLED)"
	go "$cmd" "${flags[@]}" .
}
