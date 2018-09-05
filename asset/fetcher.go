package asset

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// A Fetcher fetches a file from the specified source and returns an *os.File
// with the contents of the file found at source.
type Fetcher interface {
	Fetch(source string) (*os.File, error)
}

// URLGetter gets all content at the specified URL.
type URLGetter interface {
	Get(string) (io.ReadCloser, error)
}

// HTTPURLGetter uses an *http.Client to fetch content at the
// given URL.
type HTTPURLGetter struct {
	Timeout time.Duration
}

// Get the target URL and return an io.ReadCloser
func (h *HTTPURLGetter) Get(url string) (io.ReadCloser, error) {
	client := &http.Client{Timeout: h.Timeout}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching asset: %s", err.Error())
	}

	return resp.Body, nil
}

// An HTTPFetcher fetches the contents of files at a given URL.
type HTTPFetcher struct {
	Timeout time.Duration
	URLGetter
}

// InjectGetter injects the specified URLGetter.
func (h *HTTPFetcher) InjectGetter(g URLGetter) {
	h.URLGetter = g
}

// Fetch the file found at the specified url, and return the file or an
// error indicating why the fetch failed.
func (h *HTTPFetcher) Fetch(url string) (*os.File, error) {
	if h.URLGetter == nil {
		h.InjectGetter(&HTTPURLGetter{})
	}

	resp, err := h.Get(url)
	if err != nil {
		return nil, err
	}

	// Write response to tmp
	tmpFile, err := ioutil.TempFile(os.TempDir(), "sensu-asset")
	if err != nil {
		return nil, fmt.Errorf("can't open tmp file for asset")
	}

	if _, err = io.Copy(tmpFile, resp); err != nil {
		return nil, fmt.Errorf("error downloading asset")
	}

	return tmpFile, resetFile(tmpFile)
}

// resetFile ensures file contents are synced and rewound
func resetFile(f *os.File) error {
	if err := f.Sync(); err != nil {
		return err
	}
	_, err := f.Seek(0, 0)
	return err
}
