package util

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
	"os"
	"path/filepath"
	"time"
)

func GenerateX509KeyPair(streams IOStreams) (cert tls.Certificate, certFile, keyFile string, err error) {
	certFile, keyFile, err = GetCertAndKeyPaths(streams)
	if err == nil {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err == nil {
			fmt.Fprintln(streams.Out(), "Loaded existing tls config")
			return cert, certFile, keyFile, nil
		}
	}
	fmt.Fprintln(streams.Out(), "Generating new tls config...")
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
		return tls.Certificate{}, "", "", err
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, priv.Public(), priv)
	if err != nil {
		return tls.Certificate{}, "", "", err
	}
	certPEMBlock, keyPEMBlock := new(bytes.Buffer), new(bytes.Buffer)
	pem.Encode(certPEMBlock, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	pem.Encode(keyPEMBlock, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	err = errors.Join(
		os.WriteFile(certFile, certPEMBlock.Bytes(), 0644),
		os.WriteFile(keyFile, keyPEMBlock.Bytes(), 0644),
	)
	if err == nil {
		fmt.Fprintln(streams.Err(), "Saved tls config to disk")
	} else {
		certFile, keyFile = "in-memory", "in-memory"
		fmt.Fprintf(streams.Err(), "Failed to save tls config to disk:\n%v\n", err)
	}
	cert, err = tls.X509KeyPair(certPEMBlock.Bytes(), keyPEMBlock.Bytes())
	if err != nil {
		return tls.Certificate{}, "", "", err
	}
	return cert, certFile, keyFile, nil
}

func GetCertAndKeyPaths(streams IOStreams) (certFile, keyFile string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(streams.Err(), "Failed to get user config dir", err)
		return "", "", err
	}
	dir := filepath.Join(home, ".config", "goserve")
	err = os.MkdirAll(dir, 0700)
	if err != nil && !errors.Is(err, os.ErrExist) {
		fmt.Fprintln(streams.Err(), "Failed to create config dir", err)
		return "", "", err
	}
	certFile = filepath.Join(dir, "cert.pem")
	keyFile = filepath.Join(dir, "key.pem")
	return certFile, keyFile, nil
}
