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
	_ = tempfile.Close()
	defer os.Remove(tempfile.Name())
	if err := pipeLogsToFile(tempfile.Name()); err != nil {
		t.Fatal(err)
	}
	want := "HELLO, WORLD"
	fmt.Println(want)

	// Give the copyLines goroutines a chance to run, mea culpa
	time.Sleep(time.Second)

	os.Stdout, os.Stderr = oldStdout, oldStderr
	f, err := os.Open(tempfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(b)); got != want {
		t.Fatalf("bad output: got %q, want %q", got, want)
	}
}
