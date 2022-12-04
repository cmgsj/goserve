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
	port      = flag.Int("port", 1234, "port to listen on")
	plaintext = flag.Bool("plain-text", true, "serve files as plain-text(true), or to download(false) (default: true)")
)

func main() {
	if err := Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func Run() error {
	flag.Parse()
	root := "."
	if len(flag.Args()) > 1 {
		return fmt.Errorf("usage: goserve [-port=number] [-plain-text=bool] [path]")
	} else if len(flag.Args()) == 1 {
		root = path.Clean(flag.Arg(0))
	}
	err := util.ValidatePath(root)
	if err != nil {
		return err
	}
	fstat, err := os.Stat(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("path %s does not exist", root)
		}
		return fmt.Errorf("reading path %s failed: %v", root, err)
	}
	defaultHandler := handler.ServeDir(root, *plaintext)
	ftype := "dir"
	if !fstat.IsDir() {
		defaultHandler = handler.ServeFile(root, fstat.Size(), *plaintext)
		ftype = "file"
	}
	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("serving %s [%s] at http://localhost%s\n", ftype, root, addr)
	return http.ListenAndServe(addr, middleware.Logger(defaultHandler))
}
