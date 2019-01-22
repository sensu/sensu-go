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
// NOTE(ccressent): I'm in favour of deprecating this function:
// - it forces the CA cert bundle to be only used by the client side of the TLS
//   connection (see tls.Config.RootCAs vs tls.Config.ClientCAs).
// - its error message assumes the provided certificate is to be used by the
//   client side of the TLS connection.
//
// I suggest the functionality of ToTLSConfig() be split into smaller, more
// composable functions and getting rid of the TLSOption type altogether.
func (t *TLSOptions) ToTLSConfig() (*tls.Config, error) {
	tlsConfig := tls.Config{}
	tlsConfig.InsecureSkipVerify = t.InsecureSkipVerify

	if t.CertFile != "" || t.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(t.CertFile, t.KeyFile)
		if err != nil {
			// do something with the error
			return nil, fmt.Errorf("Error loading tls client certificate: %s", err)
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if t.TrustedCAFile != "" {
		caCertPool, err := LoadCACerts(t.TrustedCAFile)
		if err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = caCertPool
	}

	tlsConfig.BuildNameToCertificate()
	tlsConfig.CipherSuites = DefaultCipherSuites

	return &tlsConfig, nil
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
