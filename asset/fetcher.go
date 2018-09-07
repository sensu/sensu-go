package asset

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var (
	defaultHTTPGetTimeout = 30 * time.Second
	defaultFetcher        = &httpFetcher{}
)

// A Fetcher fetches a file from the specified source and returns an *os.File
// with the contents of the file found at source.
type Fetcher interface {
	Fetch(source string) (*os.File, error)
}

// URLGetter gets all content at the specified URL.
type urlGetter func(string) (io.ReadCloser, error)

// Get the target URL and return an io.ReadCloser
func httpGet(url string) (io.ReadCloser, error) {
	client := &http.Client{Timeout: defaultHTTPGetTimeout}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching asset: %s", err.Error())
	}

	return resp.Body, nil
}

// An HTTPFetcher fetches the contents of files at a given URL.
type httpFetcher struct {
	URLGetter urlGetter
}

// Fetch the file found at the specified url, and return the file or an
// error indicating why the fetch failed.
func (h *httpFetcher) Fetch(url string) (*os.File, error) {
	if h.URLGetter == nil {
		h.URLGetter = httpGet
	}

	resp, err := h.URLGetter(url)
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
