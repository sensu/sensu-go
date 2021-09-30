package logging

import (
	"os"
	"sync"
)

// RotateWriter is a special writer that re-opens the path it was opened at
// when it receives a rotate signal.
type RotateWriter struct {
	file      *os.File
	isSpecial bool
	mu        sync.Mutex
	path      string
	rotate    chan interface{}
}

// NewRotateWriter creates a new RotateWriter. It will open the path given and
// use it for writing. It will also create a goroutine that will attempt to
// receive from the rotate channel. Whenever a message is sent on the rotate
// channel, the writer will close the currently open file and re-open it at
// the given path. When the rotate channel is closed, the goroutine started
// by this function will terminate.
func NewRotateWriter(path string, rotate chan interface{}) (*RotateWriter, error) {
	writer := &RotateWriter{
		path:   path,
		rotate: rotate,
	}
	if err := writer.open(); err != nil {
		return nil, err
	}
	go writer.listenSignal()
	return writer, nil
}

// Close will stop the signal listener and close the opened file.
func (w *RotateWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}

// Write will dispatch writes to the currently-opened file. It is goroutine-safe.
func (w *RotateWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Write(b)
}

// Sync syncs the currently opened file.
func (w *RotateWriter) Sync() error {
	if w.isSpecial {
		return nil
	} else {
		return w.file.Sync()
	}
}

// listenSignal listens for the HUP signal and re-opens the log file once
// received
func (w *RotateWriter) listenSignal() {
	for range w.rotate {
		logger.Infof("reopening log file %q", w.path)
		if err := w.open(); err != nil {
			logger.WithError(err).Errorf("error reopening log file %q", w.path)
		}
	}
}

// open will open a file and assign it to the writer underlying file. It can
// also re-open a file.
func (w *RotateWriter) open() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Close the file handle in case we already had a file open
	_ = w.file.Close()

	// Open the file and keep it in the writer
	fp, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	w.file = fp

	info, err := w.file.Stat()
	if err != nil {
		return err
	}

	if !info.Mode().IsRegular() {
		w.isSpecial = true
	}

	return nil
}
