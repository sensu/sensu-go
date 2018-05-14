package statsdaemon

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

func getTLSConfiguration(caPath, certPath, keyPath string, enable bool) (*tls.Config, error) {
	if !enable {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		// Can't use SSLv3 because of POODLE and BEAST
		// Can't use TLSv1.0 because of POODLE and BEAST using CBC cipher
		// Can't use TLSv1.1 because of RC4 cipher usage
		MinVersion: tls.VersionTLS12,
	}

	if caPath != "" {
		caPEM, err := ioutil.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("[%s] error reading TLS CA: %v", BackendName, err)
		}

		tlsConfig.RootCAs = x509.NewCertPool()
		if ok := tlsConfig.RootCAs.AppendCertsFromPEM(caPEM); !ok {
			return nil, fmt.Errorf("[%s] error reading TLS CA: no certificates found", BackendName)
		}
	}

	if certPath != "" || keyPath != "" {
		if certPath == "" {
			return nil, fmt.Errorf("[%s] tls_cert_path is required when tls_key_path is set", BackendName)
		}
		if keyPath == "" {
			return nil, fmt.Errorf("[%s] tls_key_path is required when tls_cert_path is set", BackendName)
		}

		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, fmt.Errorf("[%s] error loading client certificate: %v", BackendName, err)
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	}

	return tlsConfig, nil
}
