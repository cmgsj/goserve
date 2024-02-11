package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cmgsj/goserve/pkg/files"
	utiltls "github.com/cmgsj/goserve/pkg/util/tls"
	"github.com/cmgsj/goserve/pkg/version"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	o := &RootOptions{
		Port:         1234,
		RawEnabled:   true,
		LogEnabled:   true,
		SkipDotFiles: true,
		RootPath:     ".",
	}
	cmd := &cobra.Command{
		Use:     "goserve [filepath]",
		Short:   "Static file server",
		Long:    "Http static file server with web UI.",
		Version: version.Version,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Complete(cmd, args)
			if err != nil {
				return err
			}
			return o.Run(cmd, args)
		},
	}
	cmd.PersistentFlags().IntVar(&o.Port, "port", o.Port, "port to listen on")
	cmd.PersistentFlags().BoolVar(&o.RawEnabled, "raw", o.RawEnabled, "whether to serve raw files or to download")
	cmd.PersistentFlags().BoolVar(&o.LogEnabled, "log", o.LogEnabled, "whether to log request info to stdout or not")
	cmd.PersistentFlags().BoolVar(&o.SkipDotFiles, "skip-dot-files", o.SkipDotFiles, "whether to skip files whose name starts with '.' or not")
	cmd.PersistentFlags().StringVar(&o.CertFile, "cert", o.CertFile, "path to TLS cert file")
	cmd.PersistentFlags().StringVar(&o.KeyFile, "key", o.KeyFile, "path to TLS key file")
	cmd.PersistentFlags().StringVar(&o.RootCAFile, "root-ca", o.RootCAFile, "path to root CA file")
	return cmd
}

type RootOptions struct {
	Port         int
	RawEnabled   bool
	LogEnabled   bool
	SkipDotFiles bool
	RootPath     string
	CertFile     string
	KeyFile      string
	RootCAFile   string
	TLSConfig    *tls.Config
}

func (o *RootOptions) Complete(cmd *cobra.Command, args []string) error {
	fmt.Println()
	if len(args) > 0 {
		o.RootPath = args[0]
	}
	err := o.LoadRootPath()
	if err != nil {
		return err
	}
	return o.LoadTLSConfig()
}

func (o *RootOptions) Run(cmd *cobra.Command, args []string) error {
	fsys := afero.NewReadOnlyFs(afero.NewBasePathFs(afero.NewOsFs(), o.RootPath))
	handler := files.NewServer(fsys, cmd.OutOrStdout(), cmd.ErrOrStderr(), o.SkipDotFiles, o.RawEnabled, o.LogEnabled)
	h := &http.Server{
		Addr:      fmt.Sprintf(":%d", o.Port),
		Handler:   handler,
		TLSConfig: o.TLSConfig,
	}
	o.PrintInfo()
	return h.ListenAndServeTLS("", "")
}

func (o *RootOptions) PrintInfo() {
	fmt.Println()
	fmt.Printf("Root: %s\n", o.RootPath)
	fmt.Printf("SkipDotFiles: %t\n", o.SkipDotFiles)
	fmt.Printf("RawEnabled: %t\n", o.RawEnabled)
	fmt.Printf("LogEnabled: %t\n", o.LogEnabled)
	fmt.Printf("CertFile: %s\n", o.CertFile)
	fmt.Printf("KeyFile: %s\n", o.KeyFile)
	if o.RootCAFile != "" {
		fmt.Printf("RootCAFile: %s\n", o.RootCAFile)
	}
	fmt.Printf("Address: https://localhost:%d\n", o.Port)
	fmt.Println()
	fmt.Println("Ready to accept connections")
	fmt.Println()
}

func (o *RootOptions) LoadRootPath() error {
	var err error
	o.RootPath, err = filepath.Abs(o.RootPath)
	if err != nil {
		return err
	}
	_, err = os.Stat(o.RootPath)
	return err
}

func (o *RootOptions) LoadTLSConfig() error {
	var (
		cert tls.Certificate
		err  error
	)
	if o.CertFile != "" && o.KeyFile != "" {
		cert, err = tls.LoadX509KeyPair(o.CertFile, o.KeyFile)
	} else {
		cert, o.CertFile, o.KeyFile, err = utiltls.GenerateX509KeyPair()
	}
	if err != nil {
		return err
	}
	o.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	if o.RootCAFile != "" {
		rootCA, err := os.ReadFile(o.RootCAFile)
		if err != nil {
			return err
		}
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(rootCA) {
			return fmt.Errorf("failed to add RootCA: %s", o.RootCAFile)
		}
		o.TLSConfig.RootCAs = certPool
		o.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return nil
}
