package logging

import (
	"archive/zip"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRotateFileLogger(t *testing.T) {
	td, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(td)
	}()
	logPath := filepath.Join(td, "sensu.log")

	var wg, writeWg sync.WaitGroup
	writeWg.Add(100)
	wg.Add(99)

	config := RotateFileLoggerConfig{
		Path:           logPath,
		MaxSizeBytes:   1024,
		RetentionFiles: 10,
		// sync is to cause the writer to block on zipping the file, which is
		// useful for testing, but not desirable for production use.
		sync: true,
	}

	writer, err := NewRotateFileLogger(config)
	if err != nil {
		t.Fatal(err)
	}

	msg := make([]byte, 512)
	for i := range msg {
		msg[i] = '!'
	}

	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		go func(i int) {
			time.Sleep(time.Duration(i) * time.Millisecond)
			defer writeWg.Done()
			if _, err := writer.Write(msg); err != nil {
				errors <- err
			}
		}(i)
	}

	writeWg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}

	dir, err := os.Open(td)
	if err != nil {
		t.Fatal(err)
	}
	names, err := dir.Readdirnames(0)
	if err != nil {
		t.Fatal(err)
	}
	_ = dir.Close()
	if got, want := len(names), 50; got != want {
		t.Errorf("wrong number of log files: got %d, want %d", got, want)
	}
	for i, name := range names {
		if name == "sensu.log" {
			continue
		}
		if !strings.HasSuffix(name, ".zip") && name != "sensu.log" {
			t.Errorf("bad archive name: got %s but wanted something in zip", name)
		}
		zipfile, err := zip.OpenReader(filepath.Join(td, name))
		if err != nil {
			t.Errorf("file %d: %s", i, err)
			continue
		}
		if got, want := len(zipfile.File), 1; got != want {
			t.Errorf("file %d: bad number of files in zip file: got %d, want %d", i, got, want)
		}
		if got, want := int(zipfile.File[0].UncompressedSize), 1024; got != want {
			t.Errorf("file %d: bad uncompressed size: got %d, want %d", i, got, want)
		}
		_ = zipfile.Close()
	}
	ctx, cancel := context.WithCancel(context.Background())
	ch := writer.StartReaper(ctx, time.Millisecond*100)
	if err := <-ch; err != nil {
		t.Fatal(err)
	}
	cancel()
	dir, err = os.Open(td)
	if err != nil {
		t.Fatal(err)
	}
	names, err = dir.Readdirnames(0)
	if err != nil {
		t.Fatal(err)
	}
	_ = dir.Close()
	if got, want := len(names), 11; got != want {
		// 10 zipped files plus sensu.log
		t.Fatalf("wrong number of log files: got %d, want %d", got, want)
	}
}
