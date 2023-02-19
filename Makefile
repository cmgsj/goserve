all: run

run:
	go run ./cmd/goserve

test:
	go test -v ./...

build: macM1 windows

macM1:
	GOOS=darwin GOARCH=arm64 go build -o bin ./cmd/goserve

windows:
	GOOS=windows GOARCH=amd64 go build -o bin ./cmd/goserve

install:
	go install ./cmd/goserve

uninstall:
	rm -f $$(go env GOPATH)/bin/goserve

clean:
	rm -f bin/*
