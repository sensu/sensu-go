package asset

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

// maxDrainBytes caps how much of a non-200 response body is read before the
// connection is closed. Draining lets the connection be reused, but an
// unbounded read would let a hostile or broken server stall the fetch.
const maxDrainBytes = 4 << 10

// assetIdleConnTimeout bounds how long an idle connection to an asset host is
// kept open before it is closed and its file descriptor released.
const assetIdleConnTimeout = 90 * time.Second

// cachedClient is the client for one trusted CA file, with the digest of the
// file content it was built from.
type cachedClient struct {
	digest string
	client *http.Client
}

// clientCache holds one client per trusted CA file path. The stored digest is
// of the file content, so when the file is rotated the old client is retired
// and its idle connections are closed. Retiring the client matters: a pooled
// TLS connection skips certificate verification, so keeping the old client
// would keep trusting a server that the new CA no longer covers.
var clientCache = struct {
	sync.Mutex
	clients map[string]cachedClient
}{clients: make(map[string]cachedClient)}

// newTLSConfig builds the TLS configuration for an asset client. It is set by
// the build-tagged files: the standard build uses the default configuration and
// the FIPS build restricts the curve preferences.
var newTLSConfig = func(rootCAs *x509.CertPool) *tls.Config {
	return &tls.Config{RootCAs: rootCAs}
}

// defaultClient is the shared client for fetches without a trusted CA file. Its
// TLS configuration comes from newTLSConfig, so the FIPS build restricts the
// curves here as well as for CA-file clients.
var (
	defaultClientOnce sync.Once
	defaultClient     *http.Client
)

// getClient returns the HTTP client for the given trusted CA file. The file is
// read on every call, so a CA that is rotated, or that appears after the agent
// started, is picked up on the next fetch.
//
// An empty trustedCAFile returns the shared default client. If the file cannot
// be read or parsed, getClient returns an error and caches nothing, so the next
// fetch tries again.
func getClient(trustedCAFile string) (*http.Client, error) {
	if trustedCAFile == "" {
		defaultClientOnce.Do(func() {
			transport := http.DefaultTransport.(*http.Transport).Clone()
			transport.Proxy = http.ProxyFromEnvironment
			transport.TLSClientConfig = newTLSConfig(nil)
			transport.IdleConnTimeout = assetIdleConnTimeout
			defaultClient = &http.Client{Transport: transport}
		})
		return defaultClient, nil
	}

	certs, err := os.ReadFile(trustedCAFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read trusted CA file %s: %w", trustedCAFile, err)
	}

	sum := sha256.Sum256(certs)
	digest := hex.EncodeToString(sum[:])

	clientCache.Lock()
	defer clientCache.Unlock()

	if cached, ok := clientCache.clients[trustedCAFile]; ok && cached.digest == digest {
		return cached.client, nil
	}

	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		logger.WithError(err).Error("failed to retrieve system cert pool")
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
		return nil, fmt.Errorf("no certificates found in trusted CA file %s", trustedCAFile)
	}

	appendCerts(rootCAs)

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = http.ProxyFromEnvironment
	transport.TLSClientConfig = newTLSConfig(rootCAs)
	transport.IdleConnTimeout = assetIdleConnTimeout

	// Retire the previous client for this path. Closing its idle connections
	// stops them from being reused under the old trust.
	if old, ok := clientCache.clients[trustedCAFile]; ok {
		old.client.CloseIdleConnections()
	}

	client := &http.Client{Transport: transport}
	clientCache.clients[trustedCAFile] = cachedClient{digest: digest, client: client}
	return client, nil
}
