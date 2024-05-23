package goserve

import (
	"cmp"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

type Flags struct {
	Host      string
	Port      string
	Exclude   string
	LogLevel  string
	LogFormat string
	LogOutput string
	TLSCert   string
	TLSKey    string
	Version   bool
}

func NewFlags() (Flags, error) {
	f := Flags{
		Host:      envFlag("GOSERVE_HOST"),
		Port:      envFlag("GOSERVE_PORT"),
		Exclude:   envFlag("GOSERVE_EXCLUDE", `^\..+`),
		LogLevel:  envFlag("GOSERVE_LOG_LEVEL", "info"),
		LogFormat: envFlag("GOSERVE_LOG_FORMAT", "text"),
		LogOutput: envFlag("GOSERVE_LOG_OUTPUT", "stderr"),
		TLSCert:   envFlag("GOSERVE_TLS_CERT"),
		TLSKey:    envFlag("GOSERVE_TLS_KEY"),
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

	flag.StringVar(&f.Host, "host", f.Host, "http server host")
	flag.StringVar(&f.Port, "port", f.Port, "http server port")
	flag.StringVar(&f.Exclude, "exclude", f.Exclude, "exclude regex pattern")
	flag.StringVar(&f.LogLevel, "log-level", f.LogLevel, "log level {debug|info|warn|error}")
	flag.StringVar(&f.LogFormat, "log-format", f.LogFormat, "log format {json|text}")
	flag.StringVar(&f.LogOutput, "log-output", f.LogOutput, "log output file {stdout|stderr|FILE}")
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

	var out io.Writer

	switch strings.ToLower(f.LogOutput) {
	case "stdout":
		out = os.Stdout
	case "stderr":
		out = os.Stderr
	default:
		out, err = os.Create(f.LogOutput)
		if err != nil {
			return err
		}
	}

	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: level,
	}

	switch strings.ToLower(f.LogFormat) {
	case "json":
		handler = slog.NewJSONHandler(out, opts)
	case "text":
		handler = slog.NewTextHandler(out, opts)
	default:
		return fmt.Errorf("invalid log format %q", f.LogFormat)
	}

	slog.SetDefault(slog.New(handler))

	return nil
}

func envFlag(key string, defaults ...string) string {
	value, set := os.LookupEnv(key)
	if set {
		return value
	}
	return cmp.Or(defaults...)
}
