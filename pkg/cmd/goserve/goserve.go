package goserve

import (
	"errors"
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
		fmt.Println(version.Version())
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

	var filesystem fs.FS

	if info.IsDir() {
		filesystem = os.DirFS(root)
	} else {
		filesystem = os.DirFS(filepath.Dir(root))
		filesystem, err = fs.Sub(filesystem, filepath.Base(root))
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

	flags.UploadDir, err = filepath.Abs(flags.UploadDir)
	if err != nil {
		return err
	}

	_, err = os.Stat(flags.UploadDir)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		err = os.MkdirAll(flags.UploadDir, 0755)
		if err != nil {
			return err
		}
	}

	scheme := "http"
	if flags.ServeTLS() {
		scheme = "https"
	}

	listener, err := net.Listen("tcp", net.JoinHostPort(flags.Host, flags.Port))
	if err != nil {
		return err
	}

	fmt.Printf("Starting file server at %s://%s\n", scheme, listener.Addr())
	fmt.Println()
	fmt.Println("Config:")
	fmt.Printf("  Root: %q\n", root)
	fmt.Printf("  Host: %q\n", flags.Host)
	fmt.Printf("  Port: %q\n", flags.Port)
	fmt.Printf("  Exclude: %q\n", flags.Exclude)
	if flags.Upload {
		fmt.Printf("  UploadDir: %q\n", flags.UploadDir)
	}
	fmt.Printf("  LogLevel: %q\n", flags.LogLevel)
	fmt.Printf("  LogFormat: %q\n", flags.LogFormat)
	fmt.Printf("  LogOutput: %q\n", flags.LogOutput)
	fmt.Printf("  TLSCert: %q\n", flags.TLSCert)
	fmt.Printf("  TLSKey: %q\n", flags.TLSKey)
	fmt.Println()

	mux := http.NewServeMux()

	handle := func(pattern string, handler http.Handler) {
		mux.Handle(pattern, handler)
		fmt.Printf("  %s\n", pattern)
	}

	controller := files.NewController(files.ControllerOptions{
		FileSystem:      filesystem,
		Exclude:         exclude,
		Upload:          flags.Upload,
		UploadDir:       flags.UploadDir,
		UploadTimestamp: flags.UploadTimestamp,
		Version:         version.Version(),
	})

	fmt.Println("Routes:")

	handle("GET /", http.RedirectHandler("/html", http.StatusMovedPermanently))
	handle("GET /html/{file...}", controller.FilesHTML())
	handle("GET /json/{file...}", controller.FilesJSON())
	handle("GET /text/{file...}", controller.FilesText())
	if flags.Upload {
		handle("POST /html", controller.UploadHTML("/html"))
		handle("POST /json", controller.UploadJSON("/json"))
		handle("POST /text", controller.UploadText("/text"))
	}
	handle("GET /health", controller.Health())

	fmt.Println()
	fmt.Println("Ready to accept connections")
	fmt.Println()

	handler := logging.LogRequests(mux)

	if flags.ServeTLS() {
		return http.ServeTLS(listener, handler, flags.TLSCert, flags.TLSKey)
	}

	return http.Serve(listener, handler)
}
