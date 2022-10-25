package agent

import (
	"bytes"
	"encoding/json"
	"testing"

	corev2 "github.com/sensu/core/v2"
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
