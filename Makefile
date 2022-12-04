all: macm1

build: macm1 windows

macm1:
	GOOS=darwin GOARCH=arm64 go build -o bin/goserv cmd/main.go

windows:
	GOOS=windows GOARCH=amd64 go build -o bin/goserv.exe cmd/main.go

run:
	go run cmd/main.go

clean:
	rm -rf bin