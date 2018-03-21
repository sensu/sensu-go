package dashboard

import (
	"net/http"
	"os"
)

// Assets implements http.FileSystem returning web UI's assets.
var Assets http.FileSystem = &emptyFS{}

type emptyFS struct{}

func (*emptyFS) Open(path string) (http.File, error) {
	return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
}
