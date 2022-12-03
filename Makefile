all: macm1

macm1:
	GOOS=darwin GOARCH=arm64 go build -o bin/

windows:
	GOOS=windows GOARCH=amd64 go build -o bin/

clean:
	rm -rf bin