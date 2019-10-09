package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sensu/lasr"
	bolt "go.etcd.io/bbolt"
)

type queue interface {
	Close() error
	Compact() error
	Receive(context.Context) (*lasr.Message, error)
	Send([]byte) (lasr.ID, error)
}

func newMemoryQueue(size int) *memoryQueue {
	return &memoryQueue{
		queue: make(chan *lasr.Message, size),
	}
}

type memoryQueue struct {
	id    lasr.Uint64ID
	mu    sync.Mutex
	queue chan *lasr.Message
}

func (m *memoryQueue) Close() error {
	return nil
}

func (m *memoryQueue) Compact() error {
	return nil
}

func (m *memoryQueue) Receive(ctx context.Context) (*lasr.Message, error) {
	return <-m.queue, nil
}

func (m *memoryQueue) Send(body []byte) (lasr.ID, error) {
	m.mu.Lock()
	id := m.id
	m.id++
	m.mu.Unlock()
	idBytes, err := id.MarshalBinary()
	if err != nil {
		return id, fmt.Errorf("couldn't send message to queue: %s", err)
	}

	message := &lasr.Message{
		Body: body,
		ID:   idBytes,
	}
	m.queue <- message
	return id, nil
}

func newQueue(path string) (queue, error) {
	if path == os.DevNull {
		return newMemoryQueue(1000), nil
	}
	if err := os.MkdirAll(path, 0744|os.ModeDir); err != nil {
		return nil, fmt.Errorf("could not create directory for api queue (%s): %s", path, err)
	}
	queuePath := filepath.Join(path, "queue.db")
	// Create and open the database for the queue. The FileMode given here (0600)
	// is only temporary since it will be enforced to 0600 when the queue is
	// compacted below by the queue.Compact method
	db, err := bolt.Open(queuePath, 0600, &bolt.Options{Timeout: 60 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("could not open api queue (%s): %s", queuePath, err)
	}
	queue, err := lasr.NewQ(db, "api-buffer")
	if err != nil {
		return nil, fmt.Errorf("error creating api queue: %s", err)
	}
	logger.Info("compacting api queue")
	defer logger.Info("finished api queue compaction")
	return queue, queue.Compact()
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
		return nil
	}
	return dst.Bytes()
}
