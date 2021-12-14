//go:build windows
// +build windows

package asset

import (
	"crypto/x509"
	"syscall"
	"unsafe"
)

const (
	// CRYPT_E_NOT_FOUND is an error code specific to windows cert pool.
	// See https://github.com/golang/go/issues/16736#issuecomment-540373689.
	CRYPT_E_NOT_FOUND = 0x80092004
)

func appendCerts(rootCAs *x509.CertPool) {
	storeHandle, err := syscall.CertOpenSystemStore(0, syscall.StringToUTF16Ptr("Root"))
	if err != nil {
		logger.WithError(err).Error(syscall.GetLastError())
	}

	var cert *syscall.CertContext
	for {
		cert, err = syscall.CertEnumCertificatesInStore(storeHandle, cert)
		if err != nil {
			if errno, ok := err.(syscall.Errno); ok {
				if errno == CRYPT_E_NOT_FOUND {
					break
				}
			}
			logger.WithError(err).Error(syscall.GetLastError())
		}
		if cert == nil {
			break
		}
		// Copy the buf, since ParseCertificate does not create its own copy.
		buf := (*[1 << 20]byte)(unsafe.Pointer(cert.EncodedCert))[:]
		buf2 := make([]byte, cert.Length)
		copy(buf2, buf)
		if c, err := x509.ParseCertificate(buf2); err == nil {
			rootCAs.AddCert(c)
		}
	}
}
