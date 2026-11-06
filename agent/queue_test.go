package agent

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	corev2 "github.com/sensu/core/v2"
	bolt "go.etcd.io/bbolt"
)

func TestCompressionRoundTrip(t *testing.T) {
	event := corev2.FixtureEvent("foo", "bar")
	want, _ := json.Marshal(event)
	compressed := compressMessage(want)
	if len(compressed) == 0 {
		t.Fatal("compress failed")
	}
	if len(compressed) >= len(want) {
		t.Fatal("not smaller")
	}
	got := decompressMessage(compressed)
	if !bytes.Equal(got, want) {
		t.Fatalf("bad compress: got %v, want %v", got, want)
	}
}

func BenchmarkCompressEventRoundTrip(b *testing.B) {
	event := corev2.FixtureEvent("foo", "bar")
	msg, _ := json.Marshal(event)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c := compressMessage(msg)
			_ = decompressMessage(c)
		}
	})
}

func TestNewQueueReleasesDatabaseWhenCompactFails(t *testing.T) {
	dir := t.TempDir()
	// Compaction writes to <dir>/.lasr.temp.db; a directory there makes it fail.
	if err := os.Mkdir(filepath.Join(dir, ".lasr.temp.db"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := newQueue(dir); err == nil {
		t.Fatal("expected newQueue to fail when compaction fails")
	}
	db, err := bolt.Open(filepath.Join(dir, "queue.db"), 0600, &bolt.Options{Timeout: 500 * time.Millisecond})
	if err != nil {
		t.Fatalf("queue database is still locked after newQueue failed: %v", err)
	}
	_ = db.Close()
}
