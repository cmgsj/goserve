all: build

build:
	go build -o ./bin/goserve ./cmd/goserve

install:
	go install ./cmd/goserve

uninstall:
	rm -f $$(go env GOPATH)/bin/goserve

clean:
	rm -f ./bin/*
