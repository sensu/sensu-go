package asset

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// A Collection is an extensible set of assets.
type Collection struct {
	// reads should be frequent and writes very infrequent
	bs *atomic.Value
	mu *sync.Mutex
}

// NewCollection returns new instance of a Collection.
func NewCollection() *Collection {
	atom := &atomic.Value{}
	atom.Store([]http.FileSystem{})
	return &Collection{
		bs: atom,
		mu: &sync.Mutex{},
	}
}

// Extend collection with given set of assets
func (a *Collection) Extend(b http.FileSystem) {
	a.mu.Lock()
	defer a.mu.Unlock()
	bs := a.bs.Load().([]http.FileSystem)
	a.bs.Store(append([]http.FileSystem{b}, bs...))
}

// Open return file for for given path
func (a *Collection) Open(path string) (http.File, error) {
	bs := a.bs.Load().([]http.FileSystem)
	for _, b := range bs {
		f, err := b.Open(path)
		// if the file was not found continue looping
		if f == nil || isNotFound(err) {
			continue
		} else if err != nil {
			return nil, err
		}
		// If the file descriptor is a directory, create virtual representation
		if s, err := f.Stat(); err != nil {
			return nil, err
		} else if s.IsDir() {
			return makeVDir(bs, path)
		}
		return f, err
	}
	// if we've exhausted all the bundles return not found err
	return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
}

// make virtual directory from collection of filesystems
func makeVDir(fss []http.FileSystem, path string) (*Dir, error) {
	dir := &Dir{}
	dir.entries = map[string]os.FileInfo{}
	dir.DirInfo = &DirInfo{}
	for _, fs := range fss {
		f, err := fs.Open(path)
		if f == nil || isNotFound(err) {
			continue
		}
		if err != nil {
			return dir, err
		}
		stat, err := f.Stat()
		if err != nil || !stat.IsDir() {
			return dir, err
		}
		if dir.name == "" {
			dir.name = stat.Name()
		}
		if es, _ := f.Readdir(-1); es != nil {
			dir.addEntries(es)
		}
		if stat.ModTime().After(dir.modTime) {
			dir.modTime = stat.ModTime()
		}
	}
	return dir, nil
}

// DirInfo is a static definition of a directory.
type DirInfo struct {
	name    string
	modTime time.Time
}

func (d *DirInfo) Read([]byte) (int, error) {
	return 0, fmt.Errorf("cannot Read from directory %s", d.name)
}
func (d *DirInfo) Close() error               { return nil }
func (d *DirInfo) Stat() (os.FileInfo, error) { return d, nil }

func (d *DirInfo) Name() string       { return d.name }
func (d *DirInfo) Size() int64        { return 0 }
func (d *DirInfo) Mode() os.FileMode  { return 0755 | os.ModeDir }
func (d *DirInfo) ModTime() time.Time { return d.modTime }
func (d *DirInfo) IsDir() bool        { return true }
func (d *DirInfo) Sys() interface{}   { return nil }

// Dir is an opened dir instance.
type Dir struct {
	*DirInfo
	pos     int // Position within entries for Seek and Readdir.
	entries map[string]os.FileInfo
}

func (d *Dir) Seek(offset int64, whence int) (int64, error) {
	if offset == 0 && whence == io.SeekStart {
		d.pos = 0
		return 0, nil
	}
	return 0, fmt.Errorf("unsupported Seek in directory %s", d.name)
}

func (d *Dir) Readdir(count int) ([]os.FileInfo, error) {
	entries := d.entryValues()
	if d.pos >= len(entries) && count > 0 {
		return nil, io.EOF
	}
	if count <= 0 || count > len(entries)-d.pos {
		count = len(entries) - d.pos
	}
	e := entries[d.pos : d.pos+count]
	d.pos += count
	return e, nil
}

func (d *Dir) addEntries(fs []os.FileInfo) {
	for _, f := range fs {
		d.entries[f.Name()] = f
	}
}

func (d *Dir) entryValues() []os.FileInfo {
	vals := make([]os.FileInfo, 0, len(d.entries))
	for _, val := range d.entries {
		vals = append(vals, val)
	}
	return vals
}

func isNotFound(err error) bool {
	if err, ok := err.(*os.PathError); ok {
		if err.Err == os.ErrNotExist {
			return true
		}
	}
	return false
}
