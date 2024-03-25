package goserve

import (
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/cmgsj/goserve/internal/version"
	"github.com/cmgsj/goserve/pkg/files"
	"github.com/cmgsj/goserve/pkg/files/handlers/html"
	"github.com/cmgsj/goserve/pkg/files/handlers/json"
	"github.com/cmgsj/goserve/pkg/files/handlers/text"
	"github.com/cmgsj/goserve/pkg/middleware/logging"
)

type Flags struct {
	Port       uint
	Exclude    string
	HTML       bool
	JSON       bool
	JSONIndent bool
	Text       bool
	TLSCert    string
	TLSKey     string
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
	flag.StringVar(&f.TLSCert, "tls-cert", f.TLSCert, "tls cert file")
	flag.StringVar(&f.TLSKey, "tls-key", f.TLSKey, "tls key file")
	flag.BoolVar(&f.Version, "version", f.Version, "print version")

	flag.Parse()
}

func Run() error {
	flags := Flags{
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

	root := flag.Arg(0)

	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	info, err := os.Stat(root)
	if err != nil {
		return err
	}

	var fsys fs.FS
	var exclude *regexp.Regexp
	var handlers []files.Handler

	if info.IsDir() {
		fsys = os.DirFS(root)
	} else {
		fsys = os.DirFS(filepath.Dir(root))
		fsys, err = fs.Sub(fsys, filepath.Base(root))
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

	controller := files.NewController(fsys, exclude, handlers...)

	serveTLS := flags.TLSCert != "" && flags.TLSKey != ""

	if flags.Port == 0 {
		if serveTLS {
			flags.Port = 443
		} else {
			flags.Port = 80
		}
	}

	addr := fmt.Sprintf(":%d", flags.Port)

	fmt.Println("Starting file server")
	fmt.Println()

	fmt.Println("Config:")
	fmt.Printf("  Root: %s\n", root)
	fmt.Printf("  Exclude: %s\n", flags.Exclude)
	fmt.Printf("  Content Types: %v\n", controller.ContentTypes())
	if serveTLS {
		fmt.Printf("  TLS Cert: %v\n", flags.TLSCert)
		fmt.Printf("  TLS Key: %v\n", flags.TLSKey)
		fmt.Printf("  Address: https://localhost%s\n", addr)
	} else {
		fmt.Printf("  Address: http://localhost%s\n", addr)
	}
	fmt.Println()

	mux := http.NewServeMux()

	fmt.Println("Routes:")
	if flags.HTML {
		mux.Handle("GET /", http.RedirectHandler("/html", http.StatusMovedPermanently))
	}
	handle := func(pattern string, handler http.Handler) {
		mux.Handle(pattern, logging.LogRequest(handler))
		fmt.Println("  ", pattern)
	}
	handle("GET /{content_type}", controller.FilesHandler())
	handle("GET /{content_type}/{file...}", controller.FilesHandler())
	handle("GET /content_types", controller.ContentTypesHandler())
	handle("GET /health", controller.HealthHandler())
	fmt.Println()

	fmt.Println("Ready to accept connections")
	fmt.Println()

	if serveTLS {
		return http.ListenAndServeTLS(addr, flags.TLSCert, flags.TLSKey, mux)
	}

	return http.ListenAndServe(addr, mux)
}
