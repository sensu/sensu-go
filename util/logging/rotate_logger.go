package logging

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// RotateFileLoggerConfig is configuration for creating a RotateFileLogger.
//
// If Path is not specified, then the log will be created in the current
// working directory, based on the binary name. (e.g. sensu-backend.log)
//
// If MaxSizeBytes is not specified, then each log file segment will have at
// most 128 MB written to it.
//
// If RetentionDuration is not specified, then the logger will retain archived
// log files for an unlimited duration.
//
// If RetentionFiles is not specified, then the logger will retain an unbounded
// number of archived log files.
type RotateFileLoggerConfig struct {
	Path              string
	MaxSizeBytes      int64
	RetentionDuration time.Duration
	RetentionFiles    int64

	sync bool // for testing only
}

func (f *rotateFile) archive(currentName, archiveName string) (err error) {
	defer func() {
		e := os.Remove(archiveName)
		if err == nil {
			err = e
		}
	}()
	reader, err := os.Open(archiveName)
	if err != nil {
		return err
	}
	defer func() {
		e := reader.Close()
		if err == nil {
			err = e
		}
	}()
	zipFile, err := os.Create(archiveName + ".zip")
	if err != nil {
		return err
	}
	defer zipFile.Close()
	zipper := zip.NewWriter(zipFile)
	defer zipper.Close()
	zipWriter, err := zipper.Create(archiveName)
	if err != nil {
		return err
	}
	if _, err := io.Copy(zipWriter, reader); err != nil {
		return err
	}
	return nil
}

type rotateFile struct {
	count     int64
	max       int64
	once      sync.Once
	mu        sync.RWMutex
	container *atomic.Value
	file      *os.File
	sync      bool // only for testing purposes
}

func (f *rotateFile) Rotate() (*rotateFile, error) {
	// Ensure that all pending writes finish before starting rotation
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now().UnixNano()
	currentName := f.file.Name()
	archiveName := fmt.Sprintf("%s.%d", currentName, now)

	replacement := &rotateFile{
		max:       f.max,
		container: f.container,
		sync:      f.sync,
	}
	if err := f.Close(); err != nil {
		return nil, err
	}
	if err := os.Rename(currentName, archiveName); err != nil {
		return nil, err
	}
	var err error
	replacement.file, err = os.OpenFile(currentName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return f, err
	}

	if f.sync {
		if err := f.archive(currentName, archiveName); err != nil {
			return nil, err
		}
	} else {
		// archiver errors are silently ignored in production,
		// as there is nothing that can be done about them.
		go f.archive(currentName, archiveName)
	}

	return replacement, nil
}

func (f *rotateFile) Write(p []byte) (int, error) {
	f.mu.RLock()
	projected := atomic.AddInt64(&f.count, int64(len(p)))
	if projected <= f.max {
		defer f.mu.RUnlock()
		n, err := f.file.Write(p)
		if err == nil {
			err = f.file.Sync()
		}
		return n, err
	}
	f.mu.RUnlock()

	var err error
	f.once.Do(func() {
		var fr *rotateFile
		fr, err = f.Rotate()
		if err == nil {
			f.container.Store(fr)
		}
	})

	if err != nil {
		return 0, fmt.Errorf("error rotating log: %s", err)
	}
	replacement := f.container.Load().(*rotateFile)
	if replacement == f {
		return 0, errors.New("error rotating log")
	}
	return replacement.Write(p)
}

func (f *rotateFile) Close() error {
	return f.file.Close()
}

// RotateFileLogger presents a file-like interface, but dispatches writes to
// files based on its rotation configuration. Rotated log files are compressed
// in zip format, as this device is intended to be used for Windows.
//
// For information on how to configure RotateFileLogger, see RotateFileLoggerConfig.
type RotateFileLogger struct {
	retentionFiles    int64
	closed            int64
	retentionDuration time.Duration
	container         *atomic.Value
	path              string
}

// NewRotateFileLogger creates a new configured RotateFileLogger. It opens a
// file for writing at the path it was configured to use. If there was an error
// in opening the file for writing, it will be returned.
//
// If the logger attempts to open a file that already exists, then its size will
// be taken into account when doing rotation.
func NewRotateFileLogger(cfg RotateFileLoggerConfig) (*RotateFileLogger, error) {
	if cfg.Path == "" {
		cfg.Path = fmt.Sprintf("%s.log", os.Args[0])
	}
	if cfg.MaxSizeBytes == 0 {
		// 128 MB
		cfg.MaxSizeBytes = 1 << 27
	}
	w := &RotateFileLogger{
		path:              cfg.Path,
		retentionDuration: cfg.RetentionDuration,
		retentionFiles:    cfg.RetentionFiles,
		container:         new(atomic.Value),
	}
	var count int64
	fi, err := os.Stat(cfg.Path)
	if err == nil {
		count = fi.Size()
	}
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	// Open the log file for writing in append mode whether or not it exists
	f, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	fr := &rotateFile{
		file:      f,
		max:       cfg.MaxSizeBytes,
		count:     count,
		container: w.container,
		sync:      cfg.sync,
	}
	w.container.Store(fr)
	return w, nil
}

// StartReaper starts an asynchronous worker that lists the log files written
// on an interval, and deletes ones that don't meet the retention criteria
// configured in the RotateFileLoggerConfig.
//
// When the supplied context is cancelled, the reaper will stop reaping.
//
// Every iteration of reaping will require that the caller consume errors from
// the provided error channel. If the caller does not consumer the reaper
// errors, then the reaper will hang. Errors that are delivered to the error
// channel can be nil.
func (r *RotateFileLogger) StartReaper(ctx context.Context, interval time.Duration) <-chan error {
	errors := make(chan error, 1)
	go r.reapLoop(ctx, errors, interval)
	return errors
}

func (r *RotateFileLogger) reapLoop(ctx context.Context, errors chan error, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			close(errors)
			return
		case <-ticker.C:
			r.reap(errors)
		}
	}
}

func (r *RotateFileLogger) reap(errors chan error) {
	dir := filepath.Dir(r.path)
	f, err := os.Open(dir)
	if err != nil {
		errors <- err
		return
	}
	files, err := f.Readdirnames(0)
	if err != nil {
		errors <- err
		return
	}
	filesToReap := make([]string, 0, len(files))
	base := filepath.Base(r.path)
	reapRegexp := regexp.MustCompile(fmt.Sprintf(`^%s\.([0-9]+)(\.zip)?$`, regexp.QuoteMeta(base)))
	for _, file := range files {
		if reapRegexp.MatchString(file) {
			filesToReap = append(filesToReap, file)
		}
	}
	tooOld := make(map[string]bool, len(filesToReap))
	if r.retentionDuration > 0 {
		for _, file := range filesToReap {
			matches := reapRegexp.FindStringSubmatch(file)
			if len(matches) < 2 {
				continue
			}
			var timestamp int64
			if _, err := fmt.Sscanf(matches[1], "%d", timestamp); err != nil {
				continue
			}
			archiveTime := time.Unix(0, timestamp)
			if archiveTime.Add(r.retentionDuration).Before(time.Now()) {
				tooOld[file] = true
				if err := os.Remove(file); err != nil {
					errors <- err
				}
			}
		}
	}
	notTooOld := make([]string, 0, len(filesToReap))
	for _, file := range filesToReap {
		if !tooOld[file] || r.retentionDuration == 0 {
			notTooOld = append(notTooOld, file)
		}
	}
	if r.retentionFiles > 0 && int64(len(notTooOld)) > r.retentionFiles {
		sort.Sort(sort.Reverse(sort.StringSlice(notTooOld)))
		toRemove := notTooOld[r.retentionFiles:]
		for _, file := range toRemove {
			if err := os.Remove(filepath.Join(dir, file)); err != nil {
				errors <- err
			}
		}
	}
	errors <- nil
}

// Write writes its argument to the file that is indicated by the Path field of
// RotateFileLoggerConfig. If the write would cause the file to grow beyond the
// bounds of its maximum size, then the call blocks until the existing log file
// can be moved to a new location, and the path can be opened again as a new
// file.
func (r *RotateFileLogger) Write(p []byte) (int, error) {
	writer := r.container.Load().(*rotateFile)
	n, err := writer.Write(p)
	if err == os.ErrClosed {
		if atomic.LoadInt64(&r.closed) == 0 {
			// The file was closed for rotation, write to the next one
			return r.Write(p)
		}
	}
	return n, err
}

// Close closes the currently opened log file. Calling Write after Close will
// result in an error.
func (r *RotateFileLogger) Close() error {
	atomic.StoreInt64(&r.closed, 1)
	return r.container.Load().(*rotateFile).Close()
}
