package eventd

import (
	"runtime"
	"strings"
	"testing"
	"time"
)

func ringBufferTickerGoroutines() int {
	buf := make([]byte, 1<<20)
	buf = buf[:runtime.Stack(buf, true)]
	return strings.Count(string(buf), "(*rawLogger).ringBuffer.func")
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestRingBufferTickerGoroutineExitsOnStop(t *testing.T) {
	before := ringBufferTickerGoroutines()

	const n = 10
	loggers := make([]*rawLogger, n)
	for i := range loggers {
		loggers[i] = &rawLogger{
			input:        make(chan interface{}),
			encoderInput: make(chan interface{}, 1),
			output:       make(chan []byte, 1),
			wait:         time.Millisecond,
			metrics:      newMetrics(),
			done:         make(chan interface{}),
		}
		go loggers[i].ringBuffer()
	}
	waitFor(t, "all ticker goroutines to start", func() bool { return ringBufferTickerGoroutines()-before == n })

	for _, l := range loggers {
		l.Stop()
	}
	waitFor(t, "ticker goroutines to exit after Stop", func() bool { return ringBufferTickerGoroutines()-before <= 0 })
}
