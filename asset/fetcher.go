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

// An HTTPFetcher fetches the contents of files at a given URL.
type HTTPFetcher struct{}

// Fetch the file found at the specified url, and return the file or an
// error indicating why the fetch failed.
func (h HTTPFetcher) Fetch(url string) (*os.File, error) {
	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching asset: %s", err.Error())
	}

	// Write response to tmp
	tmpFile, err := ioutil.TempFile(os.TempDir(), "sensu-asset")
	if err != nil {
		return nil, fmt.Errorf("can't open tmp file for asset")
	}

	if _, err = io.Copy(tmpFile, resp.Body); err != nil {
		return nil, fmt.Errorf("error downloading asset")
	}

	return tmpFile, resetFile(tmpFile)
}
