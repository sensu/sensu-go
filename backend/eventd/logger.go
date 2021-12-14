package eventd

import (
	"bytes"
	"fmt"
	"io"
	"runtime"
	"sync"
	"syscall"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/sensu/sensu-go/backend/logging"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sirupsen/logrus"
)

// FileLogger is a rotatable logger.
type FileLogger struct {
	Path                 string
	BufferSize           int
	BufferWait           time.Duration
	Bus                  messaging.MessageBus
	ParallelJSONEncoding bool
	notify               chan interface{}
	rawLogger            *rawLogger
	subscription         messaging.Subscription
}

// Start replaces the core event logger with the enteprise one, which logs
// events to a log file
func (f *FileLogger) Start() error {
	f.notify = make(chan interface{}, 1)

	rawLogger, err := newRawLogger(f.Path, f.BufferSize, f.BufferWait, f.notify)
	if err != nil {
		return fmt.Errorf("could not start event logging: %v", err)
	}
	f.rawLogger = rawLogger

	consumerName := fmt.Sprintf("filelogger://%s", f.Path)
	subscription, err := f.Bus.Subscribe(messaging.SignalTopic(syscall.SIGHUP), consumerName, f)
	if err != nil {
		return fmt.Errorf("failed to subscribe event logger to SIGHUP: %v", err)
	}
	f.subscription = subscription

	// Start the encoders
	numEncoders := f.numEncoders()
	for i := 0; i < numEncoders; i++ {
		go rawLogger.encoder()
	}
	logger.Infof("event logging using %d JSON encoder", numEncoders)

	// Start the ring buffer
	go rawLogger.ringBuffer()
	// Listen to the output channel of the ring buffer and write it to the log
	go rawLogger.write()
	go rawLogger.metricsWriter()
	return nil
}

func (f *FileLogger) numEncoders() int {
	numEncoders := 1
	if f.ParallelJSONEncoding {
		numEncoders = runtime.NumCPU() / 2
		if numEncoders < 2 {
			numEncoders = 2
		}
	}
	return numEncoders
}

// Receiver implements messaging.Subscriber
func (f *FileLogger) Receiver() chan<- interface{} {
	return f.notify
}

func (f *FileLogger) Stop() {
	_ = f.subscription.Cancel()
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
	input        chan interface{}
	encoderInput chan interface{}
	output       chan []byte
	writer       LogWriter
	wait         time.Duration
	metrics      *metrics
	done         chan interface{}
}

// newRawLogger initializes the raw event logger
func newRawLogger(path string, bufferSize int, bufferWait time.Duration, sighup chan interface{}) (*rawLogger, error) {
	l := &rawLogger{
		input:        make(chan interface{}),
		encoderInput: make(chan interface{}, bufferSize),
		output:       make(chan []byte, bufferSize),
		done:         make(chan interface{}),
		wait:         bufferWait,
		metrics:      newMetrics(),
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
	close(l.done)
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
	defer close(l.encoderInput)
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
		case l.encoderInput <- v:
		case <-time.After(l.wait):
			dropped := false
			select {
			case <-l.encoderInput:
				dropped = true
			case l.encoderInput <- v:
			}
			if dropped {
				// Increment the eventsDropped counter
				mu.Lock()
				eventsDropped++
				mu.Unlock()
				l.encoderInput <- v
			}
		}
	}
}

func (l *rawLogger) encoder() {
	defer close(l.output)

	var buf bytes.Buffer
	encoder := jsoniter.NewEncoder(&buf)

	for input := range l.encoderInput {
		buf.Reset()
		if err := encoder.Encode(input); err != nil {
			logger.WithError(err).Warning("could not encode data")
			continue
		}
		b := buf.Bytes()
		dup := make([]byte, len(b))
		copy(dup, b)
		l.output <- dup
	}
}

// write reads events from the ring buffer and sends them over the writer
func (l *rawLogger) write() {
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
		case b, ok := <-l.output:
			if !ok {
				return
			}

			if _, err := l.writer.Write(b); err != nil {
				logger.WithError(err).Warning("could not write event")
				continue
			}
			l.metrics.Accumulate(1, len(b))
		case <-ticker.C:
			if err := l.writer.Sync(); err != nil {
				logger.WithError(err).Error("error syncing event log")
			}
		}
	}
}

func (l *rawLogger) metricsWriter() {
	ticker := time.NewTicker(time.Minute * 1)
	defer ticker.Stop()

	for {
		select {
		case _, ok := <-ticker.C:
			if !ok {
				return
			}
			metrics := l.metrics.computeMetrics()

			logger.WithFields(logrus.Fields{
				"count":    metrics.count,
				"bytes":    metrics.totalBytes,
				"rate":     fmt.Sprintf("%.2f", metrics.rate),
				"byteRate": fmt.Sprintf("%.2f", metrics.byteRate),
				"duration": metrics.seconds,
			}).Infof("METRICS: Event log writer")
		case <-l.done:
			return
		}
	}
}

type computedMetrics struct {
	count      int
	totalBytes int
	rate       float64
	byteRate   float64
	seconds    int
}

type metrics struct {
	count     int
	bytes     int
	startTime time.Time
	mtx       sync.Mutex
}

func newMetrics() *metrics {
	return &metrics{
		startTime: time.Now(),
	}
}

func (m *metrics) computeMetrics() *computedMetrics {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	now := time.Now()
	durationSec := now.Sub(m.startTime).Seconds()

	metrics := &computedMetrics{
		count:      m.count,
		totalBytes: m.bytes,
		rate:       float64(m.count) / durationSec,
		byteRate:   float64(m.bytes) / durationSec,
		seconds:    int(durationSec),
	}

	m.count = 0
	m.bytes = 0
	m.startTime = time.Now()

	return metrics
}

func (m *metrics) Accumulate(count int, bytes int) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.count += count
	m.bytes += bytes
}
