package goserve

import (
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/cmgsj/goserve/internal/version"
	"github.com/cmgsj/goserve/pkg/files"
	"github.com/cmgsj/goserve/pkg/files/handlers/html"
	"github.com/cmgsj/goserve/pkg/files/handlers/json"
	"github.com/cmgsj/goserve/pkg/files/handlers/text"
	"github.com/cmgsj/goserve/pkg/middleware/logger"
)

type Flags struct {
	Port       uint
	Exclude    string
	HTML       bool
	JSON       bool
	JSONIndent bool
	Text       bool
	Version    bool
}

func (f *Flags) Parse() {
	flag.Usage = func() {
		fmt.Printf("HTTP file server\n\n")
		fmt.Printf("Usage:\n  goserve [flags] FILE\n\n")
		fmt.Printf("Flags:\n")
		flag.CommandLine.PrintDefaults()
	}

	flag.UintVar(&f.Port, "port", f.Port, "http port")
	flag.StringVar(&f.Exclude, "exclude", f.Exclude, "exclude pattern")
	flag.BoolVar(&f.HTML, "html", f.HTML, "enable content-type html")
	flag.BoolVar(&f.JSON, "json", f.JSON, "enable content-type json")
	flag.BoolVar(&f.JSONIndent, "json-indent", f.JSONIndent, "indent content-type json")
	flag.BoolVar(&f.Text, "text", f.Text, "enable content-type text")
	flag.BoolVar(&f.Version, "version", f.Version, "print version")

	flag.Parse()
}

func Run() error {
	flags := Flags{
		Port:       80,
		Exclude:    `^\..+`,
		HTML:       true,
		JSON:       true,
		JSONIndent: true,
		Text:       true,
	}

	flags.Parse()

	if flags.Version {
		fmt.Println(version.String())
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
	var exclude *regexp.Regexp
	var handlers []files.Handler

	if info.IsDir() {
		root = os.DirFS(path)
	} else {
		root = os.DirFS(filepath.Dir(path))
		root, err = fs.Sub(root, filepath.Base(path))
		if err != nil {
			return err
		}
	}

	if flags.Exclude != "" {
		exclude, err = regexp.Compile(flags.Exclude)
		if err != nil {
			return err
		}
	}

	if flags.HTML {
		handlers = append(handlers, html.NewHandler(version.String()))
	}
	if flags.JSON {
		handlers = append(handlers, json.NewHandler(flags.JSONIndent))
	}
	if flags.Text {
		handlers = append(handlers, text.NewHandler())
	}

	controller := files.NewController(root, exclude, handlers...)

	mux := http.NewServeMux()

	slog.Info("starting http server", "port", flags.Port, "root", path, "exclude", flags.Exclude, "content_types", controller.ContentTypes())

	register(mux, "GET /{content_type}", controller.FilesHandler())
	register(mux, "GET /{content_type}/{file...}", controller.FilesHandler())
	register(mux, "GET /content_types", controller.ContentTypesHandler())
	register(mux, "GET /health", controller.HealthHandler())

	slog.Info("ready to accept connections")

	return http.ListenAndServe(fmt.Sprintf(":%d", flags.Port), mux)
}

func register(mux *http.ServeMux, pattern string, handler http.Handler) {
	mux.Handle(pattern, logger.Log(handler))
	slog.Info(pattern)
}
