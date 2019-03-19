package agent

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	bolt "github.com/coreos/bbolt"
	"github.com/sensu/lasr"
)

func newQueue(path string) (*lasr.Q, error) {
	if err := os.MkdirAll(path, 0644); err != nil {
		return nil, fmt.Errorf("error creating api queue: %s", err)
	}
	queuePath := filepath.Join(path, "queue.db")
	db, err := bolt.Open(queuePath, 0644, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating api queue: %s", err)
	}
	return lasr.NewQ(db, "api-buffer")
}

func compressMessage(message []byte) []byte {
	buf := new(bytes.Buffer)
	src := bytes.NewReader(message)
	dst := gzip.NewWriter(buf)
	_, _ = io.Copy(dst, src)
	_ = dst.Close()
	return buf.Bytes()
}

func decompressMessage(message []byte) []byte {
	dst := new(bytes.Buffer)
	src, _ := gzip.NewReader(bytes.NewReader(message))
	defer src.Close()
	_, err := io.Copy(dst, src)
	if err != nil {
		panic(err)
	}
	return dst.Bytes()
}
