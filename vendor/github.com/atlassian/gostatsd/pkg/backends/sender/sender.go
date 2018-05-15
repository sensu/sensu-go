package sender

import (
	"bytes"
	"context"
	"net"
	"sync"
	"time"

	"github.com/atlassian/gostatsd"

	log "github.com/sirupsen/logrus"
)

const maxStreamsPerConnection = 100

type ConnFactory func() (net.Conn, error)

type Stream struct {
	Ctx context.Context
	Cb  gostatsd.SendCallback
	Buf <-chan *bytes.Buffer
}

type Sender struct {
	ConnFactory  ConnFactory
	Sink         chan Stream
	BufPool      sync.Pool
	WriteTimeout time.Duration
}

func (s *Sender) Run(ctx context.Context) {
	defer s.cleanup(ctx)
	var stream *Stream
	var errs []error
	defer func() {
		if stream != nil {
			stream.Cb(errs)
		}
	}()
	var sink <-chan Stream
	var streamCancel <-chan struct{}
	for {
		w, err := s.ConnFactory()
		if err != nil {
			log.Warnf("Failed to connect: %v", err)
			// TODO do backoff
			timer := time.NewTimer(1 * time.Second)
			for {
				if stream == nil {
					sink = s.Sink
				} else {
					streamCancel = stream.Ctx.Done()
				}
				select {
				case <-ctx.Done():
					timer.Stop()
					errs = append(errs, ctx.Err())
					return
				case st := <-sink:
					sink = nil
					stream = &st
					continue
				case <-streamCancel:
					stream.Cb(append(errs, stream.Ctx.Err()))
					stream = nil
					streamCancel = nil
					errs = nil
					continue
				case <-timer.C:
				}
				break
			}
			continue
		}
		if stream, errs, err = s.innerRun(ctx, w, stream, errs); err != nil {
			errs = append(errs, err)
			if err == context.Canceled || err == context.DeadlineExceeded {
				return
			}
		}
	}
}

func (s *Sender) innerRun(ctx context.Context, conn net.Conn, stream *Stream, errs []error) (*Stream, []error, error) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Warnf("Close failed: %v", err)
		}
	}()
	var err error
loop:
	for streamCount := 0; streamCount < maxStreamsPerConnection; streamCount++ {
		if stream == nil {
			select {
			case <-ctx.Done():
				err = ctx.Err()
				break loop
			case s := <-s.Sink:
				stream = &s
			}
		}
		for buf := range stream.Buf {
			if s.WriteTimeout > 0 {
				if e := conn.SetWriteDeadline(time.Now().Add(s.WriteTimeout)); e != nil {
					log.Warnf("Failed to set write deadline: %v", e)
				}
			}
			_, err = conn.Write(buf.Bytes())
			s.PutBuffer(buf)
			if err != nil {
				break loop
			}
		}
		stream.Cb(errs)
		stream = nil
		errs = nil
	}
	return stream, errs, err
}

func (s *Sender) GetBuffer() *bytes.Buffer {
	return s.BufPool.Get().(*bytes.Buffer)
}

func (s *Sender) PutBuffer(buf *bytes.Buffer) {
	buf.Reset() // Reset buffer before returning it into the pool
	s.BufPool.Put(buf)
}

func (s *Sender) cleanup(ctx context.Context) {
	close(s.Sink) // Sender must not be used after ctx is done
	for stream := range s.Sink {
		stream.Cb([]error{ctx.Err()})
	}
}
