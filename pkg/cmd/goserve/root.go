package goserve

import (
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	internalversion "github.com/cmgsj/goserve/internal/version"
	"github.com/cmgsj/goserve/pkg/files"
	"github.com/cmgsj/goserve/pkg/middleware/logger"
)

func Run() error {
	var dotfiles bool
	var port uint = 80
	var version bool

	flag.Usage = func() {
		fmt.Printf("HTTP file server\n\n")
		fmt.Printf("Usage:\n  goserve [flags] FILE\n\n")
		fmt.Printf("Flags:\n")
		flag.CommandLine.PrintDefaults()
	}
	flag.BoolVar(&dotfiles, "dotfiles", dotfiles, "include dotfiles")
	flag.UintVar(&port, "port", port, "port")
	flag.BoolVar(&version, "version", version, "print version")

	flag.Parse()

	if version {
		fmt.Println(internalversion.Get())
		return nil
	}

	if len(flag.Args()) != 1 {
		return fmt.Errorf("accepts %d arg(s), received %d", 1, len(flag.Args()))
	}

	rootPath := flag.Arg(0)

	rootPath, err := filepath.Abs(rootPath)
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

	server := files.NewServer(rootFS, dotfiles, internalversion.Get())

	mux := http.NewServeMux()

	slog.Info("registering routes")

	registerRoute(mux, "GET /files", server.ServePage())
	registerRoute(mux, "GET /files/{file...}", server.ServePage())
	registerRoute(mux, "GET /text/files", server.ServeText())
	registerRoute(mux, "GET /text/files/{file...}", server.ServeText())
	registerRoute(mux, "GET /health", server.Health())
	registerRoute(mux, "GET /version", server.Version())

	slog.Info("starting http server", "root", rootPath, "dotfiles", dotfiles, "port", port)

	slog.Info("ready to accept connections")

	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func registerRoute(mux *http.ServeMux, pattern string, handler http.Handler) {
	mux.Handle(pattern, logger.Log(handler))
	slog.Info(pattern)
}
