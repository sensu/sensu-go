package bytes

import (
	"bytes"
	"sync"
)

type SyncBuffer struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

func (s *SyncBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *SyncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}
