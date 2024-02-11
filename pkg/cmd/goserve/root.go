package goserve

import (
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cmgsj/goserve/internal/version"
	"github.com/cmgsj/goserve/pkg/files"
	"github.com/cmgsj/goserve/pkg/middleware/logger"
)

func Run() error {
	includeDotfiles := flag.Bool("dotfiles", false, "include dotfiles")
	port := flag.Uint("port", 80, "port")
	printVersion := flag.Bool("v", false, "print version")

	flag.Parse()

	if len(os.Args) > 2 {
		return fmt.Errorf("%s [flags] [path]\n", os.Args[0])
	}

	if *printVersion {
		fmt.Println(version.Get())
		return nil
	}

	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	abs, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	root = abs

	info, err := os.Stat(root)
	if err != nil {
		return err
	}

	var rootFS fs.FS

	if info.IsDir() {
		rootFS = os.DirFS(root)
	} else {
		rootFS = os.DirFS(filepath.Dir(root))
		rootFS, err = fs.Sub(rootFS, filepath.Base(root))
		if err != nil {
			return err
		}
	}

	fileServer := files.NewServer(rootFS, *includeDotfiles, version.Get())

	mux := http.NewServeMux()

	slog.Info("registering routes")

	registerRoute(mux, "GET /text", logger.Log(fileServer.ServeText()))
	registerRoute(mux, "GET /text/{path...}", logger.Log(fileServer.ServeText()))
	registerRoute(mux, "GET /html", logger.Log(fileServer.ServeTemplate()))
	registerRoute(mux, "GET /html/{path...}", logger.Log(fileServer.ServeTemplate()))
	registerRoute(mux, "GET /health", logger.Log(fileServer.Health()))
	registerRoute(mux, "GET /version", logger.Log(fileServer.Version()))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: mux,
	}

	slog.Info("starting server", "root", root, "port", *port)

	slog.Info("ready to accept connections")

	return server.ListenAndServe()
}

func registerRoute(mux *http.ServeMux, pattern string, handler http.Handler) {
	mux.Handle(pattern, handler)
	slog.Info(pattern)
}
