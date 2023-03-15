BINARY_NAME=goserve
VERSION=1.0.0
VERSION_PATH=github.com/cmgsj/goserve/pkg/cmd/root.version
BUILD_DIR=./bin

all: install

build: mac windows

mac:
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X '${VERSION_PATH}=${VERSION}'" -o ${BUILD_DIR}/${BINARY_NAME}-mac ./cmd/${BINARY_NAME}

windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "-X '${VERSION_PATH}=${VERSION}'" -o ${BUILD_DIR}/${BINARY_NAME}-win.exe ./cmd/${BINARY_NAME}

install:
	go install -ldflags "-X '${VERSION_PATH}=${VERSION}'" ./cmd/${BINARY_NAME}

uninstall:
	rm -f $$(go env GOPATH)/bin/${BINARY_NAME}

clean:
	rm -f ${BUILD_DIR}/*
