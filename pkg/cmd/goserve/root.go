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

	root := "."
	if len(flag.Args()) > 0 {
		root = flag.Arg(0)
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

	fileServer := files.NewServer(rootFS, includeDotfiles, version.Get())

	mux := http.NewServeMux()

	slog.Info("registering routes")

	registerRoute(mux, "GET /text", logger.Log(fileServer.ServeText()))
	registerRoute(mux, "GET /text/{path...}", logger.Log(fileServer.ServeText()))
	registerRoute(mux, "GET /html", logger.Log(fileServer.ServeTemplate()))
	registerRoute(mux, "GET /html/{path...}", logger.Log(fileServer.ServeTemplate()))
	registerRoute(mux, "GET /health", logger.Log(fileServer.Health()))
	registerRoute(mux, "GET /version", logger.Log(fileServer.Version()))

	slog.Info("starting server", "root", root, "port", port)

	slog.Info("ready to accept connections")

	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func registerRoute(mux *http.ServeMux, pattern string, handler http.Handler) {
	mux.Handle(pattern, handler)
	slog.Info(pattern)
}
