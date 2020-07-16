package v2

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

var (
	// PCI compliance as of Jun 30, 2018: anything under TLS 1.1 must be disabled
	// we bump this up to TLS 1.2 so we can support the best possible ciphers
	tlsMinVersion = uint16(tls.VersionTLS12)
	// DefaultCipherSuites overrides the default cipher suites in order to disable
	// CBC suites (Lucky13 attack) this means TLS 1.1 can't work (no GCM)
	// additionally, we should only use perfect forward secrecy ciphers
	DefaultCipherSuites = []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		// these ciphers require go 1.8+
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
	}
	// optimal EC curve preference
	// curve reference: http://safecurves.cr.yp.to/
	tlsCurvePreferences = []tls.CurveID{
		// this curve is a non-NIST curve with no NSA influence. Prefer this over all others!
		tls.X25519,
		// These curves are provided by NIST; optimal order
		tls.CurveP384,
		tls.CurveP256,
		tls.CurveP521,
	}
)

// ToServerTLSConfig should only be used for server TLS configuration. outputs a tls.Config from TLSOptions
func (t *TLSOptions) ToServerTLSConfig() (*tls.Config, error) {
	cfg := tls.Config{}

	if t.GetTrustedCAFile() != "" {
		caCertPool, err := LoadCACerts(t.TrustedCAFile)
		if err != nil {
			return nil, err
		}
		// client trust store should ONLY consist of specified CAs
		cfg.ClientCAs = caCertPool
	}

	if t.GetCertFile() != "" && t.GetKeyFile() != "" {
		cert, err := tls.LoadX509KeyPair(t.GetCertFile(), t.GetKeyFile())
		if err != nil {
			return nil, fmt.Errorf("Error loading tls server certificate: %s", err)
		}

		cfg.Certificates = []tls.Certificate{cert}
	}

	// useful when we present multiple certificates
	//nolint:staticcheck // ignore SA1019 for old code
	cfg.BuildNameToCertificate()

	// apply hardened TLS settings
	cfg.MinVersion = tlsMinVersion
	cfg.CurvePreferences = tlsCurvePreferences
	cfg.CipherSuites = DefaultCipherSuites
	// Tell the server to prefer it's own cipher suite ordering over the client's preferred ordering
	cfg.PreferServerCipherSuites = true

	// Enable TLS client authentication if configured
	if t.GetClientAuthType() {
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return &cfg, nil
}

// ToClientTLSConfig is like ToServerTLSConfig but intended for TLS client config.
func (t *TLSOptions) ToClientTLSConfig() (*tls.Config, error) {
	cfg := tls.Config{}
	cfg.InsecureSkipVerify = t.GetInsecureSkipVerify()

	if t.GetTrustedCAFile() != "" {
		caCertPool, err := LoadCACerts(t.TrustedCAFile)
		if err != nil {
			return nil, err
		}
		// client trust store should ONLY consist of specified CAs
		cfg.RootCAs = caCertPool
	}

	if t.GetCertFile() != "" && t.GetKeyFile() != "" {
		cert, err := tls.LoadX509KeyPair(t.GetCertFile(), t.GetKeyFile())
		if err != nil {
			return nil, fmt.Errorf("Error loading tls client certificate: %s", err)
		}

		cfg.Certificates = []tls.Certificate{cert}
	}

	// apply hardened TLS settings
	cfg.MinVersion = tlsMinVersion
	cfg.CurvePreferences = tlsCurvePreferences
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
		return nil, fmt.Errorf("No certificates could be parsed out of %s", path)
	}

	return caCertPool, nil
}
