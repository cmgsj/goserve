package cmd

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	cmdutil "github.com/cmgsj/goserve/pkg/cmd/util"
	"github.com/cmgsj/goserve/pkg/files"
	"github.com/cmgsj/goserve/pkg/version"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func ExecuteRootCmd() error {
	return NewRootCmd().Execute()
}

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
	cmdutil.IOStreams
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
	o.IOStreams = cmdutil.NewIOStreamsFromCmd(cmd)
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
	fmt.Fprintln(o.Out())
	fmt.Fprintf(o.Out(), "Root: %s\n", o.RootPath)
	fmt.Fprintf(o.Out(), "SkipDotFiles: %t\n", o.SkipDotFiles)
	fmt.Fprintf(o.Out(), "RawEnabled: %t\n", o.RawEnabled)
	fmt.Fprintf(o.Out(), "LogEnabled: %t\n", o.LogEnabled)
	fmt.Fprintf(o.Out(), "CertFile: %s\n", o.CertFile)
	fmt.Fprintf(o.Out(), "KeyFile: %s\n", o.KeyFile)
	if o.RootCAFile != "" {
		fmt.Fprintf(o.Out(), "RootCAFile: %s\n", o.RootCAFile)
	}
	fmt.Fprintf(o.Out(), "Address: https://localhost:%d\n", o.Port)
	fmt.Fprintln(o.Out())
	fmt.Fprintln(o.Out(), "Ready to accept connections")
	fmt.Fprintln(o.Out())
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

var (
	certFile = filepath.FromSlash("/tmp/goserve.cert")
	keyFile  = filepath.FromSlash("/tmp/goserve.key")
)

func (o *RootOptions) LoadTLSConfig() error {
	var (
		cert tls.Certificate
		err  error
	)
	if o.CertFile != "" && o.KeyFile != "" {
		cert, err = tls.LoadX509KeyPair(o.CertFile, o.KeyFile)
	} else {
		cert, err = o.GenerateX509KeyPair()
		o.CertFile, o.KeyFile = certFile, keyFile
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

func (o *RootOptions) GenerateX509KeyPair() (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err == nil {
		fmt.Fprintln(o.Out(), "Loaded existing cert and key files")
		return cert, nil
	}
	fmt.Fprintln(o.Out(), "Generating new cert and key files...")
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(now.Unix()),
		Subject: pkix.Name{
			Organization: []string{"GoServe, INC."},
		},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:             now,
		NotAfter:              now.AddDate(0, 0, 1),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageKeyAgreement,
		BasicConstraintsValid: true,
	}
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return tls.Certificate{}, err
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, priv.Public(), priv)
	if err != nil {
		return tls.Certificate{}, err
	}
	certPEMBlock := new(bytes.Buffer)
	pem.Encode(certPEMBlock, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	keyPEMBlock := new(bytes.Buffer)
	pem.Encode(keyPEMBlock, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
	err = errors.Join(
		os.WriteFile(certFile, certPEMBlock.Bytes(), 0644),
		os.WriteFile(keyFile, keyPEMBlock.Bytes(), 0644),
	)
	if err != nil {
		fmt.Fprintln(o.Err(), "Failed to save cert and key files:", err)
	}
	return tls.X509KeyPair(certPEMBlock.Bytes(), keyPEMBlock.Bytes())
}
