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
	"github.com/cmgsj/goserve/pkg/middleware/logging"
)

type Flags struct {
	Port    uint
	Exclude string
	TLSCert string
	TLSKey  string
	Version bool
}

func (f *Flags) Parse() {
	flag.Usage = func() {
		fmt.Printf("HTTP file server\n\n")
		fmt.Printf("Usage:\n  goserve [flags] FILE\n\n")
		fmt.Printf("Flags:\n")
		flag.CommandLine.PrintDefaults()
	}

	flag.UintVar(&f.Port, "port", f.Port, "http port")
	flag.StringVar(&f.Exclude, "exclude", f.Exclude, "exclude regex pattern")
	flag.StringVar(&f.TLSCert, "tls-cert", f.TLSCert, "tls cert file")
	flag.StringVar(&f.TLSKey, "tls-key", f.TLSKey, "tls key file")
	flag.BoolVar(&f.Version, "version", f.Version, "print version")

	flag.Parse()
}

func Run() error {
	flags := Flags{
		Exclude: `^\..+`,
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

	controller := files.NewController(fsys, exclude)

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
	handle := func(pattern string, handler http.Handler) {
		mux.Handle(pattern, logging.LogRequest(handler))
		fmt.Println("  ", pattern)
	}
	handle("GET /", http.RedirectHandler("/html", http.StatusMovedPermanently))
	handle("GET /html/{file...}", controller.FilesHTML())
	handle("GET /json/{file...}", controller.FilesJSON())
	handle("GET /text/{file...}", controller.FilesText())
	handle("GET /health", controller.Health())
	fmt.Println()

	fmt.Println("Ready to accept connections")
	fmt.Println()

	if serveTLS {
		return http.ListenAndServeTLS(addr, flags.TLSCert, flags.TLSKey, mux)
	}

	return http.ListenAndServe(addr, mux)
}
