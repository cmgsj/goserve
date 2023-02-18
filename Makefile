all: macM1

build: macM1 windows

macM1:
	GOOS=darwin GOARCH=arm64 go build -o bin/goserve cmd/main.go

windows:
	GOOS=windows GOARCH=amd64 go build -o bin/goserve cmd/main.go

publish: macM1
	sudo ln -sF $${PWD}/bin/goserve /usr/local/bin

unpublish:
	sudo rm -f /usr/local/bin/goserve
	
run:
	go run cmd/main.go

clean:
	rm -f bin/*
