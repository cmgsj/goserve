package goserve

import (
	"cmp"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Flags struct {
	Host      string
	Port      string
	Exclude   string
	LogLevel  string
	LogSource string
	LogFormat string
	LogOutput string
	TLSCert   string
	TLSKey    string
	Version   bool
}

func NewFlags() (Flags, error) {
	f := Flags{
		Host:      cmp.Or(os.Getenv("GOSERVE_HOST")),
		Port:      cmp.Or(os.Getenv("GOSERVE_PORT")),
		Exclude:   cmp.Or(os.Getenv("GOSERVE_EXCLUDE"), `^\..+`),
		LogLevel:  cmp.Or(os.Getenv("GOSERVE_LOG_LEVEL"), "info"),
		LogSource: cmp.Or(os.Getenv("GOSERVE_LOG_SOURCE"), "false"),
		LogFormat: cmp.Or(os.Getenv("GOSERVE_LOG_FORMAT"), "text"),
		LogOutput: cmp.Or(os.Getenv("GOSERVE_LOG_OUTPUT"), "stderr"),
		TLSCert:   cmp.Or(os.Getenv("GOSERVE_TLS_CERT")),
		TLSKey:    cmp.Or(os.Getenv("GOSERVE_TLS_KEY")),
	}
	f.parse()
	f.complete()
	err := f.loadLogger()
	return f, err
}

func (f *Flags) ServeTLS() bool {
	return f.TLSCert != "" && f.TLSKey != ""
}

func (f *Flags) parse() {
	flag.Usage = func() {
		fmt.Printf("HTTP file server\n\n")
		fmt.Printf("Usage:\n  goserve [flags] FILE\n\n")
		fmt.Printf("Flags:\n")
		flag.CommandLine.PrintDefaults()
	}

	flag.StringVar(&f.Host, "host", f.Host, "http host")
	flag.StringVar(&f.Port, "port", f.Port, "http port")
	flag.StringVar(&f.Exclude, "exclude", f.Exclude, "exclude regex pattern")
	flag.StringVar(&f.LogLevel, "log-level", f.LogLevel, "log level")
	flag.StringVar(&f.LogSource, "log-source", f.LogSource, "log source")
	flag.StringVar(&f.LogFormat, "log-format", f.LogFormat, "log format")
	flag.StringVar(&f.LogOutput, "log-output", f.LogOutput, "log output")
	flag.StringVar(&f.TLSCert, "tls-cert", f.TLSCert, "tls cert file")
	flag.StringVar(&f.TLSKey, "tls-key", f.TLSKey, "tls key file")
	flag.BoolVar(&f.Version, "version", f.Version, "print version")

	flag.Parse()
}

func (f *Flags) complete() {
	if f.Port == "" {
		if f.ServeTLS() {
			f.Port = "443"
		} else {
			f.Port = "80"
		}
	}
}

func (f *Flags) loadLogger() error {
	var level slog.Level
	err := level.UnmarshalText([]byte(f.LogLevel))
	if err != nil {
		return err
	}

	addSource, err := strconv.ParseBool(f.LogSource)
	if err != nil {
		return err
	}

	var out io.Writer
	switch strings.ToLower(f.LogOutput) {
	case "stdout":
		out = os.Stdout
	case "stderr":
		out = os.Stderr
	default:
		out, err = os.Open(f.LogOutput)
		if err != nil {
			return err
		}
	}

	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: addSource,
	}

	switch strings.ToLower(f.LogFormat) {
	case "json":
		handler = slog.NewJSONHandler(out, opts)
	case "text":
		handler = slog.NewTextHandler(out, opts)
	default:
		return fmt.Errorf("invalid LOG_FORMAT env var %q: must be one of [json, text]", f.LogFormat)
	}

	slog.SetDefault(slog.New(handler))

	return nil
}
