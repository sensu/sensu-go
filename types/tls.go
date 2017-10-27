package types

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// ToTLSConfig outputs a tls.Config from TLSOptions
func (t *TLSOptions) ToTLSConfig() (*tls.Config, error) {
	tlsConfig := tls.Config{}
	tlsConfig.InsecureSkipVerify = t.InsecureSkipVerify

	// Client cert
	if t.CertFile != "" || t.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(t.CertFile, t.KeyFile)
		if err != nil {
			// do something with the error
			return nil, fmt.Errorf("Error loading tls client certificate: %s", err)
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// CA Cert
	if t.TrustedCAFile != "" {
		caCert, err := ioutil.ReadFile(t.TrustedCAFile)
		if err != nil {
			return nil, fmt.Errorf("Error loading tls CA cert: %s", err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	tlsConfig.BuildNameToCertificate()

	return &tlsConfig, nil
}
