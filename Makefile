all: macm1

build: macm1 windows

macm1:
	GOOS=darwin GOARCH=arm64 go build -o bin/goserve cmd/main.go

windows:
	GOOS=windows GOARCH=amd64 go build -o bin/goserve.exe cmd/main.go

run:
	go run cmd/main.go

clean:
	rm -rf bin