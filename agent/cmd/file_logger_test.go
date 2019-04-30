package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPipeLogsToFile(t *testing.T) {
	oldStdout, oldStderr := os.Stdout, os.Stderr
	tempfile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempfile.Name())
	defer tempfile.Close()
	if err := pipeLogsToFile(tempfile); err != nil {
		t.Fatal(err)
	}
	want := "HELLO, WORLD"
	fmt.Println(want)

	// Give the copyLines goroutines a chance to run, mea culpa
	time.Sleep(time.Second)

	os.Stdout, os.Stderr = oldStdout, oldStderr
	if _, err := tempfile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	b, err := ioutil.ReadAll(tempfile)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(b)); got != want {
		t.Fatalf("bad output: got %q, want %q", got, want)
	}
}
