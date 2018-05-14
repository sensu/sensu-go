package main

import (
	"bytes"
	"net"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

func (s *Server) load() {
	for {
		select {
		case <-s.start:
			atomic.StoreInt64(&s.Stats.NumMetrics, 0)
			atomic.StoreInt64(&s.Stats.NumPackets, 0)
			s.Stats.StartTime = time.Now()
			atomic.StoreInt32(&s.Started, 1)

			go func() {
				for worker := 0; worker < s.Concurrency; worker++ {
					go func() {
						conn, err := net.Dial("udp", s.MetricsAddr)
						if err != nil {
							log.Panicf("Error connecting to statsd backend: %v", err)
						}
						defer conn.Close()
						buf := new(bytes.Buffer)
						for atomic.LoadInt32(&s.Started) != 0 {
							s.writeLines(conn, buf)
							if buf.Len() > 0 {
								s.write(conn, buf)
							}
							time.Sleep(100 * time.Microsecond)
						}
					}()
				}
			}()
		case <-s.stop:
			atomic.StoreInt32(&s.Started, 0)
			s.gatherStats()
		}
	}
}
