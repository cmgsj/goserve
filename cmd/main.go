package main

import (
	"errors"
	"flag"
	"fmt"
	"goserve/pkg/handler"
	"goserve/pkg/middleware"
	"goserve/pkg/util"
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
	err := util.ValidatePath(*root)
	if err != nil {
		return err
	}
	fstat, err := os.Stat(*root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("path %s does not exist", *root)
		}
		return fmt.Errorf("error reading path %s: %v", *root, err)
	}
	defaultHandler := handler.ServeDir(*root, *download)
	ftype := "dir"
	if !fstat.IsDir() {
		defaultHandler = handler.ServeFile(*root, fstat.Size(), *download)
		ftype = "file"
	}
	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("serving %s [%s] at http://localhost%s\n", ftype, *root, addr)
	return http.ListenAndServe(addr, middleware.Logger(defaultHandler))
}
