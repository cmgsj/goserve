all: macm1

build: macm1 windows

macm1:
	GOOS=darwin GOARCH=arm64 go build -o bin .

windows:
	GOOS=windows GOARCH=amd64 go build -o bin .

publish: macm1
	sudo ln -sF /Users/cmg/dev/projects/goserve/bin/goserve /usr/local/bin

unpublish:
	sudo rm -f /usr/local/bin/goserve
	
run:
	go run .

clean:
	rm -f bin/*
