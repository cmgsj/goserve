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
	"github.com/cmgsj/goserve/pkg/files/handlers/html"
	"github.com/cmgsj/goserve/pkg/files/handlers/json"
	"github.com/cmgsj/goserve/pkg/files/handlers/text"
	"github.com/cmgsj/goserve/pkg/middleware/logger"
)

type Flags struct {
	Port     uint
	DotFiles bool
	HTML     bool
	JSON     bool
	Text     bool
	Version  bool
}

func (f *Flags) Parse() {
	flag.Usage = func() {
		fmt.Printf("HTTP file server\n\n")
		fmt.Printf("Usage:\n  goserve [flags] FILE\n\n")
		fmt.Printf("Flags:\n")
		flag.CommandLine.PrintDefaults()
	}

	flag.BoolVar(&f.DotFiles, "dotfiles", f.DotFiles, "include dotfiles")
	flag.UintVar(&f.Port, "port", f.Port, "http port")
	flag.BoolVar(&f.HTML, "html", f.HTML, "enable content-type html")
	flag.BoolVar(&f.JSON, "json", f.JSON, "enable content-type json")
	flag.BoolVar(&f.Text, "text", f.Text, "enable content-type text")
	flag.BoolVar(&f.Version, "version", f.Version, "print version")

	flag.Parse()
}

func Run() error {
	flags := Flags{
		Port: 80,
		HTML: true,
		JSON: true,
		Text: true,
	}

	flags.Parse()

	if flags.Version {
		fmt.Println(internalversion.Get())
		return nil
	}

	if len(flag.Args()) != 1 {
		return fmt.Errorf("accepts %d arg(s), received %d", 1, len(flag.Args()))
	}

	path := flag.Arg(0)

	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	var root fs.FS

	if info.IsDir() {
		root = os.DirFS(path)
	} else {
		root = os.DirFS(filepath.Dir(path))
		root, err = fs.Sub(root, filepath.Base(path))
		if err != nil {
			return err
		}
	}

	var factories []files.HandlerFactory

	if flags.HTML {
		factories = append(factories, html.HandlerFactory())
	}
	if flags.JSON {
		factories = append(factories, json.HandlerFactory())
	}
	if flags.Text {
		factories = append(factories, text.HandlerFactory())
	}

	server := files.NewServer(root, flags.DotFiles, internalversion.Get(), factories...)

	mux := http.NewServeMux()

	register(mux, "GET /{content_type}", server.FilesHandler())
	register(mux, "GET /{content_type}/{file...}", server.FilesHandler())
	register(mux, "GET /content_types", server.ContentTypesHandler())
	register(mux, "GET /health", server.HealthHandler())
	register(mux, "GET /version", server.VersionHandler())

	slog.Info("starting http server", "root", path, "dotfiles", flags.DotFiles, "port", flags.Port, "content_types", server.ContentTypes())

	slog.Info("ready to accept connections")

	return http.ListenAndServe(fmt.Sprintf(":%d", flags.Port), mux)
}

func register(mux *http.ServeMux, pattern string, handler http.Handler) {
	mux.Handle(pattern, logger.Log(handler))
	slog.Info(pattern)
}
