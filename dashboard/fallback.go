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
	if len(path) == 0 || path == "/" {
		info := fallbackFileInfo{name: ".", body: []byte(f.msg), isDir: true}
		dir := fallbackFile{&info, bytes.NewReader(info.body)}
		return &dir, nil
	}

	info := fallbackFileInfo{name: path, body: []byte(f.msg)}
	file := fallbackFile{&info, bytes.NewReader(info.body)}
	return &file, nil
}

type fallbackFileInfo struct {
	name  string
	body  []byte
	isDir bool
}

func (f *fallbackFileInfo) Name() string               { return f.name }
func (f *fallbackFileInfo) Size() int64                { return int64(len(f.body)) }
func (f *fallbackFileInfo) Mode() os.FileMode          { return 0444 }
func (f *fallbackFileInfo) ModTime() time.Time         { return time.Now() }
func (f *fallbackFileInfo) IsDir() bool                { return f.isDir }
func (f *fallbackFileInfo) Sys() interface{}           { return nil }
func (f *fallbackFileInfo) Stat() (os.FileInfo, error) { return f, nil }
func (f *fallbackFileInfo) Readdir(count int) ([]os.FileInfo, error) {
	if f.isDir {
		newFile := fallbackFileInfo{name: "index.html", body: f.body}
		return []os.FileInfo{&newFile}, nil
	}
	return nil, fmt.Errorf("cannot Readdir from file %s", f.name)
}

type fallbackFile struct {
	*fallbackFileInfo
	*bytes.Reader
}

func (f *fallbackFile) Close() error {
	return nil
}
