package v2

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// Go default cipher suite minus 3DES
var DefaultCipherSuites = []uint16{
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
	tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_RSA_WITH_AES_256_CBC_SHA,
}

// ToTLSConfig outputs a tls.Config from TLSOptions
//
// ToTLSConfig should only be used for server TLS configuration.
func (t *TLSOptions) ToTLSConfig() (*tls.Config, error) {
	cfg := tls.Config{}

	if t.CertFile != "" || t.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(t.CertFile, t.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("Error loading tls client certificate: %s", err)
		}

		cfg.Certificates = []tls.Certificate{cert}
	}

	cfg.BuildNameToCertificate()
	cfg.CipherSuites = DefaultCipherSuites

	// Tell the server to prefer its cipher suite ordering over what the client
	// prefers.
	cfg.PreferServerCipherSuites = true

	return &cfg, nil
}

// ToTLSClientConfig is like ToTLSConfig but intended for TLS client config.
func (t *TLSOptions) ToTLSClientConfig() (*tls.Config, error) {
	cfg := tls.Config{}
	cfg.InsecureSkipVerify = t.InsecureSkipVerify

	if t.TrustedCAFile != "" {
		caCertPool, err := LoadCACerts(t.TrustedCAFile)
		if err != nil {
			return nil, err
		}
		cfg.RootCAs = caCertPool
	}

	if t.CertFile != "" || t.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(t.CertFile, t.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("Error loading tls client certificate: %s", err)
		}

		cfg.Certificates = []tls.Certificate{cert}
	}

	cfg.CipherSuites = DefaultCipherSuites

	return &cfg, nil
}

// LoadCACerts takes the path to a certificate bundle file in PEM format and try
// to create a x509.CertPool out of it.
func LoadCACerts(path string) (*x509.CertPool, error) {
	caCerts, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error reading CA file: %s", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCerts) {
		return nil, fmt.Errorf("No certificates could be parsed out of %s", err)
	}

	return caCertPool, nil
}
