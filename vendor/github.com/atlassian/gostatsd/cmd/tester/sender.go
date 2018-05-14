package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"sync/atomic"
	"time"

	"github.com/atlassian/gostatsd"

	log "github.com/sirupsen/logrus"
)

// Metrics store the metrics to send.
type Metrics []*gostatsd.Metric

var metrics = &Metrics{
	{
		Name: "statsd.tester.counter",
		Type: gostatsd.COUNTER,
	},
	{
		Name: "statsd.tester.gauge",
		Type: gostatsd.GAUGE,
	},
	{
		Name: "statsd.tester.timer",
		Type: gostatsd.TIMER,
	},
	{
		Name: "statsd.tester.set",
		Type: gostatsd.SET,
	},
}

func (s *Server) write(conn net.Conn, buf *bytes.Buffer) {
	_, err := buf.WriteTo(conn)
	if err != nil {
		log.Errorf("Error sending to statsd backend: %v", err)
	}
	atomic.AddInt64(&s.Stats.NumPackets, 1)
}

func (s *Server) writeLine(conn net.Conn, buf *bytes.Buffer, format, name string, value ...interface{}) {
	_, _ = fmt.Fprintf(buf, format, value...)
	atomic.AddInt64(&s.Stats.NumMetrics, 1)
	// Make sure we don't go over max udp datagram size of 1500
	if buf.Len() > s.MaxPacketSize {
		log.Debugf("Buffer length: %d", buf.Len())
		s.write(conn, buf)
		buf.Reset()
	}
}

func (s *Server) writeLines(conn net.Conn, buf *bytes.Buffer) {
	for _, metric := range *metrics {
		num := rand.Intn(10000) // Randomize metric name
		switch metric.Type {
		case gostatsd.COUNTER:
			value := rand.Float64() * 100
			s.writeLine(conn, buf, "%s%s_%d:%f|c\n", s.Namespace, metric.Name, num, value)
		case gostatsd.TIMER:
			n := rand.Intn(9) + 1
			for i := 0; i < n; i++ {
				value := rand.Float64() * 100
				s.writeLine(conn, buf, "%s%s_%d:%f|ms\n", s.Namespace, metric.Name, num, value)
			}
		case gostatsd.GAUGE:
			value := rand.Float64() * 100
			s.writeLine(conn, buf, "%s%s_%d:%f|g\n", s.Namespace, metric.Name, num, value)
		case gostatsd.SET:
			for i := 0; i < 100; i++ {
				value := rand.Intn(9) + 1
				s.writeLine(conn, buf, "%s%s_%d:%d|s\n", s.Namespace, metric.Name, num, value)
			}
		}
	}
}

func (s *Server) send() {
	var flushTicker *time.Ticker

	for {
		select {
		case <-s.start:
			flushTicker = time.NewTicker(s.FlushInterval)
			atomic.StoreInt64(&s.Stats.NumMetrics, 0)
			atomic.StoreInt64(&s.Stats.NumPackets, 0)
			s.Stats.StartTime = time.Now()
			atomic.StoreInt32(&s.Started, 1)
			go func() {
				conn, err := net.Dial("udp", s.MetricsAddr)
				if err != nil {
					log.Panicf("Error connecting to statsd backend: %v", err)
				}
				defer conn.Close()
				buf := new(bytes.Buffer)
				for t := range flushTicker.C {
					log.Debugf("Tick at %v", t)
					s.writeLines(conn, buf)
					if buf.Len() > 0 {
						s.write(conn, buf)
					}
				}
			}()
		case <-s.stop:
			atomic.StoreInt32(&s.Started, 0)
			flushTicker.Stop()
			s.gatherStats()
		}
	}
}
