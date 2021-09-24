package logging

import (
	"io/ioutil"
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
	file, err := ioutil.TempFile(os.TempDir(), "event.*.log")
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
