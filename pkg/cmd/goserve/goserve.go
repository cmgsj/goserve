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
	"strings"

	"github.com/cmgsj/go-lib/cli"

	"github.com/cmgsj/goserve/pkg/files"
	"github.com/cmgsj/goserve/pkg/middleware/logging"
)

var (
	exclude          = cli.String("exclude", "exclude file pattern")
	host             = cli.String("host", "http host")
	logFormat        = cli.String("log-format", "log format (json|text)", cli.Default("text"))
	logLevel         = cli.String("log-level", "log level (debug|info|warn|error)", cli.Default("info"))
	logOutput        = cli.String("log-output", "log output (stdout|stderr|FILE|none)", cli.Default("stderr"))
	port             = cli.Uint64("port", "http port")
	quiet            = cli.Bool("quiet", "quiet output")
	silent           = cli.Bool("silent", "silent output")
	tlsCert          = cli.String("tls-cert", "tls cert file")
	tlsKey           = cli.String("tls-key", "tls key file")
	uploads          = cli.Bool("uploads", "enable uploads")
	uploadsDir       = cli.String("uploads-dir", "uploads directory")
	uploadsTimestamp = cli.Bool("uploads-timestamp", "add upload timestamp")
	version          = cli.Bool("version", "print version")
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

	rootInfo, err := os.Stat(root)
	if err != nil {
		return err
	}

	var fileSystem fs.FS

	if rootInfo.IsDir() {
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

	if uploads.Value() {
		if uploadsDir.Value() == "" {
			uploadsDir.SetValue(os.TempDir())
		}

		uploadsDirAbs, err := filepath.Abs(uploadsDir.Value())
		if err != nil {
			return err
		}

		uploadsDir.SetValue(uploadsDirAbs)

		_, err = os.Stat(uploadsDir.Value())
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return err
			}

			err = os.MkdirAll(uploadsDir.Value(), 0755)
			if err != nil {
				return err
			}
		}
	}

	serveTLS := tlsCert.Value() != "" && tlsKey.Value() != ""

	if host.Value() == "" {
		host.SetValue("0.0.0.0")
	}

	if port.Value() == 0 {
		if serveTLS {
			port.SetValue(443)
		} else {
			port.SetValue(80)
		}
	}

	address := net.JoinHostPort(host.Value(), strconv.FormatUint(port.Value(), 10))

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
		Host:   strings.ReplaceAll(address, "0.0.0.0", "localhost"),
	}

	controller := files.NewController(fileSystem, files.ControllerConfig{
		FilesURL:         "/files",
		ExcludePattern:   excludePattern,
		Uploads:          uploads.Value(),
		UploadsDir:       uploadsDir.Value(),
		UploadsTimestamp: uploadsTimestamp.Value(),
		Version:          Version(),
	})

	mux := http.NewServeMux()

	printfln("")
	printfln(`   __________  ________  ______   _____ `)
	printfln(`  / __  / __ \/ ___/ _ \/ ___/ | / / _ \`)
	printfln(` / /_/ / /_/ (__  )  __/ /   | |/ /  __/`)
	printfln(` \__, /\____/____/\___/_/    |___/\___/ `)
	printfln(`/____/                                  `)
	printfln("")
	printfln("")
	printfln("Starting HTTP file server")
	printfln("")
	printfln("Config:")

	err = printConfigs([]config{
		{
			key:   "Root",
			value: root,
		},
		{
			key:   "Host",
			value: host.Value(),
		},
		{
			key:   "Port",
			value: port.Value(),
		},
		{
			key:      "Exclude Pattern",
			value:    excludePattern,
			disabled: excludePattern == nil,
		},
		{
			key:      "Uploads Dir",
			value:    uploadsDir.Value(),
			disabled: !uploads.Value(),
		},
		{
			key:   "Log Level",
			value: logLevel.Value(),
		},
		{
			key:   "Log Format",
			value: logFormat.Value(),
		},
		{
			key:   "Log Output",
			value: logOutput.Value(),
		},
		{
			key:      "TLS Cert",
			value:    tlsCert.Value(),
			disabled: !serveTLS,
		},
		{
			key:      "TLS Key",
			value:    tlsKey.Value(),
			disabled: !serveTLS,
		},
	})
	if err != nil {
		return err
	}

	printfln("")
	printfln("Routes:")

	err = registerRoutes(mux, []route{
		{
			pattern:     "/",
			description: "Redirect /files",
			handler:     http.RedirectHandler("/files", http.StatusMovedPermanently),
		},
		{
			pattern:     "GET /files",
			description: "List Files",
			handler:     controller.ListFiles(),
		},
		{
			pattern:     "GET /files/{file...}",
			description: "List Files",
			handler:     controller.ListFiles(),
			disabled:    !rootInfo.IsDir(),
		},
		{
			pattern:     "POST /files",
			description: "Upload File",
			handler:     controller.UploadFile(),
			disabled:    !uploads.Value(),
		},
		{
			pattern:     "GET /health",
			description: "Health Check",
			handler:     health(),
		},
	})
	if err != nil {
		return err
	}

	printfln("")
	printfln("Serving files at %s", url)
	printfln("")
	printfln("Ready to accept connections")
	printfln("")

	return http.Serve(listener, logging.LogRequests(mux))
}
