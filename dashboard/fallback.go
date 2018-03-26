package dashboard

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"
)

type fallbackFS struct {
	msg string
}

func (f *fallbackFS) Open(path string) (http.File, error) {
	info := fallbackFileInfo{name: path, body: []byte(f.msg)}
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
