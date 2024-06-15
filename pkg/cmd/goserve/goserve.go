package goserve

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/cmgsj/goserve/pkg/cli"
	"github.com/cmgsj/goserve/pkg/files"
	"github.com/cmgsj/goserve/pkg/middleware/logging"
)

var (
	host            = cli.StringFlag("host", "http server host", false)
	port            = cli.Uint64Flag("port", "http server port", false)
	exclude         = cli.StringFlag("exclude", "exclude file pattern", false)
	upload          = cli.BoolFlag("upload", "enable uploads", false)
	uploadDir       = cli.StringFlag("upload-dir", "uploads directory", false)
	uploadTimestamp = cli.BoolFlag("upload-timestamp", "add upload timestamp", false)
	logLevel        = cli.StringFlag("log-level", "log level { debug | info | warn | error }", false, "info")
	logFormat       = cli.StringFlag("log-format", "log format { json | text }", false, "text")
	logOutput       = cli.StringFlag("log-output", "log output { stdout | stderr | FILE }", false, "stderr")
	tlsCert         = cli.StringFlag("tls-cert", "tls cert file", false)
	tlsKey          = cli.StringFlag("tls-key", "tls key file", false)
	version         = cli.BoolFlag("version", "print version", false)
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

	var excludeRegexp *regexp.Regexp

	if exclude.Value() != "" {
		excludeRegexp, err = regexp.Compile(exclude.Value())
		if err != nil {
			return err
		}
	}

	var uploadDirPath string

	if upload.Value() {
		uploadDirPath = uploadDir.Value()

		if uploadDirPath == "" {
			uploadDirPath = os.TempDir()
		}

		uploadDirPath, err = filepath.Abs(uploadDirPath)
		if err != nil {
			return err
		}

		_, err = os.Stat(uploadDirPath)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return err
			}

			err = os.MkdirAll(uploadDirPath, 0755)
			if err != nil {
				return err
			}
		}
	}

	serveTLS := tlsCert.Value() != "" && tlsKey.Value() != ""

	scheme := "http"

	if serveTLS {
		scheme = "https"
	}

	port := port.Value()

	if port == 0 {
		if serveTLS {
			port = 443
		} else {
			port = 80
		}
	}

	listener, err := net.Listen("tcp", net.JoinHostPort(host.Value(), strconv.FormatUint(port, 10)))
	if err != nil {
		return err
	}

	fmt.Printf("Starting file server at %s://%s\n", scheme, listener.Addr())
	fmt.Println()
	fmt.Println("Config:")
	fmt.Printf("  Root: %q\n", root)
	fmt.Printf("  Host: %q\n", host.Value())
	fmt.Printf("  Port: %d\n", port)
	fmt.Printf("  Exclude: %q\n", excludeRegexp)
	if upload.Value() {
		fmt.Printf("  UploadDir: %q\n", uploadDirPath)
	}
	fmt.Printf("  LogLevel: %q\n", logLevel.Value())
	fmt.Printf("  LogFormat: %q\n", logFormat.Value())
	fmt.Printf("  LogOutput: %q\n", logOutput.Value())
	if serveTLS {
		fmt.Printf("  TLSCert: %q\n", tlsCert.Value())
		fmt.Printf("  TLSKey: %q\n", tlsKey.Value())
	}
	fmt.Println()

	mux := http.NewServeMux()

	handle := func(pattern string, handler http.Handler) {
		mux.Handle(pattern, handler)
		fmt.Printf("  %s\n", pattern)
	}

	controller := files.NewController(files.ControllerOptions{
		FileSystem:      fileSystem,
		ExcludeRegexp:   excludeRegexp,
		Upload:          upload.Value(),
		UploadDir:       uploadDirPath,
		UploadTimestamp: uploadTimestamp.Value(),
		Version:         Version(),
	})

	fmt.Println("Routes:")

	handle("GET /", http.RedirectHandler("/html", http.StatusMovedPermanently))
	handle("GET /html/{file...}", controller.FilesHTML())
	handle("GET /json/{file...}", controller.FilesJSON())
	handle("GET /text/{file...}", controller.FilesText())
	if upload.Value() {
		handle("POST /html", controller.UploadHTML("/html"))
		handle("POST /json", controller.UploadJSON("/json"))
		handle("POST /text", controller.UploadText("/text"))
	}
	handle("GET /health", controller.Health())

	fmt.Println()
	fmt.Println("Ready to accept connections")
	fmt.Println()

	handler := logging.LogRequests(mux)

	if serveTLS {
		return http.ServeTLS(listener, handler, tlsCert.Value(), tlsKey.Value())
	}

	return http.Serve(listener, handler)
}
