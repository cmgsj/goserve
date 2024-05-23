package goserve

import (
	"flag"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/cmgsj/goserve/internal/version"
	"github.com/cmgsj/goserve/pkg/files"
	"github.com/cmgsj/goserve/pkg/middleware/logging"
)

func Run() error {
	flags, err := NewFlags()
	if err != nil {
		return err
	}

	if flags.Version {
		fmt.Println(version.String())
		return nil
	}

	if len(flag.Args()) != 1 {
		return fmt.Errorf("accepts %d arg(s), received %d", 1, len(flag.Args()))
	}

	root := flag.Arg(0)

	root, err = filepath.Abs(root)
	if err != nil {
		return err
	}

	info, err := os.Stat(root)
	if err != nil {
		return err
	}

	var fsys fs.FS

	if info.IsDir() {
		fsys = os.DirFS(root)
	} else {
		fsys = os.DirFS(filepath.Dir(root))
		fsys, err = fs.Sub(fsys, filepath.Base(root))
		if err != nil {
			return err
		}
	}

	var exclude *regexp.Regexp

	if flags.Exclude != "" {
		exclude, err = regexp.Compile(flags.Exclude)
		if err != nil {
			return err
		}
	}

	controller := files.NewController(fsys, exclude)

	listener, err := net.Listen("tcp", net.JoinHostPort(flags.Host, flags.Port))
	if err != nil {
		return err
	}

	fmt.Println("Starting file server")
	fmt.Println()

	fmt.Println("Config:")
	fmt.Printf("  Root: %s\n", root)
	fmt.Printf("  Exclude: %s\n", flags.Exclude)
	if flags.ServeTLS() {
		fmt.Printf("  TLS Cert: %v\n", flags.TLSCert)
		fmt.Printf("  TLS Key: %v\n", flags.TLSKey)
		fmt.Printf("  Address: https://%s\n", listener.Addr())
	} else {
		fmt.Printf("  Address: http://%s\n", listener.Addr())
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

	if flags.ServeTLS() {
		return http.ServeTLS(listener, mux, flags.TLSCert, flags.TLSKey)
	}

	return http.Serve(listener, mux)
}
