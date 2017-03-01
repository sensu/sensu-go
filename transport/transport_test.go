package transport

import (
	"crypto/rand"
	"testing"
)

// This was all mostly to prove that performance of encoding/decoding was
// linear.

var (
	encodingTestMessages = map[int][]byte{}
)

func init() {
	sizes := []int{32, 64, 128, 1024, 32 * 1024, 128 * 1024}
	for _, sz := range sizes {
		encodingTestMessages[sz] = makeMessage(sz)
	}
}

func makeMessage(i int) []byte {
	msg := make([]byte, i)
	_, err := rand.Read(msg)
	if err != nil {
		panic(err)
	}
	return msg
}

func benchmarkEncode(i int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		Encode("type", encodingTestMessages[i])
	}
}

func BenchmarkEncode32(b *testing.B) {
	benchmarkEncode(32, b)
}

func BenchmarkEncode64(b *testing.B) {
	benchmarkEncode(64, b)
}

func BenchmarkEncode128(b *testing.B) {
	benchmarkEncode(128, b)
}

func BenchmarkEncode1k(b *testing.B) {
	benchmarkEncode(1024, b)
}

func BenchmarkEncode32k(b *testing.B) {
	benchmarkEncode(32*1024, b)
}

func BenchmarkEncode128k(b *testing.B) {
	benchmarkEncode(128*1024, b)
}
