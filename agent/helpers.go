package agent

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// See https://golang.org/pkg/os/exec/#LookPath
func lookPath(file string, env []string) (string, error) {
	if strings.Contains(file, "/") {
		err := findExecutable(file)
		if err == nil {
			return file, nil
		}
		return "", &exec.Error{Name: file, Err: err}
	}
	var path string
	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			path = strings.TrimPrefix(e, "PATH=")
		}
	}
	for _, dir := range filepath.SplitList(path) {
		if dir == "" {
			// Unix shell semantics: path element "" means "."
			dir = "."
		}
		path := filepath.Join(dir, file)
		if err := findExecutable(path); err == nil {
			return path, nil
		}
	}
	return "", &exec.Error{Name: file, Err: exec.ErrNotFound}
}

// See https://golang.org/src/os/exec/lp_unix.go
func findExecutable(file string) error {
	d, err := os.Stat(file)
	if err != nil {
		return err
	}
	if m := d.Mode(); !m.IsDir() && m&0111 != 0 {
		return nil
	}
	return os.ErrPermission
}
