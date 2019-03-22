package asset

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
)

// ListContents returns all paths inside the given filesystem; sub-directories
// are also expanded.
func ListContents(fs http.FileSystem, path string) ([]PathInfo, error) {
	paths := []PathInfo{}
	f, err := fs.Open(path)
	if f == nil || err != nil {
		return paths, err
	}
	s, err := f.Stat()
	if err != nil {
		return paths, err
	}
	if !s.IsDir() {
		return paths, errors.New("expected path to be directory")
	}
	files, err := f.Readdir(-1)
	defer f.Close()
	if err != nil {
		return paths, err
	}
	for _, file := range files {
		fPath := filepath.Join(path, file.Name())
		if file.IsDir() {
			sub, err := ListContents(fs, fPath)
			if err != nil {
				return paths, err
			}
			paths = append(paths, sub...)
		}
		paths = append(paths, PathInfo{Name: fPath, FileInfo: file})
	}
	return paths, nil
}

// PathInfo includes the file descriptor of a path.
type PathInfo struct {
	os.FileInfo
	Name string
}
