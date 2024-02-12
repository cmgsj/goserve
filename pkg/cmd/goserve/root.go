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
	var includeDotfiles bool
	var port uint = 80
	var printVersion bool

	flag.Usage = func() {
		fmt.Println("HTTP file server")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  goserve [flags] [path]")
		fmt.Println()
		fmt.Println("Flags:")
		flag.CommandLine.PrintDefaults()
	}
	flag.BoolVar(&includeDotfiles, "dotfiles", includeDotfiles, "include dotfiles")
	flag.UintVar(&port, "port", port, "port")
	flag.BoolVar(&printVersion, "version", printVersion, "print version")

	flag.Parse()

	if len(flag.Args()) > 1 {
		flag.Usage()
		return fmt.Errorf("invalid number of arguments, expected at most 1, received %d", len(flag.Args()))
	}

	if printVersion {
		fmt.Println(version.Get())
		return nil
	}

	rootPath := "."

	if len(flag.Args()) > 0 {
		rootPath = flag.Arg(0)
	}

	var err error

	rootPath, err = filepath.Abs(rootPath)
	if err != nil {
		return err
	}

	info, err := os.Stat(rootPath)
	if err != nil {
		return err
	}

	var rootFS fs.FS

	if info.IsDir() {
		rootFS = os.DirFS(rootPath)
	} else {
		rootFS = os.DirFS(filepath.Dir(rootPath))
		rootFS, err = fs.Sub(rootFS, filepath.Base(rootPath))
		if err != nil {
			return err
		}
	}

	server := files.NewServer(rootFS, includeDotfiles, version.Get())

	mux := http.NewServeMux()

	slog.Info("registering routes")

	registerRoute(mux, "GET /files", server.ServeTemplate())
	registerRoute(mux, "GET /files/{path...}", server.ServeTemplate())
	registerRoute(mux, "GET /text/files", server.ServeText())
	registerRoute(mux, "GET /text/files/{path...}", server.ServeText())
	registerRoute(mux, "GET /health", server.Health())
	registerRoute(mux, "GET /version", server.Version())

	slog.Info("starting server", "root", rootPath, "dotfiles", includeDotfiles, "port", port)

	slog.Info("ready to accept connections")

	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func registerRoute(mux *http.ServeMux, pattern string, handler http.Handler) {
	mux.Handle(pattern, logger.Log(handler))
	slog.Info(pattern)
}
