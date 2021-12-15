//go:build !windows
// +build !windows

package asset

import "crypto/x509"

func appendCerts(rootCAs *x509.CertPool) {}
