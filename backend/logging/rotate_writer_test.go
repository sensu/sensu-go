package logging

import (
	"log"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestRotateWriter(t *testing.T) {
	// Create a temporary files that will be used as the event log file
	file, err := os.CreateTemp(os.TempDir(), "event.*.log")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
	rotatedFilename := file.Name() + ".1"
	defer os.Remove(rotatedFilename)

	// Setup our custom writer
	rotate := make(chan interface{}, 1)
	defer close(rotate)
	w, err := NewRotateWriter(file.Name(), rotate)
	if err != nil {
		t.Fatal(err)
	}

	// We should be able to write 3 bytes to our log file
	_, err = w.Write([]byte("foo"))
	assert.NoError(t, err)

	// Rename our file to simulate a log rotation and make sure the original filename does not exist anymore
	err = os.Rename(file.Name(), rotatedFilename)
	assert.NoError(t, err)
	_, err = os.Stat(file.Name())
	assert.Error(t, err)

	// Hook ourselves into logrus to capture the log entry sent by the code below
	l, hook := test.NewNullLogger()
	logger = l.WithField("test", "TestRotateWriter")

	// Send SIGHUP to our test process so the writer re-open the log file
	go func() {
		rotate <- syscall.SIGHUP
	}()

	// Wait the log file to be reopened, which is done by waiting for logrus to
	// receive a log entry
	max := 5 * time.Second
	start := time.Now()
	for {
		entries := hook.AllEntries()
		if len(entries) > 0 {
			break
		}
		if time.Since(start) > max {
			t.Fatal("log file was not reopened in a timely manner")
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Write to our new log file and make sure it's the right size and does
	// contain what was wrote before
	_, err = w.Write([]byte("foobar"))
	assert.NoError(t, err)
	info, _ := os.Stat(file.Name())
	assert.Equal(t, int64(6), info.Size())

	w.Close()
}

func TestSpecialFileSync(t *testing.T) {
	stdoutPath := "/dev/stdout"

	w, err := NewRotateWriter(stdoutPath, nil)
	if err != nil {
		t.Fatal(err)
	}

	if w.isSpecial != true {
		t.Errorf("expected %s to be detected as special", stdoutPath)
	}
	if err := w.Sync(); err != nil {
		t.Error(err)
	}
}

func TestRotateWriterSizeBasedRotation(t *testing.T) {
	dir := t.TempDir()
	logPath := dir + "/metrics.log"

	rotate := make(chan interface{}, 1)
	defer close(rotate)

	w, err := NewRotateWriter(logPath, rotate)
	if err != nil {
		t.Fatal(err)
	}

	// Set a small max size for testing
	w.maxSize = 100

	// Write data to exceed the max size
	data := make([]byte, 60)
	for i := range data {
		data[i] = 'x'
	}

	_, err = w.Write(data)
	assert.NoError(t, err)

	// This write should trigger rotation (60 + 60 = 120 > 100)
	_, err = w.Write(data)
	assert.NoError(t, err)

	// The rotated file should exist
	_, err = os.Stat(logPath + ".1")
	assert.NoError(t, err, "rotated file should exist after exceeding max size")

	// The current log file should be small (only the post-rotation content)
	info, err := os.Stat(logPath)
	assert.NoError(t, err)
	assert.Less(t, info.Size(), int64(100), "current file should be smaller than max size after rotation")

	w.Close()
}

func TestRotateWriterMultipleRotations(t *testing.T) {
	dir := t.TempDir()
	logPath := dir + "/metrics.log"

	rotate := make(chan interface{}, 1)
	defer close(rotate)

	w, err := NewRotateWriter(logPath, rotate)
	if err != nil {
		t.Fatal(err)
	}

	// Set a small max size for testing
	w.maxSize = 50

	data := make([]byte, 60)
	for i := range data {
		data[i] = 'a'
	}

	// Trigger multiple rotations
	for i := 0; i < 3; i++ {
		_, err = w.Write(data)
		assert.NoError(t, err)
	}

	// We should have rotated files .1 and .2
	_, err = os.Stat(logPath + ".1")
	assert.NoError(t, err, "rotated file .1 should exist")
	_, err = os.Stat(logPath + ".2")
	assert.NoError(t, err, "rotated file .2 should exist")

	w.Close()
}

func TestRotateWriterRespectsMaxRotatedFiles(t *testing.T) {
	dir := t.TempDir()
	logPath := dir + "/metrics.log"

	rotate := make(chan interface{}, 1)
	defer close(rotate)

	w, err := NewRotateWriter(logPath, rotate)
	if err != nil {
		t.Fatal(err)
	}

	// Set a small max size for testing
	w.maxSize = 20

	data := make([]byte, 30)
	for i := range data {
		data[i] = 'b'
	}

	// Trigger more rotations than maxRotatedFiles (5)
	for i := 0; i < 8; i++ {
		_, err = w.Write(data)
		assert.NoError(t, err)
	}

	// Files beyond maxRotatedFiles should not exist
	_, err = os.Stat(logPath + ".5")
	assert.NoError(t, err, "rotated file .5 should exist (max)")

	// File .6 should not exist due to the shift overwriting
	// (the oldest gets pushed off)
	_, err = os.Stat(logPath + ".6")
	assert.True(t, os.IsNotExist(err), "rotated file .6 should not exist (exceeds max)")

	w.Close()
}
