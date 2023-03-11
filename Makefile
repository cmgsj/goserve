all: install

build: mac windows

mac:
	GOOS=darwin GOARCH=arm64 go build -o ./bin/goserve-mac ./cmd/goserve

windows:
	GOOS=windows GOARCH=amd64 go build -o ./bin/goserve-win.exe ./cmd/goserve

install:
	go install ./cmd/goserve

uninstall:
	rm -f $$(go env GOPATH)/bin/goserve

clean:
	rm -f bin/*
