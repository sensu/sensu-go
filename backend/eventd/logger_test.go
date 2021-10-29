// +build !windows

package eventd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sirupsen/logrus"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogger(t *testing.T) {
	// Hook ourselves into logrus to capture the log entries
	log, hook := test.NewNullLogger()
	logger = log.WithField("test", "TestLogger")

	bus, _ := messaging.NewWizardBus(messaging.WizardBusConfig{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = bus.Start(ctx)

	wt, _ := time.ParseDuration("10ms")
	l := &FileLogger{
		BufferSize: 1,
		BufferWait: wt,
		Bus:        bus,
	}

	// Providing an empty path should return an error
	l.Start()
	assert.Contains(t, hook.LastEntry().Message, "event logging is disabled")

	// Providing an invalid path should return an error
	l.Path = "/"
	l.Start()
	assert.Contains(t, hook.LastEntry().Message, "could not start event logging")

	// Providing a valid path should start the logger
	file, err := ioutil.TempFile(os.TempDir(), "event.*.log")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
	l.Path = file.Name()
	l.Start()
	assert.Contains(t, hook.LastEntry().Message, "event logging using 1 JSON encoder")
}

func TestNewRawLogger(t *testing.T) {
	// temporary file
	file, err := ioutil.TempFile(os.TempDir(), "event.*.log")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	tests := []struct {
		name       string
		path       string
		bufferSize int
		bufferWait string
		wantErr    bool
	}{
		{
			name:    "cannot open file",
			wantErr: true,
		},
		{
			name: "valid file",
			path: file.Name(),
		},
	}

	wt, _ := time.ParseDuration("10ms")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newRawLogger(tt.path, tt.bufferSize, wt, make(chan interface{}, 1))
			if (err != nil) != tt.wantErr {
				t.Errorf("newRawLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

type mockWriter struct {
	mock.Mock
	wg *sync.WaitGroup
}

func (w *mockWriter) Write(b []byte) (int, error) {
	args := w.Called(string(b))
	w.wg.Done()
	return args.Int(0), args.Error(1)
}

func (w *mockWriter) Close() error {
	_ = w.Called()
	return nil
}

func (w *mockWriter) Sync() error {
	return nil
}

type testHook struct {
	wg *sync.WaitGroup
}

func (h *testHook) Fire(entry *logrus.Entry) error {
	h.wg.Done()
	return nil
}

func (h *testHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.WarnLevel,
		logrus.ErrorLevel,
	}
}

func TestRawLogger(t *testing.T) {
	type writeFunc func(*mockWriter)

	tests := []struct {
		name     string
		msg      interface{}
		preMock  writeFunc
		postMock writeFunc
	}{
		{
			name: "JSON log message",
			msg:  json.RawMessage(`{"foo":"bar"}`),
			preMock: func(w *mockWriter) {
				w.wg.Add(1)
				w.On("Write", fmt.Sprintln(`{"foo":"bar"}`)).Return(0, nil)
				w.On("Close")
			},
			postMock: func(w *mockWriter) {
				w.AssertCalled(t, "Write", mock.AnythingOfType("string"))
			},
		},
		{
			name: "invalid JSON message",
			msg:  math.Inf(1),
			preMock: func(w *mockWriter) {
				w.wg.Add(1)
				w.On("Write", mock.AnythingOfType("string")).Return(0, nil)
				w.On("Close")
			},
			postMock: func(w *mockWriter) {
				w.AssertNotCalled(t, "Write", mock.AnythingOfType("string"))
			},
		},
		{
			name: "writer error",
			msg:  json.RawMessage(`{"foo":"bar"}`),
			preMock: func(w *mockWriter) {
				w.wg.Add(2)
				w.On("Write", mock.AnythingOfType("string")).Return(0, errors.New("err"))
				w.On("Close")
			},
			postMock: func(w *mockWriter) {
				w.AssertCalled(t, "Write", mock.AnythingOfType("string"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg := &sync.WaitGroup{}
			writer := &mockWriter{
				wg: wg,
			}
			log := logrus.New()
			log.AddHook(&testHook{wg: wg})
			logger = log.WithField("test", tt.name)
			tt.preMock(writer)

			// Once the waitgroup is done, close the c channel
			c := make(chan struct{})
			go func() {
				wg.Wait()
				c <- struct{}{}
			}()

			wt, _ := time.ParseDuration("10ms")
			l := &rawLogger{
				input:        make(chan interface{}),
				encoderInput: make(chan interface{}),
				output:       make(chan []byte, 1),
				writer:       writer,
				wait:         wt,
				metrics:      newMetrics(),
				done:         make(chan interface{}),
			}
			go l.ringBuffer()
			go l.encoder()
			go l.write()

			l.Println(tt.msg)

			// Wait for the waitgroup to close the c channel
			select {
			case <-c:
				l.Stop()
				tt.postMock(writer)
			case <-time.After(5 * time.Second):
				l.Stop()
				t.Fatal("timed out")
			}
		})
	}
}

type nilWriter struct{}

func (w *nilWriter) Write(b []byte) (int, error) {
	return 0, nil
}

func (w *nilWriter) Close() error {
	return nil
}

func (w *nilWriter) Sync() error {
	return nil
}

func TestRawLogger_ringBuffer(t *testing.T) {
	tests := []struct {
		name         string
		input        chan interface{}
		encoderInput chan interface{}
		output       chan []byte
		writer       LogWriter
		want         interface{}
		wantLog      bool
	}{
		{
			name:         "all messages are passed when within buffer size",
			input:        make(chan interface{}),
			encoderInput: make(chan interface{}, 5),
			output:       make(chan []byte, 5),
			writer:       &nilWriter{},
			want:         []interface{}{0, 1, 2, 3, 4},
		},
		{
			name:         "older messages are removed from the buffer when over the buffer size",
			input:        make(chan interface{}),
			encoderInput: make(chan interface{}, 4),
			output:       make(chan []byte, 4),
			writer:       &nilWriter{},
			want:         []interface{}{1, 2, 3, 4},
			wantLog:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nullLogger, hook := test.NewNullLogger()
			logger = nullLogger.WithField("test", tt.name)

			wt, _ := time.ParseDuration("10ms")
			l := &rawLogger{
				input:        tt.input,
				encoderInput: tt.encoderInput,
				output:       tt.output,
				writer:       tt.writer,
				wait:         wt,
				metrics:      newMetrics(),
				done:         make(chan interface{}),
			}
			go l.ringBuffer()

			// Send messages over input channel
			for i := 0; i < 5; i++ {
				l.input <- i
			}

			// Wait for any event to be logged
			time.Sleep(time.Second * 2)
			l.Stop()

			// Get the resulting messages sent over the output channel
			var results []interface{}
			for result := range l.encoderInput {
				results = append(results, result)
			}

			if !reflect.DeepEqual(results, tt.want) {
				t.Errorf("rawLogger.ringBuffer() = %#v, want %#v", results, tt.want)
			}

			if tt.wantLog && len(hook.AllEntries()) == 0 {
				t.Error("rawLogger.ringBuffer() expected a log entry, got 0")
			}
		})
	}
}

func TestRawLogger_encoder(t *testing.T) {
	type Tmp struct {
		Id int `json:"id"`
	}

	tests := []struct {
		name         string
		input        chan interface{}
		encoderInput chan interface{}
		output       chan []byte
		writer       LogWriter
		want         []string
	}{
		{
			name:         "all messages are passed when within buffer size",
			input:        make(chan interface{}),
			encoderInput: make(chan interface{}),
			output:       make(chan []byte, 5),
			writer:       &nilWriter{},
			want:         []string{"{\"id\":0}\n", "{\"id\":1}\n", "{\"id\":2}\n", "{\"id\":3}\n", "{\"id\":4}\n"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nullLogger, _ := test.NewNullLogger()
			logger = nullLogger.WithField("test", tt.name)

			l := &rawLogger{
				input:        tt.input,
				encoderInput: tt.encoderInput,
				output:       tt.output,
				writer:       tt.writer,
				wait:         0,
				metrics:      newMetrics(),
				done:         make(chan interface{}),
			}
			go l.ringBuffer()
			go l.encoder()

			// Send messages over input channel
			for i := 0; i < 5; i++ {
				l.encoderInput <- Tmp{i}
			}

			// Wait for any event to be logged
			time.Sleep(time.Second * 2)
			l.Stop()

			// Get the resulting messages sent over the output channel
			var results []string
			for result := range l.output {
				results = append(results, string(result))
			}

			if !reflect.DeepEqual(results, tt.want) {
				t.Errorf("RawLogger.encoder() = %#v, want %#v", results, tt.want)
			}
		})
	}
}
