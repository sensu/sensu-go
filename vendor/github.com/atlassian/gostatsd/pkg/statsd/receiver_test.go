package statsd

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/atlassian/gostatsd/pkg/fakesocket"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkReceive(b *testing.B) {
	// Small values result in the channel aggressively blocking and causing slowdowns.
	// Large values result in the channel consuming lots of memory before the scheduler
	// gets to it, causing GC related slowdowns.
	//
	// ... so this is pretty arbitrary.
	ch := make(chan []*Datagram, 5000)
	mr := NewDatagramReceiver(ch, DefaultReceiveBatchSize)
	c, done := fakesocket.NewCountedFakePacketConn(uint64(b.N))

	var wg sync.WaitGroup
	// runtime.GOMAXPROCS() is listed as "will go away".
	// The **intent** is to get the current -cpu value per:
	// https://github.com/golang/go/blob/release-branch.go1.9/src/testing/benchmark.go#L430
	numProcs := runtime.GOMAXPROCS(0)
	wg.Add(numProcs + 1)
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case dgs := <-ch:
				for _, dg := range dgs {
					dg.DoneFunc()
				}
			case <-ctx.Done():
				wg.Done()
				return
			}
		}
	}()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < numProcs; i++ {
		go func() {
			// mr.Receive() will loop until it runs out of packets to read, then it will
			// start having errors because the socket is closed.  When the socket is closed,
			// the read of the done channel will return allowing the context to be cancelled.
			mr.Receive(ctx, c)
			wg.Done()
		}()
	}

	<-done
}

func TestDatagramReceiver_Receive(t *testing.T) {
	ch := make(chan []*Datagram, 1)
	mr := NewDatagramReceiver(ch, 2)
	c := fakesocket.NewFakePacketConn()

	ctx, cancel := context.WithCancel(context.Background())

	go mr.Receive(ctx, c)

	var dgs []*Datagram
	select {
	case dgs = <-ch:
	case <-time.After(time.Second):
		t.Errorf("Timeout, failed to read datagram")
	}

	cancel()

	require.Len(t, dgs, 1)

	dg := dgs[0]
	assert.Equal(t, string(dg.IP), fakesocket.FakeAddr.IP.String())
	assert.Equal(t, dg.Msg, fakesocket.FakeMetric)
}
