package agent

import (
	"fmt"
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
