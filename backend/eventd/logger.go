package eventd

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"syscall"
	"time"

	"github.com/sensu/sensu-go/backend/logging"
	"github.com/sensu/sensu-go/backend/messaging"
)

// FileLogger is a rotatable logger.
type FileLogger struct {
	Path       string
	BufferSize int
	BufferWait time.Duration
	Bus        messaging.MessageBus
	notify     chan interface{}
	rawLogger  *rawLogger
}

// Start replaces the core event logger with the enteprise one, which logs
// events to a log file
func (f *FileLogger) Start() {
	if f.Path == "" {
		logger.Info("no event log file specified, event logging is disabled")
		return
	}

	f.notify = make(chan interface{}, 1)
	consumerName := fmt.Sprintf("filelogger://%s", f.Path)
	f.Bus.Subscribe(messaging.SignalTopic(syscall.SIGHUP), consumerName, f)

	rawLogger, err := newRawLogger(f.Path, f.BufferSize, f.BufferWait, f.notify)
	if err != nil {
		logger.WithError(err).Error("could not start event logging")
		return
	}

	f.rawLogger = rawLogger

	logger.Infof("event logging is now enabled and writing to %q", f.Path)

	// Start the ring buffer
	go rawLogger.ringBuffer()
	// Listen to the output channel of the ring buffer and write it to the log
	go rawLogger.write()
}

// Receiver implements messaging.Subscriber
func (f *FileLogger) Receiver() chan<- interface{} {
	return f.notify
}

func (f *FileLogger) Stop() {
	f.rawLogger.Stop()
}

func (f *FileLogger) Println(v interface{}) {
	f.rawLogger.Println(v)
}

type LogWriter interface {
	io.WriteCloser
	Sync() error
}

// rawLogger represents the raw events logger and consists of a ring buffer and
// a writer
type rawLogger struct {
	input  chan interface{}
	output chan interface{}
	writer LogWriter
	wait   time.Duration
}

// newRawLogger initializes the raw event logger
func newRawLogger(path string, bufferSize int, bufferWait time.Duration, sighup chan interface{}) (*rawLogger, error) {
	l := &rawLogger{
		input:  make(chan interface{}),
		output: make(chan interface{}, bufferSize),
		wait:   bufferWait,
	}

	writer, err := logging.NewRotateWriter(path, sighup)
	if err != nil {
		return nil, err
	}

	l.writer = writer

	return l, nil
}

// Println takes a raw event and sends it over to the ring buffer
func (l *rawLogger) Println(v interface{}) {
	l.input <- v
}

// Stop ends the ring buffer by closing the input channel, which in turns closes
// the output channel
func (l *rawLogger) Stop() {
	close(l.input)
}

// ringBuffer forwards events from the input channel to the output buffered
// channel and eliminates the oldest events when full
func (l *rawLogger) ringBuffer() {
	var eventsDropped int
	mu := sync.Mutex{}

	// Keep an eye on the eventsDropper counter and log an error when events are
	// dropped
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	defer close(l.output)
	go func() {
		for range ticker.C {
			mu.Lock()
			if eventsDropped > 0 {
				logger.Errorf("the event buffer is full, %d event(s) lost", eventsDropped)
				eventsDropped = 0
			}
			mu.Unlock()
		}
	}()

	for v := range l.input {
		select {
		case l.output <- v:
		case <-time.After(l.wait):
			// The new event could not be placed on the outgoing channel, therefore
			// take the oldest event in the buffer, drop it, and place the new event
			// at its place
			<-l.output
			l.output <- v

			// Increment the eventsDropped counter
			mu.Lock()
			eventsDropped++
			mu.Unlock()
		}
	}
}

// write reads events from the ring buffer and sends them over the writer
func (l *rawLogger) write() {
	encoder := json.NewEncoder(l.writer)
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	defer func() {
		// At this point the output channel was closed, which means the writer needs
		// to clean up after itself
		if err := l.writer.Close(); err != nil {
			logger.WithError(err).Error("could not close the event log file")
		}
	}()
	for {
		select {
		case v, ok := <-l.output:
			if !ok {
				return
			}
			if err := encoder.Encode(v); err != nil {
				logger.WithError(err).Warning("could not encode event")
				continue
			}
		case <-ticker.C:
			if err := l.writer.Sync(); err != nil {
				logger.WithError(err).Error("error syncing event log")
			}
		}
	}
}
