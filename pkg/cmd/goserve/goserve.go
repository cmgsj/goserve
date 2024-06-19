package goserve

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	"github.com/cmgsj/go-lib/cli"

	"github.com/cmgsj/goserve/pkg/files"
	"github.com/cmgsj/goserve/pkg/middleware/logging"
)

var (
	host             = cli.StringFlag("host", "http server host", false)
	port             = cli.Uint64Flag("port", "http server port", false)
	exclude          = cli.StringFlag("exclude", "exclude file pattern", false)
	uploads          = cli.BoolFlag("uploads", "enable uploads", false)
	uploadsDir       = cli.StringFlag("uploads-dir", "uploads directory", false)
	uploadsTimestamp = cli.BoolFlag("uploads-timestamp", "add upload timestamp", false)
	logLevel         = cli.StringFlag("log-level", "log level { debug | info | warn | error }", false, "info")
	logFormat        = cli.StringFlag("log-format", "log format { json | text }", false, "text")
	logOutput        = cli.StringFlag("log-output", "log output { stdout | stderr | FILE }", false, "stderr")
	tlsCert          = cli.StringFlag("tls-cert", "tls cert file", false)
	tlsKey           = cli.StringFlag("tls-key", "tls key file", false)
	version          = cli.BoolFlag("version", "print version", false)
)

func Run() error {
	cli.SetEnvPrefix("goserve")

	cli.SetUsage(func(flagSet *cli.FlagSet) {
		fmt.Printf("HTTP file server\n\n")
		fmt.Printf("Usage:\n  goserve [flags] PATH\n\n")
		fmt.Printf("Flags:\n")
		flagSet.PrintDefaults()
	})

	err := cli.Parse()
	if err != nil {
		return err
	}

	if version.Value() {
		fmt.Println(Version())
		return nil
	}

	err = initLogger()
	if err != nil {
		return err
	}

	if len(cli.Args()) != 1 {
		return fmt.Errorf("accepts %d arg(s), received %d", 1, len(cli.Args()))
	}

	root := cli.Arg(0)

	root, err = filepath.Abs(root)
	if err != nil {
		return err
	}

	info, err := os.Stat(root)
	if err != nil {
		return err
	}

	var fileSystem fs.FS

	if info.IsDir() {
		fileSystem = os.DirFS(root)
	} else {
		fileSystem = os.DirFS(filepath.Dir(root))
		fileSystem, err = fs.Sub(fileSystem, filepath.Base(root))
		if err != nil {
			return err
		}
	}

	var excludePattern *regexp.Regexp

	if exclude.Value() != "" {
		excludePattern, err = regexp.Compile(exclude.Value())
		if err != nil {
			return err
		}
	}

	var uploadsDirPath string

	if uploads.Value() {
		uploadsDirPath = uploadsDir.Value()

		if uploadsDirPath == "" {
			uploadsDirPath = os.TempDir()
		}

		uploadsDirPath, err = filepath.Abs(uploadsDirPath)
		if err != nil {
			return err
		}

		_, err = os.Stat(uploadsDirPath)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return err
			}

			err = os.MkdirAll(uploadsDirPath, 0755)
			if err != nil {
				return err
			}
		}
	}

	serveTLS := tlsCert.Value() != "" && tlsKey.Value() != ""

	host := host.Value()

	if host == "" {
		host = "0.0.0.0"
	}

	port := port.Value()

	if port == 0 {
		if serveTLS {
			port = 443
		} else {
			port = 80
		}
	}

	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	scheme := "http"

	if serveTLS {
		scheme = "https"

		certificate, err := tls.LoadX509KeyPair(tlsCert.Value(), tlsKey.Value())
		if err != nil {
			return err
		}

		listener = tls.NewListener(listener, &tls.Config{
			Certificates: []tls.Certificate{certificate},
		})
	}

	url := &url.URL{
		Scheme: scheme,
		Host:   address,
	}

	controller := files.NewController(fileSystem, files.ControllerConfig{
		ExcludePattern:   excludePattern,
		Uploads:          uploads.Value(),
		UploadsDir:       uploadsDirPath,
		UploadsTimestamp: uploadsTimestamp.Value(),
		Version:          Version(),
	})

	mux := http.NewServeMux()

	handler := logging.LogRequests(mux)

	fmt.Println("Starting HTTP file server")
	fmt.Println()
	fmt.Println("Config:")
	fmt.Printf("  Root: %q\n", root)
	fmt.Printf("  Host: %q\n", host)
	fmt.Printf("  Port: %d\n", port)
	if exclude.Value() != "" {
		fmt.Printf("  Exclude Pattern: %q\n", excludePattern)
	}
	if uploads.Value() {
		fmt.Printf("  Uploads Dir: %q\n", uploadsDirPath)
	}
	fmt.Printf("  Log Level: %q\n", logLevel.Value())
	fmt.Printf("  Log Format: %q\n", logFormat.Value())
	fmt.Printf("  Log Output: %q\n", logOutput.Value())
	if serveTLS {
		fmt.Printf("  TLS Cert: %q\n", tlsCert.Value())
		fmt.Printf("  TLS Key: %q\n", tlsKey.Value())
	}
	fmt.Println()

	fmt.Println("Routes:")

	err = registerRoutes(mux, []route{
		{
			patterns:    []string{"GET /"},
			description: "Redirect (/html)",
			handler:     http.RedirectHandler("/html", http.StatusMovedPermanently),
		},
		{
			patterns:    []string{"GET /html", "GET /html/{file...}"},
			description: "List Files HTML",
			handler:     controller.ListFilesHTML(),
		},
		{
			patterns:    []string{"GET /json", "GET /json/{file...}"},
			description: "List Files JSON",
			handler:     controller.ListFilesJSON(),
		},
		{
			patterns:    []string{"GET /text", "GET /text/{file...}"},
			description: "List Files Text",
			handler:     controller.ListFilesText(),
		},
		{
			patterns:    []string{"POST /html"},
			description: "Upload File HTML",
			handler:     controller.UploadFileHTML("/html"),
			disabled:    !uploads.Value(),
		},
		{
			patterns:    []string{"POST /json"},
			description: "Upload File JSON",
			handler:     controller.UploadFileJSON("/json"),
			disabled:    !uploads.Value(),
		},
		{
			patterns:    []string{"POST /text"},
			description: "Upload File Text",
			handler:     controller.UploadFileText("/text"),
			disabled:    !uploads.Value(),
		},
		{
			patterns:    []string{"GET /health"},
			description: "Health Check",
			handler:     health(),
		},
	})
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("Listening at %s\n", url)
	fmt.Println()
	fmt.Println("Ready to accept connections")
	fmt.Println()

	return http.Serve(listener, handler)
}
