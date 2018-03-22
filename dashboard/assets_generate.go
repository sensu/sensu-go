//go:generate yarn install
//go:generate yarn build
//go:generate go run assets_generate.go

package dashboard

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Assets implements http.FileSystem returning web UI's assets.
var Assets http.FileSystem = &fallbackFS{}

//
// -----------------------------------------------------------------------------
// Fallback
// -----------------------------------------------------------------------------
//
// To allow for builds of the Sensu web UI without requiring the end-user to
// have node.js and yarn installed. Or, when a developer may want to quickly
// build the backend without re-building the entire dashboard.
//
// Fallback simply provides a empty filesystem implementation that returns a
// message informing the user that the dashboard is not present in the build.
//

const fallbackMessage = `
Sensu web UI was not included in this build.
Find build instructions in Sensu Github repository or use a pre-built binary.
`

type fallbackFS struct{}

func (*fallbackFS) Open(path string) (http.File, error) {
	info := fallbackFileInfo{name: path, body: []byte(fallbackMessage)}
	file := fallbackFile{&info, bytes.NewReader(info.body)}
	return &file, nil
}

type fallbackFileInfo struct {
	name string
	body []byte
}

func (f *fallbackFileInfo) Name() string               { return f.name }
func (f *fallbackFileInfo) Size() int64                { return int64(len(f.body)) }
func (f *fallbackFileInfo) Mode() os.FileMode          { return 0444 }
func (f *fallbackFileInfo) ModTime() time.Time         { return time.Now() }
func (f *fallbackFileInfo) IsDir() bool                { return false }
func (f *fallbackFileInfo) Sys() interface{}           { return nil }
func (f *fallbackFileInfo) Stat() (os.FileInfo, error) { return f, nil }
func (f *fallbackFileInfo) Readdir(count int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("cannot Readdir from file %s", f.name)
}

type fallbackFile struct {
	*fallbackFileInfo
	*bytes.Reader
}

func (f *fallbackFile) Close() error {
	return nil
}
