package asset

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"golang.org/x/time/rate"
)

const (
	// DefaultAssetsRateLimit defines the rate limit for assets fetched per second.
	// It equates to GitHub's user to server rate limit of 5000 requests per hour.
	DefaultAssetsRateLimit rate.Limit = 1.39

	// DefaultAssetsBurstLimit defines the burst ceiling for a rate limited asset fetch.
	// If 0, then the setting has no effect.
	DefaultAssetsBurstLimit int = 100
)

// A Fetcher fetches a file from the specified source and returns an *os.File
// with the contents of the file found at source.
type Fetcher interface {
	Fetch(ctx context.Context, source string, headers map[string]string) (*os.File, error)
}

// URLGetter gets all content at the specified URL.
type urlGetter func(context.Context, string, string, map[string]string) (io.ReadCloser, error)

// Get the target URL and return an io.ReadCloser
func httpGet(ctx context.Context, path, trustedCAFile string, headers map[string]string) (io.ReadCloser, error) {
	client := &http.Client{}

	if trustedCAFile != "" {
		rootCAs, err := x509.SystemCertPool()
		if err != nil {
			logger.WithError(err).Error("failed to retrieve system cert pool")
		}
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		certs, err := ioutil.ReadFile(trustedCAFile)
		if err != nil {
			logger.WithError(err).Errorf("failed to read trusted CA file: %s", trustedCAFile)
		}

		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			logger.Errorf("failed to append %s to RootCAs, using system certs only", trustedCAFile)
		}

		appendCerts(rootCAs)

		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: rootCAs,
				},
			},
		}
	}

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("error fetching asset: %s", err)
	}
	req = req.WithContext(ctx)

	for k, v := range headers {
		values := strings.Split(v, ",")
		for _, value := range values {
			req.Header.Add(k, strings.TrimSpace(value))
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching asset: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching asset: Response Code %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// An HTTPFetcher fetches the contents of files at a given URL.
type httpFetcher struct {
	URLGetter     urlGetter
	Limiter       *rate.Limiter
	trustedCAFile string
}

// Fetch the file found at the specified url, and return the file or an
// error indicating why the fetch failed.
func (h *httpFetcher) Fetch(ctx context.Context, url string, headers map[string]string) (*os.File, error) {
	if h.URLGetter == nil {
		h.URLGetter = httpGet
	}

	if h.Limiter != nil {
		if !h.Limiter.Allow() {
			return nil, fmt.Errorf("can't download asset due to rate limit")
		}
	}

	resp, err := h.URLGetter(ctx, url, h.trustedCAFile, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	// Write response to tmp
	tmpFile, err := ioutil.TempFile(os.TempDir(), "sensu-asset")
	if err != nil {
		return nil, fmt.Errorf("can't open tmp file for asset: %s", err)
	}

	buffered := bufio.NewWriter(tmpFile)
	if _, err = io.Copy(buffered, resp); err != nil {
		return nil, fmt.Errorf("error downloading asset: %s", err)
	}
	if err := buffered.Flush(); err != nil {
		return nil, fmt.Errorf("error downloading asset: %s", err)
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return nil, err
	}

	if _, err := tmpFile.Seek(0, 0); err != nil {
		tmpFile.Close()
		return nil, err
	}

	return tmpFile, nil
}
