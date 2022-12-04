package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"goserve/pkg/handler"
	"goserve/pkg/middleware"
	"goserve/pkg/util"
	"net"
	"net/http"
	"os"
	"path"
)

var (
	root     = flag.String("root", ".", "root path to serve")
	port     = flag.Int("port", 1234, "port to listen on")
	download = flag.Bool("download", false, "serve files to download (true), plain text (false) (default: false)")
)

func main() {
	if err := Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func Run() error {
	flag.Parse()

	*root = path.Clean(*root)
	if !util.IsValidPath(*root) {
		return fmt.Errorf("root path cannot contain '..' or '~'")
	}

	fstat, err := os.Stat(*root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("path %s does not exist", *root)
		}
		return fmt.Errorf("error reading path %s: %v", *root, err)
	}

	httpHandler := handler.ServeDir(*root, *download)
	ftype := "dir"
	if !fstat.IsDir() {
		httpHandler = handler.ServeFile(*root, fstat.Size(), *download)
		ftype = "file"
	}
	http.Handle("/", middleware.Logger(httpHandler))

	addr := fmt.Sprintf(":%d", *port)
	serverUrl := fmt.Sprintf("http://localhost%s", addr)

	fmt.Printf("serving %s [%s] at %s\n", ftype, *root, serverUrl)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer lis.Close()

	go util.OpenBrowser(serverUrl)

	return http.Serve(lis, nil)
}
