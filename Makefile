all: build

build:
	go build .

install:
	go install .

uninstall:
	rm -f $$(go env GOPATH)/bin/goserve

clean:
	rm -f goserve
