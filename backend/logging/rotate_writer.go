package logging

import (
	"fmt"
	"os"
	"sync"
)

const (
	// DefaultMaxFileSize is the default maximum log file size (100 MB) before
	// automatic rotation occurs.
	DefaultMaxFileSize int64 = 100 * 1024 * 1024

	// maxRotatedFiles is the maximum number of rotated log files to retain.
	maxRotatedFiles = 5
)

// RotateWriter is a special writer that re-opens the path it was opened at
// when it receives a rotate signal. It also supports automatic size-based
// rotation to prevent unbounded disk usage.
type RotateWriter struct {
	file        *os.File
	isSpecial   bool
	maxSize     int64
	mu          sync.Mutex
	path        string
	rotate      chan interface{}
	writtenSize int64
}

// NewRotateWriter creates a new RotateWriter. It will open the path given and
// use it for writing. It will also create a goroutine that will attempt to
// receive from the rotate channel. Whenever a message is sent on the rotate
// channel, the writer will close the currently open file and re-open it at
// the given path. When the rotate channel is closed, the goroutine started
// by this function will terminate.
func NewRotateWriter(path string, rotate chan interface{}) (*RotateWriter, error) {
	writer := &RotateWriter{
		path:    path,
		rotate:  rotate,
		maxSize: DefaultMaxFileSize,
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
// If the file exceeds the maximum size after writing, it is automatically rotated.
func (w *RotateWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n, err := w.file.Write(b)
	if err != nil {
		return n, err
	}

	w.writtenSize += int64(n)

	if !w.isSpecial && w.maxSize > 0 && w.writtenSize >= w.maxSize {
		w.rotateLocked()
	}

	return n, nil
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

// rotateLocked performs file rotation while holding the mutex.
// It renames the current file and opens a new one.
func (w *RotateWriter) rotateLocked() {
	_ = w.file.Close()

	// Shift existing rotated files
	for i := maxRotatedFiles - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", w.path, i)
		newPath := fmt.Sprintf("%s.%d", w.path, i+1)
		_ = os.Rename(oldPath, newPath)
	}
	_ = os.Rename(w.path, fmt.Sprintf("%s.1", w.path))

	// Open a fresh file
	fp, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.WithError(err).Errorf("error rotating log file %q", w.path)
		return
	}
	w.file = fp
	w.writtenSize = 0
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

	w.writtenSize = info.Size()

	return nil
}
