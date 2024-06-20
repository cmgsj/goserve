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
	"strconv"

	"github.com/cmgsj/go-lib/cli"

	"github.com/cmgsj/goserve/pkg/files"
	"github.com/cmgsj/goserve/pkg/middleware/logging"
)

var (
	exclude          = cli.StringFlag("exclude", "exclude file pattern", false)
	host             = cli.StringFlag("host", "http server host", false)
	logFormat        = cli.StringFlag("log-format", "log format { json | text }", false, "text")
	logLevel         = cli.StringFlag("log-level", "log level { debug | info | warn | error }", false, "info")
	logOutput        = cli.StringFlag("log-output", "log output { stdout | stderr | FILE }", false, "stderr")
	port             = cli.Uint64Flag("port", "http server port", false)
	silent           = cli.BoolFlag("silent", "silent mode", false)
	tlsCert          = cli.StringFlag("tls-cert", "tls cert file", false)
	tlsKey           = cli.StringFlag("tls-key", "tls key file", false)
	uploads          = cli.BoolFlag("uploads", "enable uploads", false)
	uploadsDir       = cli.StringFlag("uploads-dir", "uploads directory", false)
	uploadsTimestamp = cli.BoolFlag("uploads-timestamp", "add upload timestamp", false)
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

	address := net.JoinHostPort(host, strconv.FormatUint(port, 10))

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
		FilesURL:         "/files",
		ExcludePattern:   excludePattern,
		Uploads:          uploads.Value(),
		UploadsDir:       uploadsDirPath,
		UploadsTimestamp: uploadsTimestamp.Value(),
		Version:          Version(),
	})

	mux := http.NewServeMux()

	handler := logging.LogRequests(mux)

	println()
	println(`   __________  ________  ______   _____ `)
	println(`  / __  / __ \/ ___/ _ \/ ___/ | / / _ \`)
	println(` / /_/ / /_/ (__  )  __/ /   | |/ /  __/`)
	println(` \__, /\____/____/\___/_/    |___/\___/ `)
	println(`/____/                                  `)
	println()
	println()
	println("Starting HTTP file server")
	println()
	println("Config:")

	err = printConfigs([]config{
		{key: "Root", value: root},
		{key: "Host", value: host},
		{key: "Port", value: port},
		{key: "Exclude Pattern", value: excludePattern, disabled: exclude.Value() == ""},
		{key: "Uploads Dir", value: uploadsDirPath, disabled: !uploads.Value()},
		{key: "Log Level", value: logLevel.Value()},
		{key: "Log Format", value: logFormat.Value()},
		{key: "Log Output", value: logOutput.Value()},
		{key: "TLS Cert", value: tlsCert.Value(), disabled: !serveTLS},
		{key: "TLS Key", value: tlsKey.Value(), disabled: !serveTLS},
	})
	if err != nil {
		return err
	}

	println()
	println("Routes:")

	err = registerRoutes(mux, []route{
		{
			patterns:    []string{"GET /"},
			description: "Redirect /files",
			handler:     http.RedirectHandler("/files", http.StatusMovedPermanently),
		},
		{
			patterns:    []string{"GET /files", "GET /files/{file...}"},
			description: "List Files",
			handler:     controller.ListFiles(),
		},
		{
			patterns:    []string{"POST /files"},
			description: "Upload File",
			handler:     controller.UploadFile(),
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

	println()
	printfln("Listening at %s", url)
	println()
	println("Ready to accept connections")
	println()

	return http.Serve(listener, handler)
}
